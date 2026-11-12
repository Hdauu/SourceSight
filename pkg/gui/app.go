package gui

import (
	"encoding/json"
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
	"sourcesight/pkg/radar"
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
)

// app wraps the webview so Go can call page functions with JSON-encoded arguments.
// pageHost is what the app needs from the window: run code in the page on the UI
// thread. webview2.WebView satisfies it; tests use a fake that records the calls.
type pageHost interface {
	Dispatch(f func())
	Eval(js string)
}

type app struct {
	w    webview2.WebView // nil in tests
	host pageHost
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
	a.host.Dispatch(func() { a.host.Eval(script) })
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

	a := &app{w: w, host: w}

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
		paths, err := OpenDemoDialog(uintptr(w.Window()))
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

	_ = w.Bind("cancelAnalysis", func() error {
		if b := currentBatch.Load(); b != nil {
			b.stop()
			applog.Printf("cola: cancelada por el usuario")
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

	// saveConfig returns the game folder actually stored (csgo.exe paths become the
	// "csgo" folder demos are copied into), or an error if it isn't a CS:GO install.
	_ = w.Bind("saveConfig", func(path string, preroll, shiftRadius float64) (string, error) {
		cfg := launcher.LoadConfig() // keep fields this dialog doesn't edit
		cfg.CSGOPath = ""
		if p := strings.Trim(strings.TrimSpace(path), `"`); p != "" {
			dir := launcher.CleanCSGODir(p)
			if st, err := os.Stat(dir); err != nil || !st.IsDir() {
				return "", fmt.Errorf("la carpeta %s no existe", dir)
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "csgo.exe")); err != nil {
				if _, err := os.Stat(filepath.Join(dir, "csgo.exe")); err != nil {
					return "", fmt.Errorf("no encontré csgo.exe en %s; elegí csgo.exe con el botón Buscar", filepath.Dir(dir))
				}
			}
			cfg.CSGOPath = dir
		}
		cfg.PrerollSeconds = preroll
		cfg.ShiftRadius = shiftRadius
		applog.Printf("config: carpeta de CS:GO %q", cfg.CSGOPath)
		return cfg.CSGOPath, launcher.SaveConfig(cfg)
	})

	_ = w.Bind("saveSidebarWidth", func(width int) error {
		cfg := launcher.LoadConfig()
		cfg.SidebarWidth = width
		return launcher.SaveConfig(cfg)
	})

	// getRadar returns the radar of a map from the installed CS:GO, for the minimap.
	_ = w.Bind("getRadar", func(mapName string) (map[string]any, error) {
		dir := launcher.LoadConfig().CSGOPath
		if dir == "" {
			dir, _ = launcher.FindCSGODirectory()
		}
		if dir == "" {
			return map[string]any{"error": "Configurá la carpeta de CS:GO en Configuración para ver el radar del mapa."}, nil
		}
		r, err := radar.Load(dir, mapName)
		if err != nil {
			return map[string]any{"error": err.Error()}, nil
		}
		return map[string]any{"radar": r}, nil
	})

	_ = w.Bind("browseCSGOPath", func() (string, error) {
		path, err := OpenCSGODialog(uintptr(w.Window()))
		if err != nil || path == "" {
			return "", nil
		}
		return launcher.CleanCSGODir(path), nil
	})

	_ = w.Bind("launchInGame", func(tick int) error {
		rep, dPath := current()
		if dPath == "" {
			return nil
		}
		tickrate := 0
		if rep != nil {
			tickrate = rep.Header.Tickrate
		}
		if err := launcher.LaunchInGame(dPath, tick, tickrate); err != nil {
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
