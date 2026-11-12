package gui

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"sourcesight/pkg/applog"
	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
	"sourcesight/pkg/history"
	"sourcesight/pkg/parser"
)

// A batch analyzes several demos at once:
//
//   - workers (one per CPU thread, capped by free memory) do the heavy part in
//     parallel: cache lookup, file hash and parsing;
//   - a single committer stores results one by one, in completion order: duplicate
//     questions, history and reoffender database are never touched concurrently, so
//     two copies of the same match in one batch can't both slip through;
//   - an emitter pushes the state of every row to the page every 200 ms, batching
//     what would otherwise be hundreds of progress updates per second.

// Row states, as shown in the page.
const (
	stQueued   = "queued"
	stHashing  = "hashing"
	stParsing  = "parsing"
	stSaving   = "saving"
	stDone     = "done"
	stCached   = "cached"
	stSkipped  = "skipped"
	stError    = "error"
	stCanceled = "canceled"
)

// errSkipped marks a demo the user chose not to keep (duplicate dialog).
var errSkipped = errors.New("omitida")

type batchRow struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	State     string   `json:"state"`
	Pct       int      `json:"pct"`
	Level     string   `json:"level,omitempty"`
	Flags     []string `json:"flags,omitempty"`
	HistoryID string   `json:"historyId,omitempty"`
	Error     string   `json:"error,omitempty"`
	Seconds   float64  `json:"seconds,omitempty"`
}

type batchInfo struct {
	Total    int     `json:"total"`
	Finished int     `json:"finished"`
	Workers  int     `json:"workers"`
	Elapsed  float64 `json:"elapsed"`
	Eta      float64 `json:"eta"` // seconds, 0 when unknown
	Done     bool    `json:"done"`
	Canceled bool    `json:"canceled"`
}

type batch struct {
	mu       sync.Mutex
	rows     []batchRow
	dirty    bool
	started  time.Time
	workers  int
	cancel   chan struct{}
	canceled atomic.Bool
}

func (b *batch) update(i int, f func(r *batchRow)) {
	b.mu.Lock()
	f(&b.rows[i])
	b.dirty = true
	b.mu.Unlock()
}

// snapshot copies the rows and computes totals; ok is false when nothing changed.
func (b *batch) snapshot(force bool) ([]batchRow, batchInfo, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.dirty && !force {
		return nil, batchInfo{}, false
	}
	b.dirty = false
	rows := append([]batchRow(nil), b.rows...)
	info := batchInfo{Total: len(rows), Workers: b.workers, Elapsed: time.Since(b.started).Seconds(), Canceled: b.canceled.Load()}
	progress := 0.0
	for _, r := range rows {
		switch r.State {
		case stDone, stCached, stSkipped, stError, stCanceled:
			info.Finished++
			progress++
		default:
			progress += float64(r.Pct) / 100
		}
	}
	if progress > 0.05 && info.Finished < info.Total {
		info.Eta = info.Elapsed * (float64(info.Total) - progress) / progress
	}
	return rows, info, true
}

func (b *batch) stop() {
	if b.canceled.CompareAndSwap(false, true) {
		close(b.cancel)
	}
}

// parsed is what a worker hands to the committer.
type parsed struct {
	i      int
	path   string
	rep    *core.DemoReport
	entry  *history.Entry
	sha    string
	cached bool
	err    error
	took   time.Duration
}

var (
	busy         atomic.Bool
	currentBatch atomic.Pointer[batch]
)

