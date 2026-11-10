package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/pkg/edge"
	"sourcesight/pkg/appdata"
	"sourcesight/pkg/applog"
	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
	"sourcesight/pkg/evidence"
	"sourcesight/pkg/history"
	"sourcesight/pkg/launcher"
	"sourcesight/pkg/parser"
	"sourcesight/pkg/report"
	"sourcesight/pkg/steamid"
	"sourcesight/pkg/update"
	"sourcesight/pkg/version"
)

var (
	appLock         sync.Mutex
	currentReport   *core.DemoReport
	currentDemoPath string
	pendingUpdate   *update.Release
	busy            atomic.Bool
)

// app wraps the webview so Go can call page functions with JSON-encoded arguments.
type app struct {
	w webview2.WebView
}

// call runs fn(args...) in the page on the UI thread. Every argument is JSON-encoded,
// so strings never need manual escaping.
func (a *app) call(fn string, args ...any) {
	parts := make([]string, len(args))
	for i, arg := range args {
		b, err := json.Marshal(arg)
		if err != nil {
			b = []byte("null")
		}
		parts[i] = string(b)
	}
	script := fn + "(" + strings.Join(parts, ",") + ");"
	a.w.Dispatch(func() { a.w.Eval(script) })
}

func (a *app) toast(msg string) { a.call("showToast", msg) }

func openExternal(target string) {
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
}

// expandDemoPaths turns dropped/selected paths into .dem files; folders contribute
// the demos directly inside them.
func expandDemoPaths(paths []string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		key := strings.ToLower(p)
		if !seen[key] && strings.EqualFold(filepath.Ext(p), ".dem") {
			seen[key] = true
			out = append(out, p)
		}
	}
	for _, p := range paths {
		p = strings.Trim(p, `"' `)
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if st.IsDir() {
			matches, _ := filepath.Glob(filepath.Join(p, "*.dem"))
			for _, m := range matches {
				add(m)
			}
			continue
		}
		add(p)
	}
	return out
}

