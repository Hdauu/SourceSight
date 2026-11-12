package gui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sourcesight/pkg/applog"
	"sourcesight/pkg/archive"
	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
	"sourcesight/pkg/history"
	"sourcesight/pkg/launcher"
	"sourcesight/pkg/parser"
)

// Automatic archive (pkg/archive): while the app is open and the option is on, it
// checks the demo page every interval, analyzes new demos into the history and keeps
// the flagged ones in the chosen folder. The bot can run the same thing; a lock file
// keeps them from doing it at the same time.

type archiveStatus struct {
	Running  bool   `json:"running"`
	LastRun  string `json:"lastRun,omitempty"`
	LastNote string `json:"lastNote,omitempty"`
	Kept     int    `json:"kept"`     // this session
	Analyzed int    `json:"analyzed"` // this session
}

var (
	archMu      sync.Mutex
	archState   archiveStatus
	archWakeup  = make(chan struct{}, 1)
	archiveOnce sync.Once // the page can call onAppReady again after a reload
	hotkeysOnce sync.Once
)

func archStatus() archiveStatus {
	archMu.Lock()
	defer archMu.Unlock()
	return archState
}

// archiveLoop runs for the life of the app.
func (a *app) archiveLoop() {
	var last time.Time
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		cfg := launcher.LoadConfig().Archive
		now := false
		select {
		case <-archWakeup:
			now = true
		case <-tick.C:
		}
		if !cfg.Enabled || (!now && !last.IsZero() && time.Since(last) < cfg.Interval()) {
			if !cfg.Enabled && now {
				a.toast("Activá el archivo automático en Configuración")
			}
			continue
		}
		last = time.Now()
		a.runArchive(cfg)
	}
}

func (a *app) runArchive(cfg archive.Config) {
	archMu.Lock()
	archState.Running = true
	archMu.Unlock()
	applog.Printf("archivo: revisando %s", cfg.URL)

	shiftRadius := launcher.LoadConfig().ShiftRadius
	analyze := func(path string) (*core.DemoReport, error) {
		rep, err := parser.ParseDemo(path, parser.ParserOptions{ShiftRadius: shiftRadius})
		if err != nil {
			return nil, err
		}
		storeQuietly(path, rep)
		a.call("refreshHistory")
		return rep, nil
	}
	err := archive.Run(context.Background(), cfg, analyze, func(r archive.Result) {
		archMu.Lock()
		defer archMu.Unlock()
		if r.Err != nil {
			applog.Printf("archivo: %s: %v", r.Demo.File, r.Err)
			return
		}
		archState.Analyzed++
		if r.Kept {
			archState.Kept++
			applog.Printf("archivo: guardada %s (%s %s)", r.Demo.File, r.Level, strings.Join(r.Flags, ", "))
			a.toast(fmt.Sprintf("Archivo: guardé %s (%s: %s)", r.Demo.File, levelName(r.Level), strings.Join(r.Flags, ", ")))
		}
	})

	archMu.Lock()
	archState.Running = false
	archState.LastRun = time.Now().Format("2006-01-02 15:04")
	switch {
	case errors.Is(err, archive.ErrBusy):
		archState.LastNote = "Otra copia (el bot o la app) estaba revisando en ese momento."
	case err != nil:
		archState.LastNote = "Error: " + err.Error()
		applog.Printf("archivo: %v", err)
	default:
		archState.LastNote = fmt.Sprintf("Revisado. En esta sesión: %d analizadas, %d guardadas.", archState.Analyzed, archState.Kept)
	}
	archMu.Unlock()
}

// storeQuietly saves an archived demo to the history without the duplicate
// questions of the manual flow: a match already stored isn't stored twice.
func storeQuietly(path string, rep *core.DemoReport) {
	ownID, _ := history.ID(path)
	if len(history.FindSimilar(rep, ownID)) > 0 {
		return
	}
	sha, _ := history.FileSHA256(path)
	if _, err := history.Save(path, rep, sha); err != nil {
		applog.Printf("archivo: no se pudo guardar %s en el historial: %v", filepath.Base(path), err)
	}
	db.GetDatabase().RecordReport(rep)
}

func levelName(l string) string {
	switch l {
	case core.LevelCheat:
		return "cheat"
	case core.LevelSuspect:
		return "sospechosa"
	}
	return l
}