// processQueue analyzes demos in parallel. A single demo opens its report when it
// finishes; several show a summary with every result.
func (a *app) processQueue(paths []string, force bool) {
	applog.Printf("cola: entrada %q", paths)
	paths = expandDemoPaths(paths)
	if len(paths) == 0 {
		a.call("restoreIdle")
		a.toast("No se encontraron archivos .dem")
		return
	}
	if !busy.CompareAndSwap(false, true) {
		a.toast("Ya hay un análisis en curso")
		return
	}
	defer busy.Store(false)

	b := &batch{started: time.Now(), cancel: make(chan struct{}), workers: workerCount(len(paths))}
	for i, p := range paths {
		b.rows = append(b.rows, batchRow{ID: i, Name: filepath.Base(p), State: stQueued})
	}
	currentBatch.Store(b)
	defer currentBatch.Store(nil)
	applog.Printf("cola: %d demos, %d en paralelo", len(paths), b.workers)

	single := len(paths) == 1
	if single {
		a.call("showLoading")
	} else {
		rows, info, _ := b.snapshot(true)
		a.call("batchStart", rows, info)
	}

	// Emitter. emitStopped lets the final result wait for it: an update sent after
	// the summary would put the page back in "analyzing".
	emitDone, emitStopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(emitStopped)
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-emitDone:
				return
			case <-t.C:
				rows, info, ok := b.snapshot(false)
				if !ok {
					continue
				}
				if single {
					r := rows[0]
					a.call("setProgress", r.Pct, stateLabel(r)+" "+r.Name)
				} else {
					a.call("batchUpdate", rows, info)
				}
			}
		}
	}()

	// Workers
	jobs := make(chan int)
	results := make(chan parsed)
	var wg sync.WaitGroup
	for w := 0; w < b.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results <- a.parseOne(b, i, paths[i], force)
			}
		}()
	}
	go func() {
		defer close(jobs)
		for i := range paths {
			select {
			case jobs <- i:
			case <-b.cancel:
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	// Committer: one result at a time
	var lastReport *core.DemoReport
	var lastMeta reportMeta
	for res := range results {
		rep, meta, err := a.safeCommit(b, res, force)
		if err == nil && rep != nil {
			lastReport, lastMeta = rep, meta
			a.call("refreshHistory")
		}
	}
	// Demos never handed to a worker because of a cancel
	b.mu.Lock()
	for i := range b.rows {
		if b.rows[i].State == stQueued {
			b.rows[i].State = stCanceled
		}
	}
	b.dirty = true
	b.mu.Unlock()
	close(emitDone)
	<-emitStopped

	rows, info, _ := b.snapshot(true)
	info.Done = true
	applog.Printf("cola: terminada en %.1fs (%d en paralelo)", info.Elapsed, b.workers)
	if !single {
		a.call("batchDone", rows, info)
		return
	}
	switch r := rows[0]; r.State {
	case stDone, stCached:
		a.show(lastReport, lastMeta)
		if r.State == stCached {
			a.toast("Abierta desde el historial (ya estaba analizada)")
		}
	case stSkipped:
		a.call("restoreIdle")
		a.toast("No se agregó: la demo estaba repetida")
	case stCanceled:
		a.call("restoreIdle")
		a.toast("Análisis cancelado")
	default:
		a.call("restoreIdle")
		a.toast("No se pudo analizar " + r.Name + ": " + r.Error)
	}
}

func stateLabel(r batchRow) string {
	switch r.State {
	case stHashing:
		return "Verificando"
	case stSaving:
		return "Guardando"
	default:
		return fmt.Sprintf("Analizando (%d%%)", r.Pct)
	}
}

// parseOne is the parallel part: cached report, file hash and parsing.
func (a *app) parseOne(b *batch, i int, path string, force bool) (res parsed) {
	res = parsed{i: i, path: path}
	start := time.Now()
	defer func() {
		// A parser panic must fail this demo, not the whole app
		if r := recover(); r != nil {
			res.err = fmt.Errorf("error interno: %v", r)
		}
		res.took = time.Since(start)
	}()

	if b.canceled.Load() {
		res.err = parser.ErrCanceled
		return res
	}
	if !force {
		if rep, entry := history.Lookup(path); rep != nil {
			res.rep, res.entry, res.cached = rep, entry, true
			return res
		}
	}
	b.update(i, func(r *batchRow) { r.State = stHashing })
	res.sha, _ = history.FileSHA256(path)

	b.update(i, func(r *batchRow) { r.State = stParsing })
	res.rep, res.err = parser.ParseDemo(path, parser.ParserOptions{
		Cancel: b.cancel,
		OnProgress: func(f float64) {
			b.update(i, func(r *batchRow) { r.Pct = min(99, int(f*100)) })
		},
	})
	return res
}

// safeCommit runs commitOne so that an unexpected failure while storing one demo
// marks that demo as failed instead of closing the app.
func (a *app) safeCommit(b *batch, res parsed, force bool) (rep *core.DemoReport, meta reportMeta, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error interno guardando la demo: %v", r)
			applog.Printf("cola: %s: %v", filepath.Base(res.path), err)
			b.update(res.i, func(row *batchRow) { row.State, row.Pct, row.Error = stError, 100, err.Error() })
			rep = nil
		}
	}()
	return a.commitOne(b, res, force)
}