// locateDemoFile finds a dropped demo by name when the WebView2 runtime is too old
// to report the real path. It searches the usual folders a few levels deep.
func locateDemoFile(inputPath string) string {
	clean := strings.Trim(inputPath, `"'`)
	if clean == "" {
		return ""
	}
	if _, err := os.Stat(clean); err == nil {
		return clean
	}

	base := filepath.Base(clean)
	if base == "" || base == "." {
		return ""
	}

	cwd, _ := os.Getwd()
	userProfile := os.Getenv("USERPROFILE")
	roots := []string{
		cwd,
		filepath.Dir(cwd),
		filepath.Join(userProfile, "Desktop"),
		filepath.Join(userProfile, "Downloads"),
		filepath.Join(userProfile, "Documents"),
	}
	if cfg := launcher.LoadConfig(); cfg.CSGOPath != "" {
		roots = append(roots, cfg.CSGOPath)
	}

	for _, root := range roots {
		if root == "" {
			continue
		}
		var found string
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(root, path)
			if strings.Count(rel, string(filepath.Separator)) > 3 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.IsDir() && strings.EqualFold(d.Name(), base) {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	return ""
}

type reportMeta struct {
	HistoryID  string `json:"historyId"`
	DemoPath   string `json:"demoPath"`
	FromCache  bool   `json:"fromCache"`
	QueueIndex int    `json:"queueIndex"`
	QueueTotal int    `json:"queueTotal"`
}

// errSkipped marks a demo the user chose not to keep (duplicate dialog).
var errSkipped = errors.New("omitida")

type promptOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Primary bool   `json:"primary,omitempty"`
}

var (
	promptSeq     atomic.Int64
	promptWaiters sync.Map // id -> chan string
)

// ask shows a modal question in the page and blocks until the user picks an option.
func (a *app) ask(title, text string, options []promptOption) string {
	id := fmt.Sprint(promptSeq.Add(1))
	ch := make(chan string, 1)
	promptWaiters.Store(id, ch)
	defer promptWaiters.Delete(id)
	a.call("showPrompt", id, title, text, options)
	select {
	case choice := <-ch:
		return choice
	case <-time.After(30 * time.Minute):
		return ""
	}
}

func shortDate(ts string) string {
	if len(ts) < 10 {
		return ts
	}
	return ts[8:10] + "/" + ts[5:7] + "/" + ts[0:4]
}

// analyze returns the report for one demo. Unless force is set it reuses the cached
// report of the same file, and asks before storing a duplicate of a stored match.
func (a *app) analyze(path string, force bool, label string) (*core.DemoReport, *history.Entry, bool, error) {
	if !force {
		if rep, entry := history.Lookup(path); rep != nil {
			return rep, entry, true, nil
		}
	}
	name := filepath.Base(path)
	ownID, _ := history.ID(path)

	// Byte-identical copy stored under another name or folder
	a.call("setProgress", 0, label+" — verificando")
	sha, _ := history.FileSHA256(path)
	if !force {
		if dup := history.FindByFileHash(sha); dup != nil && dup.ID != ownID {
			choice := a.ask("Demo repetida",
				fmt.Sprintf("«%s» es el mismo archivo que «%s», que analizaste el %s.", name, dup.DemoName, shortDate(dup.AnalyzedAt)),
				[]promptOption{{"open", "Abrir la guardada", true}, {"analyze", "Analizar igual", false}, {"skip", "Saltear", false}})
			switch choice {
			case "open":
				rep, entry, err := history.Open(dup.ID)
				return rep, entry, true, err
			case "analyze":
			default:
				return nil, nil, false, errSkipped
			}
		}
	}

	rep, err := parser.ParseDemo(path, parser.ParserOptions{
		OnProgress: func(fraction float64) {
			pct := int(fraction * 100)
			if pct > 99 {
				pct = 99
			}
			a.call("setProgress", pct, fmt.Sprintf("%s — %d%%", label, pct))
		},
	})
	if err != nil {
		return nil, nil, false, err
	}

	pdb := db.GetDatabase()
	if force {
		// Re-analysis: forget what this file counted before, then record the new result
		if _, old := history.Lookup(path); old != nil {
			pdb.RemoveMatch(old.MatchID)
		}
		pdb.ForgetLegacyDemo(filepath.Base(path))
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

	entry, err := history.Save(path, rep, sha)
	if err != nil {
		a.toast("No se pudo guardar en el historial: " + err.Error())
	}
	pdb.RecordReport(rep)
	return rep, entry, false, nil
}

// processQueue analyzes demos one after another and finally shows the most relevant
// report: the first one with infractions, otherwise the last one.
func (a *app) processQueue(paths []string, force bool) {
	applog.Printf("cola: entrada %q", paths)
	paths = expandDemoPaths(paths)
	applog.Printf("cola: demos %q", paths)
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

	defer func() {
		// A parser panic would otherwise kill the whole app without a word
		if r := recover(); r != nil {
			a.call("restoreIdle")
			a.toast(fmt.Sprintf("Error interno analizando la demo: %v", r))
		}
	}()

	a.call("showLoading")

	var shown *core.DemoReport
	var shownMeta reportMeta
	var failures []string
	skipped := 0
	for i, path := range paths {
		label := "Analizando " + filepath.Base(path)
		if len(paths) > 1 {
			label = fmt.Sprintf("Demo %d de %d · %s", i+1, len(paths), filepath.Base(path))
		}
		a.call("setProgress", 0, label)

		rep, entry, cached, err := a.analyze(path, force, label)
		if errors.Is(err, errSkipped) {
			skipped++
			continue
		}
		if err != nil {
			failures = append(failures, filepath.Base(path)+": "+err.Error())
			continue
		}
		meta := reportMeta{DemoPath: path, FromCache: cached, QueueIndex: i + 1, QueueTotal: len(paths)}
		if entry != nil {
			meta.HistoryID = entry.ID
		}
		a.call("refreshHistory")

		hasFlags := false
		for _, p := range rep.Players {
			if len(p.Flags) > 0 {
				hasFlags = true
				break
			}
		}
		if shown == nil || (hasFlags && !reportHasFlags(shown)) || (!reportHasFlags(shown) && !hasFlags) {
			shown, shownMeta = rep, meta
		}
	}

	if shown == nil {
		a.call("restoreIdle")
		if len(failures) > 0 {
			a.toast("No se pudo analizar: " + strings.Join(failures, " · "))
		} else if skipped > 0 {
			a.toast("No se agregó ninguna demo: estaban repetidas")
		}
		return
	}
	a.show(shown, shownMeta)

	switch {
	case len(failures) > 0:
		a.toast(fmt.Sprintf("%d demo(s) fallaron: %s", len(failures), strings.Join(failures, " · ")))
	case len(paths) > 1 && skipped > 0:
		a.toast(fmt.Sprintf("%d demos analizadas, %d salteadas por repetidas", len(paths)-skipped, skipped))
	case len(paths) > 1:
		a.toast(fmt.Sprintf("%d demos analizadas — están todas en el historial", len(paths)))
	case shownMeta.FromCache:
		a.toast("Abierta desde el historial (ya estaba analizada)")
	}
}

func reportHasFlags(rep *core.DemoReport) bool {
	for _, p := range rep.Players {
		if len(p.Flags) > 0 {
			return true
		}
	}
	return false
}

func (a *app) show(rep *core.DemoReport, meta reportMeta) {
	appLock.Lock()
	currentReport = rep
	currentDemoPath = meta.DemoPath
	appLock.Unlock()
	a.call("loadDemoReport", rep, meta)
}

func current() (*core.DemoReport, string) {
	appLock.Lock()
	defer appLock.Unlock()
	return currentReport, currentDemoPath
}

func findPlayer(rep *core.DemoReport, steamID string) (core.PlayerStats, bool) {
	id, _ := strconv.ParseUint(steamID, 10, 64)
	for _, p := range rep.Players {
		if p.SteamID == id {
			return p, true
		}
	}
	return core.PlayerStats{}, false
}

// LaunchApp initializes the native desktop GUI and runs the event loop.
func LaunchApp(initialDemoPath string) error {
	update.CleanupOld()
	applog.Printf("inicio SourceSight %s", version.Version)

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "SourceSight",
			Width:  1320,
			Height: 860,
			IconId: 1, // RT_GROUP_ICON #1 in release builds; setWindowIcon covers local ones
			Center: true,
		},
	})
	if w == nil {
		return fmt.Errorf("no se pudo inicializar WebView2. Instalá Microsoft Edge WebView2 Runtime")
	}
	defer w.Destroy()
	w.SetSize(1320, 860, webview2.HintNone)
	setWindowIcon(uintptr(w.Window()))

	a := &app{w: w}

	// Real paths of files dropped onto the window (patched go-webview2, see third_party/)
	edge.DebugLog = applog.Printf
	edge.OnFilesDropped = func(paths []string) {
		go a.processQueue(paths, false)
	}

	// Fallback for WebView2 runtimes older than 1.0.1774: only names arrive, so search for them
	_ = w.Bind("onDemoFilesDroppedByName", func(names []string) error {
		var paths, missing []string
		for _, n := range names {
			if p := locateDemoFile(n); p != "" {
				paths = append(paths, p)
			} else {
				missing = append(missing, filepath.Base(n))
			}
		}
		if len(missing) > 0 {
			a.toast("No encontré " + strings.Join(missing, ", ") + ". Usá el botón Abrir demos.")
		}
		if len(paths) == 0 {
			a.call("restoreIdle")
			return nil
		}
		go a.processQueue(paths, false)
		return nil
	})

	_ = w.Bind("selectDemoFile", func() error {
		paths, err := OpenDemoDialog()
		if err != nil || len(paths) == 0 {
			a.call("restoreIdle")
			return nil
		}
		go a.processQueue(paths, false)
		return nil
	})

	// reanalyzeOutdated re-runs every stored demo analyzed by older detectors
	_ = w.Bind("reanalyzeOutdated", func() error {
		var paths []string
		missing := 0
		for _, e := range history.Outdated() {
			if _, err := os.Stat(e.DemoPath); err != nil {
				missing++
				continue
			}
			paths = append(paths, e.DemoPath)
		}
		if missing > 0 {
			a.toast(fmt.Sprintf("%d demo(s) ya no están en su carpeta y no se pueden re-analizar", missing))
		}
		if len(paths) > 0 {
			go a.processQueue(paths, true)
		}
		return nil
	})

	_ = w.Bind("reanalyzeCurrent", func() error {
		_, path := current()
		if path == "" {
			return nil
		}
		if _, err := os.Stat(path); err != nil {
			a.toast("La demo ya no está en " + path)
			return nil
		}
		go a.processQueue([]string{path}, true)
		return nil
	})

	_ = w.Bind("listHistory", func() []history.Entry {
		return history.List()
	})

	_ = w.Bind("openHistory", func(id string) error {
		rep, entry, err := history.Open(id)
		if err != nil {
			a.toast("No se pudo abrir: " + err.Error())
			return nil
		}
		a.show(rep, reportMeta{HistoryID: entry.ID, DemoPath: entry.DemoPath, FromCache: true, QueueIndex: 1, QueueTotal: 1})
		return nil
	})

	_ = w.Bind("removeHistory", func(id string) error {
		mid, err := history.Remove(id)
		// Its players stop counting as seen in that match, unless a duplicate still holds it
		if mid != "" && !history.MatchInUse(mid) {
			db.GetDatabase().RemoveMatch(mid)
		}
		return err
	})

	_ = w.Bind("answerPrompt", func(id, choice string) error {
		if ch, ok := promptWaiters.Load(id); ok {
			select {
			case ch.(chan string) <- choice:
			default:
			}
		}
		return nil
	})

	_ = w.Bind("getOverview", func() history.Overview {
		return history.GetOverview()
	})

	// listPlayers searches by name, or by SteamID in any common format
	_ = w.Bind("listPlayers", func(query string) map[string]any {
		if steamid.IsVanityURL(query) {
			return map[string]any{"rows": []history.PlayerRow{}, "hint": "Los links con nombre (/id/...) no se pueden buscar sin conexión a Steam. Usá el link /profiles/ o el SteamID64."}
		}
		if id, ok := steamid.Parse(query); ok {
			return map[string]any{"rows": history.ListPlayers("", strconv.FormatUint(id, 10)), "steamId": strconv.FormatUint(id, 10)}
		}
		return map[string]any{"rows": history.ListPlayers(query, "")}
	})

	_ = w.Bind("getPlayerDetail", func(steamID string) *history.PlayerDetail {
		return history.GetPlayer(steamID)
	})

	_ = w.Bind("getFlagLevels", func() map[string]string {
		out := map[string]string{}
		for _, f := range core.KnownFlags() {
			out[f] = core.FlagLevel(f)
		}
		return out
	})

	_ = w.Bind("listConfirmed", func() map[string]db.Confirmation {
		return db.GetDatabase().Confirmations()
	})

	// setConfirmed is the staff decision after watching the demo; the app never sets it
	_ = w.Bind("setConfirmed", func(steamID, name, note string, confirmed bool) error {
		id, err := strconv.ParseUint(steamID, 10, 64)
		if err != nil {
			return err
		}
		db.GetDatabase().SetConfirmed(id, name, note, confirmed)
		applog.Printf("staff: %s (%s) confirmado=%v", name, steamID, confirmed)
		return nil
	})

	_ = w.Bind("exportPlayerEvidence", func(steamID string) error {
		rep, path := current()
		if rep == nil {
			return nil
		}
		p, ok := findPlayer(rep, steamID)
		if !ok {
			a.toast("Jugador no encontrado en este reporte")
			return nil
		}
		out, err := evidence.Write(appdata.ExportDir(), rep, p, path)
		if err != nil {
			a.toast("Error exportando la evidencia: " + err.Error())
			return nil
		}
		openExternal(out)
		a.toast("Evidencia guardada en Documentos\\SourceSight\\" + filepath.Base(out))
		return nil
	})

	_ = w.Bind("exportPNGReport", func() error {
		rep, path := current()
		if rep == nil {
			return nil
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if name == "" || name == "." {
			name = "demo"
		}
		out := filepath.Join(appdata.ExportDir(), fmt.Sprintf("reporte_%s_%s.png", name, time.Now().Format("20060102_150405")))
		if err := report.RenderReportPNG(rep, out); err != nil {
			a.toast("Error generando el PNG: " + err.Error())
			return nil
		}
		openExternal(out)
		a.toast("PNG guardado en Documentos\\SourceSight\\" + filepath.Base(out))
		return nil
	})

	_ = w.Bind("copyToClipboard", func(text string) error {
		return SetWindowsClipboard(text)
	})

	_ = w.Bind("openURL", func(target string) error {
		if strings.HasPrefix(target, "https://") {
			openExternal(target)
		}
		return nil
	})

	_ = w.Bind("openExportFolder", func() error {
		openExternal(appdata.ExportDir())
		return nil
	})

	_ = w.Bind("getAppInfo", func() map[string]string {
		return map[string]string{
			"version":         version.Version,
			"analysisVersion": strconv.Itoa(core.AnalysisVersion),
			"repoUrl":         version.RepoURL,
			"dataDir":         appdata.Dir(),
			"exportDir":       appdata.ExportDir(),
		}
	})

	_ = w.Bind("getConfig", func() (string, error) {
		cfg := launcher.LoadConfig()
		if cfg.CSGOPath == "" {
			if autoDir, err := launcher.FindCSGODirectory(); err == nil {
				cfg.CSGOPath = autoDir
			}
		}
		b, err := json.Marshal(cfg)
		return string(b), err
	})

	_ = w.Bind("saveConfig", func(path string, preroll float64) error {
		cfg := launcher.LoadConfig() // keep fields this dialog doesn't edit
		cfg.CSGOPath = launcher.CleanCSGODir(path)
		cfg.PrerollSeconds = preroll
		return launcher.SaveConfig(cfg)
	})

	_ = w.Bind("saveSidebarWidth", func(width int) error {
		cfg := launcher.LoadConfig()
		cfg.SidebarWidth = width
		return launcher.SaveConfig(cfg)
	})

	_ = w.Bind("browseCSGOPath", func() (string, error) {
		path, err := OpenCSGODialog()
		if err != nil || path == "" {
			return "", nil
		}
		return launcher.CleanCSGODir(path), nil
	})

	_ = w.Bind("launchInGame", func(tick int) error {
		_, dPath := current()
		if dPath == "" {
			return nil
		}
		if err := launcher.LaunchInGame(dPath, tick); err != nil {
			a.toast("Error al abrir CS:GO: " + err.Error())
			return nil
		}
		a.toast("Abriendo CS:GO en el tick del incidente...")
		return nil
	})

	_ = w.Bind("checkUpdate", func() map[string]any {
		rel, err := update.Check()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		if rel == nil {
			return map[string]any{"available": false, "current": version.Version}
		}
		appLock.Lock()
		pendingUpdate = rel
		appLock.Unlock()
		return map[string]any{
			"available": true,
			"current":   version.Version,
			"version":   rel.Version,
			"notes":     rel.Notes,
			"pageUrl":   rel.PageURL,
			"canApply":  rel.CanApply(),
		}
	})

	_ = w.Bind("applyUpdate", func() error {
		appLock.Lock()
		rel := pendingUpdate
		appLock.Unlock()
		if rel == nil {
			return nil
		}
		go func() {
			a.toast("Descargando " + rel.Version + "...")
			if err := rel.Apply(); err != nil {
				a.toast("No se pudo actualizar: " + err.Error())
				return
			}
			os.Exit(0) // the new version was already launched
		}()
		return nil
	})

	_ = w.Bind("logClient", func(msg string) error {
		applog.Printf("ui: %s", msg)
		return nil
	})

	_ = w.Bind("onAppReady", func() error {
		if os.Getenv("SOURCESIGHT_SELFTEST") != "" {
			go selfTest(a)
		}
		if initialDemoPath != "" {
			go a.processQueue([]string{initialDemoPath}, false)
		}
		return nil
	})

	w.SetHtml(AppHTML)
	w.Run()

	return nil
}

// selfTest exercises the drop pipeline without a mouse: the page posts a synthetic
// File through postMessageWithAdditionalObjects, then the app closes. Results land in
// the log. Run with SOURCESIGHT_SELFTEST=1.
func selfTest(a *app) {
	time.Sleep(1500 * time.Millisecond)
	applog.Printf("selftest: inicio")
	a.w.Dispatch(func() {
		a.w.Eval(`(function(){
			var wv = window.chrome && window.chrome.webview;
			window.logClient('selftest: chrome.webview=' + !!wv + ' postMessageWithAdditionalObjects=' + !!(wv && wv.postMessageWithAdditionalObjects));
			if (wv && wv.postMessageWithAdditionalObjects) {
				try {
					var f = new File(['x'], 'selftest.dem');
					wv.postMessageWithAdditionalObjects(FILE_DROP_MESSAGE, [f]);
					window.logClient('selftest: postMessageWithAdditionalObjects enviado');
				} catch (err) {
					window.logClient('selftest: error ' + err);
				}
			}
		})();`)
	})
	time.Sleep(4 * time.Second)
	applog.Printf("selftest: fin")
	a.w.Dispatch(func() { a.w.Terminate() })
}
