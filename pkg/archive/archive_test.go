package archive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sourcesight/pkg/core"
)

func TestParsePage(t *testing.T) {
	b, err := os.ReadFile("testdata/page.html")
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse("https://csgo.3y33.com.ar/demos")
	demos := ParsePage(string(b), base)
	if len(demos) != 3 {
		t.Fatalf("demos = %d, want 3", len(demos))
	}
	d := demos[1]
	if d.Server != 2 || d.Mode != "Competitivo" || d.Map != "de_mirage" || d.File != "s2-20261112-0408-de_mirage.dem" ||
		d.URL != "https://csgo.3y33.com.ar/dl/s2/s2-20261112-0408-de_mirage.dem" || d.SizeMB < 120.6 || d.SizeMB > 120.8 {
		t.Fatalf("demo = %+v", d)
	}
	if demos[0].Mode != "Retake" || demos[2].SizeMB > 1 {
		t.Fatalf("demos = %+v", demos)
	}
	if !hasPage(string(b), 2) || hasPage(string(b), 9) {
		t.Fatal("pagination not detected")
	}
}

func TestWanted(t *testing.T) {
	comp := Demo{Server: 2, Mode: "Competitivo", Map: "de_mirage", SizeMB: 120}
	tiny := Demo{Server: 1, Mode: "Competitivo", Map: "de_dust2", SizeMB: 0.9}
	retake := Demo{Server: 6, Mode: "Retake", Map: "de_tuscan", SizeMB: 44}

	all := Config{}
	if !all.Wanted(comp) || !all.Wanted(retake) || all.Wanted(tiny) {
		t.Fatal("no filters: every demo but the tiny one")
	}
	onlyComp := Config{Modes: []string{"competitivo"}}
	if !onlyComp.Wanted(comp) || onlyComp.Wanted(retake) {
		t.Fatal("mode filter")
	}
	srv := Config{Servers: []int{6}}
	if srv.Wanted(comp) || !srv.Wanted(retake) {
		t.Fatal("server filter")
	}
	maps := Config{Maps: []string{" de_mirage"}}
	if !maps.Wanted(comp) || maps.Wanted(retake) {
		t.Fatal("map filter")
	}
}

func TestRunKeepsOnlyFlaggedDemos(t *testing.T) {
	page := `<article class="demo"><span class="srv-tag"><b>#2</b>Competitivo</span><div class="d-map"><span>de_mirage</span></div><em>20 MB</em><a href="/dl/s2/cheat.dem">x</a></article>` +
		`<article class="demo"><span class="srv-tag"><b>#2</b>Competitivo</span><div class="d-map"><span>de_mirage</span></div><em>20 MB</em><a href="/dl/s2/clean.dem">x</a></article>` +
		`<article class="demo"><span class="srv-tag"><b>#6</b>Retake</span><div class="d-map"><span>de_nuke</span></div><em>20 MB</em><a href="/dl/s6/retake.dem">x</a></article>`
	downloads := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/dl/") {
			downloads[r.URL.Path]++
			_, _ = w.Write([]byte("demo " + r.URL.Path))
			return
		}
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()

	dir := t.TempDir()
	cfg := Config{Enabled: true, URL: srv.URL + "/demos", Dir: dir, Modes: []string{"Competitivo"}}
	analyze := func(path string) (*core.DemoReport, error) {
		level := core.LevelClean
		if strings.Contains(path, "cheat") {
			level = core.LevelCheat
		}
		return &core.DemoReport{Players: []core.PlayerStats{{Name: "p", Level: level, Flags: []string{"AIMLOCK"}}}}, nil
	}
	var results []Result
	if err := Run(context.Background(), cfg, analyze, func(r Result) { results = append(results, r) }); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %+v", results)
	}
	if _, err := os.Stat(filepath.Join(dir, "cheat.dem")); err != nil {
		t.Fatal("flagged demo not kept")
	}
	if _, err := os.Stat(filepath.Join(dir, "clean.dem")); err == nil {
		t.Fatal("clean demo kept")
	}
	if downloads["/dl/s6/retake.dem"] != 0 {
		t.Fatal("filtered-out demo downloaded")
	}

	// Second run: nothing new to download
	if err := Run(context.Background(), cfg, analyze, nil); err != nil {
		t.Fatal(err)
	}
	if downloads["/dl/s2/cheat.dem"] != 1 || downloads["/dl/s2/clean.dem"] != 1 {
		t.Fatalf("downloads = %v, each demo once", downloads)
	}
}