// commitOne is the serial part: duplicate questions, history and reoffender database.
func (a *app) commitOne(b *batch, res parsed, force bool) (*core.DemoReport, reportMeta, error) {
	meta := reportMeta{DemoPath: res.path, FromCache: res.cached, QueueIndex: res.i + 1, QueueTotal: len(b.rows)}
	finish := func(state string, rep *core.DemoReport, entry *history.Entry, err error) {
		b.update(res.i, func(r *batchRow) {
			r.State, r.Pct, r.Seconds = state, 100, res.took.Seconds()
			if err != nil && state == stError {
				r.Error = err.Error()
			}
			if rep != nil {
				r.Level, r.Flags = reportLevel(rep)
			}
			if entry != nil {
				r.HistoryID = entry.ID
			}
		})
	}

	switch {
	case errors.Is(res.err, parser.ErrCanceled):
		finish(stCanceled, nil, nil, nil)
		return nil, meta, res.err
	case res.err != nil:
		applog.Printf("cola: %s falló: %v", filepath.Base(res.path), res.err)
		finish(stError, nil, nil, res.err)
		return nil, meta, res.err
	case res.cached:
		meta.HistoryID = res.entry.ID
		finish(stCached, res.rep, res.entry, nil)
		return res.rep, meta, nil
	}

	b.update(res.i, func(r *batchRow) { r.State = stSaving })
	rep, entry, cached, err := a.store(res, force)
	switch {
	case errors.Is(err, errSkipped):
		finish(stSkipped, nil, nil, nil)
		return nil, meta, err
	case err != nil:
		finish(stError, nil, nil, err)
		return nil, meta, err
	}
	meta.FromCache = cached
	if entry != nil {
		meta.HistoryID = entry.ID
	}
	state := stDone
	if cached {
		state = stCached
	}
	finish(state, rep, entry, nil)
	return rep, meta, nil
}

// reportLevel returns the worst player level and the distinct detections of a report.
func reportLevel(rep *core.DemoReport) (string, []string) {
	level := core.LevelClean
	seen := map[string]bool{}
	var flags []string
	for _, p := range rep.Players {
		level = core.MaxLevel(level, p.Level)
		for _, f := range p.Flags {
			if !seen[f] {
				seen[f] = true
				flags = append(flags, f)
			}
		}
	}
	return level, flags
}

func shortDate(ts string) string {
	if len(ts) < 10 {
		return ts
	}
	return ts[8:10] + "/" + ts[5:7] + "/" + ts[0:4]
}

// store saves a freshly parsed demo, asking first when it duplicates a stored one.
// cached is true when the user chose to open the stored copy instead.
func (a *app) store(res parsed, force bool) (*core.DemoReport, *history.Entry, bool, error) {
	name := filepath.Base(res.path)
	ownID, _ := history.ID(res.path)
	rep := res.rep

	// Byte-identical copy stored under another name or folder
	if !force {
		if dup := history.FindByFileHash(res.sha); dup != nil && dup.ID != ownID {
			choice := a.ask("Demo repetida",
				fmt.Sprintf("«%s» es el mismo archivo que «%s», que analizaste el %s.", name, dup.DemoName, shortDate(dup.AnalyzedAt)),
				[]promptOption{{"open", "Abrir la guardada", true}, {"analyze", "Guardar igual", false}, {"skip", "Saltear", false}})
			switch choice {
			case "open":
				saved, entry, err := history.Open(dup.ID)
				return saved, entry, true, err
			case "analyze":
			default:
				return nil, nil, false, errSkipped
			}
		}
	}

	pdb := db.GetDatabase()
	if force {
		// Re-analysis: forget what this file counted before, then record the new result
		if _, old := history.Lookup(res.path); old != nil {
			pdb.RemoveMatch(old.MatchID)
		}
		pdb.ForgetLegacyDemo(name)
	}

	// Same match by content: a renamed copy, a second recording or a trimmed file
	// (skipped on re-analysis: the user already decided about duplicates when adding the demo)
	if similar := history.FindSimilar(rep, ownID); !force && len(similar) > 0 {
		old := similar[0]
		choice := a.ask("Partida repetida",
			fmt.Sprintf("«%s» parece la misma partida que «%s» (%s, %d rondas, mismos jugadores), analizada el %s. Si guardás las dos, la partida cuenta doble en los promedios del historial.",
				name, old.DemoName, old.MapName, old.Rounds, shortDate(old.AnalyzedAt)),
			[]promptOption{{"replace", "Reemplazar la anterior", true}, {"keep", "Guardar las dos", false}, {"discard", "Descartar esta", false}})
		switch choice {
		case "replace":
			for _, s := range similar {
				if mid, _ := history.Remove(s.ID); mid != "" && !history.MatchInUse(mid) {
					pdb.RemoveMatch(mid)
				}
			}
		case "keep":
		default:
			return nil, nil, false, errSkipped
		}
	}

	entry, err := history.Save(res.path, rep, res.sha)
	if err != nil {
		a.toast("No se pudo guardar en el historial: " + err.Error())
	}
	pdb.RecordReport(rep)
	return rep, entry, false, nil
}
