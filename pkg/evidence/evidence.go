// Package evidence writes a self-contained, printable HTML file documenting one
// player's detections in a demo, for sharing with other staff members.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"sourcesight/pkg/core"
	"sourcesight/pkg/version"
)

type incidentRow struct {
	Round   int
	Tick    int
	Type    string
	Note    string
	GotoCmd string
}

type typeCount struct {
	Type  string
	Count int
}

type page struct {
	Player      core.PlayerStats
	Header      core.DemoHeader
	DemoName    string
	DemoSHA256  string
	GeneratedAt string
	Version     string
	RepoURL     string
	Incidents   []incidentRow
	ByType      []typeCount
}

var unsafeChars = regexp.MustCompile(`[^\p{L}\p{N}_-]+`)

// RoundAt maps a tick to its 1-based round using the report's round start ticks.
func RoundAt(roundStarts []int, tick int) int {
	return sort.Search(len(roundStarts), func(i int) bool { return roundStarts[i] > tick })
}

// Write renders the evidence file for one player into dir and returns its path.
func Write(dir string, rep *core.DemoReport, p core.PlayerStats, demoPath string) (string, error) {
	sum, err := fileSHA256(demoPath)
	if err != nil {
		sum = "no disponible (" + err.Error() + ")"
	}

	counts := map[string]int{}
	var rows []incidentRow
	for _, inc := range p.Incidents {
		rows = append(rows, incidentRow{
			Round:   RoundAt(rep.Header.RoundStarts, inc.Tick),
			Tick:    inc.Tick,
			Type:    inc.Type,
			Note:    inc.Note,
			GotoCmd: inc.GotoCmd,
		})
		counts[inc.Type]++
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Tick < rows[j].Tick })
	var byType []typeCount
	for t, c := range counts {
		byType = append(byType, typeCount{t, c})
	}
	sort.Slice(byType, func(i, j int) bool { return byType[i].Count > byType[j].Count })

	name := unsafeChars.ReplaceAllString(p.Name, "_")
	if name == "" || name == "_" {
		name = fmt.Sprint(p.SteamID)
	}
	out := filepath.Join(dir, fmt.Sprintf("evidencia_%s_%s.html", name, time.Now().Format("2006-01-02_150405")))
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = tmpl.Execute(f, page{
		Player:      p,
		Header:      rep.Header,
		DemoName:    filepath.Base(demoPath),
		DemoSHA256:  sum,
		GeneratedAt: time.Now().Format("2006-01-02 15:04"),
		Version:     version.Version,
		RepoURL:     version.RepoURL,
		Incidents:   rows,
		ByType:      byType,
	})
	return out, err
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

var tmpl = template.Must(template.New("evidence").Funcs(template.FuncMap{
	"pct": func(v float64) string { return fmt.Sprintf("%.1f%%", v) },
}).Parse(`<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Evidencia — {{.Player.Name}}</title>
<style>
  :root { --ink:#1d2024; --mute:#5f6670; --line:#d9dce1; --bg:#f6f6f4; --red:#b42318; }
  * { box-sizing:border-box; }
  body { margin:0; background:var(--bg); color:var(--ink); font:14px/1.5 "Segoe UI", system-ui, sans-serif; }
  .page { max-width:960px; margin:0 auto; padding:40px 32px 56px; background:#fff; min-height:100vh; }
  h1 { font-size:24px; margin:0 0 4px; }
  h2 { font-size:13px; text-transform:uppercase; letter-spacing:.06em; color:var(--mute); margin:32px 0 10px; }
  a { color:inherit; }
  .sub { color:var(--mute); }
  .flags { margin:14px 0 0; display:flex; gap:6px; flex-wrap:wrap; }
  .flag { border:1px solid var(--red); color:var(--red); font-weight:600; font-size:12px; padding:2px 8px; border-radius:3px; }
  dl { display:grid; grid-template-columns:180px 1fr; gap:6px 16px; margin:0; }
  dt { color:var(--mute); }
  dd { margin:0; word-break:break-all; }
  table { width:100%; border-collapse:collapse; font-size:13px; }
  th, td { text-align:left; padding:7px 10px; border-bottom:1px solid var(--line); vertical-align:top; }
  th { font-size:12px; color:var(--mute); font-weight:600; }
  td.num, th.num { text-align:right; font-variant-numeric:tabular-nums; }
  code { font:12px Consolas, "Cascadia Mono", monospace; background:#f0f0ee; padding:1px 5px; border-radius:3px; white-space:nowrap; }
  .stats { display:grid; grid-template-columns:repeat(4, 1fr); gap:10px; }
  .stat { border:1px solid var(--line); border-radius:4px; padding:10px 12px; }
  .stat b { display:block; font-size:18px; }
  .stat span { color:var(--mute); font-size:12px; }
  .note { margin-top:36px; padding-top:14px; border-top:1px solid var(--line); color:var(--mute); font-size:12px; }
  @media print { body { background:#fff; } .page { padding:0; } }
  @media (max-width:640px) { .stats { grid-template-columns:repeat(2, 1fr); } dl { grid-template-columns:1fr; } }
</style>
</head>
<body>
<div class="page">
  <h1>{{.Player.Name}}</h1>
  <div class="sub">SteamID64 {{.Player.SteamID}}{{if .Player.SteamID}} · <a href="https://steamcommunity.com/profiles/{{.Player.SteamID}}">perfil de Steam</a>{{end}}</div>
  {{if .Player.Flags}}<div class="flags">{{range .Player.Flags}}<span class="flag">{{.}}</span>{{end}}</div>{{end}}

  <h2>Demo</h2>
  <dl>
    <dt>Archivo</dt><dd>{{.DemoName}}</dd>
    <dt>SHA-256</dt><dd><code>{{.DemoSHA256}}</code></dd>
    <dt>Mapa</dt><dd>{{.Header.MapName}}</dd>
    <dt>Servidor</dt><dd>{{.Header.ServerName}}</dd>
    <dt>Tickrate</dt><dd>{{.Header.Tickrate}}</dd>
    <dt>Ticks totales</dt><dd>{{.Header.TotalTicks}}</dd>
  </dl>

  <h2>Métricas</h2>
  <div class="stats">
    <div class="stat"><b>{{pct .Player.PerfRatio}}</b><span>bhop perfectos ({{.Player.PerfectHops}}/{{.Player.TotalJumps}})</span></div>
    <div class="stat"><b>{{.Player.MaxStreak}}</b><span>racha máxima de perfs</span></div>
    <div class="stat"><b>{{.Player.SnapCount}}</b><span>giros de más de 45° en un tick</span></div>
    <div class="stat"><b>{{.Player.SilentCount}}</b><span>silent aim</span></div>
    <div class="stat"><b>{{if .Player.TriggerAvgMs}}{{.Player.TriggerAvgMs}} ms{{else}}—{{end}}</b><span>reacción promedio ({{.Player.TriggerSamples}} muestras)</span></div>
    <div class="stat"><b>{{.Player.WallTrackCount}}</b><span>seguimientos tras pared</span></div>
    <div class="stat"><b>{{.Player.FlashKills}} / {{.Player.SmokeKills}}</b><span>kills cegado / por humo</span></div>
    <div class="stat"><b>{{.Player.MacroHits}}</b><span>series de macro</span></div>
  </div>

  {{if .ByType}}
  <h2>Resumen de incidentes</h2>
  <table>
    <tr><th>Tipo</th><th class="num">Cantidad</th></tr>
    {{range .ByType}}<tr><td>{{.Type}}</td><td class="num">{{.Count}}</td></tr>{{end}}
  </table>
  {{end}}

  <h2>Incidentes ({{len .Incidents}})</h2>
  {{if .Incidents}}
  <table>
    <tr><th class="num">Ronda</th><th class="num">Tick</th><th>Tipo</th><th>Detalle</th><th>Comando</th></tr>
    {{range .Incidents}}<tr><td class="num">{{.Round}}</td><td class="num">{{.Tick}}</td><td>{{.Type}}</td><td>{{.Note}}</td><td><code>{{.GotoCmd}}</code></td></tr>{{end}}
  </table>
  {{else}}<p class="sub">Sin incidentes registrados.</p>{{end}}

  <p class="note">
    Generado el {{.GeneratedAt}} con SourceSight {{.Version}} ({{.RepoURL}}).
    Para revisarlo: abrí la demo en CS:GO con <code>playdemo</code> y pegá el comando de cada incidente en la consola.
    El SHA-256 permite comprobar que se revisa exactamente el mismo archivo. Los resultados son indicios automáticos y
    se tienen que confirmar mirando la demo.
  </p>
</div>
</body>
</html>`))
