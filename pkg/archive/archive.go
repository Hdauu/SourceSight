// Package archive keeps the demos worth reviewing from a server's public demo page
// (built for csgo.3y33.com.ar/demos, which deletes demos after about 3 days).
//
// Every run it reads the list, downloads the demos it hasn't seen yet that pass the
// filters (server, mode, map, minimum size), one at a time, and hands each to the
// caller to analyze and store in the history. The file is kept in the chosen folder
// only when some player came out as cheat or suspicious; otherwise it is deleted
// right away, so disk use stays small. What was already handled is remembered, and a
// lock file keeps the app and the bot on the same PC from downloading the same demo.
package archive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
	"sourcesight/pkg/core"
)

// DefaultURL is the demo page of 3y33's CS:GO servers.
const DefaultURL = "https://csgo.3y33.com.ar/demos"

// Config is shared by the app (settings) and the bot (bot_config.json).
type Config struct {
	Enabled  bool     `json:"enabled"`
	URL      string   `json:"url,omitempty"`      // demo page; DefaultURL if empty
	Dir      string   `json:"dir"`                // where kept demos go
	Servers  []int    `json:"servers,omitempty"`  // server numbers; empty = all
	Modes    []string `json:"modes,omitempty"`    // "Competitivo", "Retake"…; empty = all
	Maps     []string `json:"maps,omitempty"`     // "de_mirage"…; empty = all
	MinMB    int      `json:"minMB,omitempty"`    // skip smaller demos (warmups, empty servers); default 5
	EveryMin int      `json:"everyMin,omitempty"` // minutes between checks; default 60
}

// Demo is one entry of the demo page.
type Demo struct {
	Server int
	Mode   string
	Map    string
	File   string
	SizeMB float64
	URL    string
}

// Result is what happened to one demo.
type Result struct {
	Demo  Demo
	Kept  bool   // file kept in Config.Dir
	Level string // worst player level
	Flags []string
	Path  string // kept file
	Err   error
}

// Analyze parses a downloaded demo and stores it in the history; the archive only
// needs the report back to decide whether to keep the file.
type Analyze func(path string) (*core.DemoReport, error)

const userAgent = "SourceSight (archivo de demos)"

var (
	cardRe  = regexp.MustCompile(`(?s)<article class="demo[^"]*"(.*?)</article>`)
	srvRe   = regexp.MustCompile(`(?s)class="srv-tag"[^>]*>\s*<b[^>]*>#(\d+)</b>\s*([^<]*)<`)
	mapRe   = regexp.MustCompile(`(?s)class="d-map".*?<span[^>]*>([^<]+)</span>`)
	sizeRe  = regexp.MustCompile(`<em[^>]*>([\d.,]+)\s*(KB|MB|GB)</em>`)
	linkRe  = regexp.MustCompile(`href="(/dl/[^"]+\.dem)"`)
	nextRe  = regexp.MustCompile(`href="[^"]*[?&]page=(\d+)"`)
	runMu   sync.Mutex
	lockTTL = 3 * time.Hour
)

// ParsePage reads the demo cards of one page of the list.
func ParsePage(page string, base *url.URL) []Demo {
	var out []Demo
	for _, m := range cardRe.FindAllStringSubmatch(page, -1) {
		card := m[1]
		link := linkRe.FindStringSubmatch(card)
		if link == nil {
			continue
		}
		u, err := base.Parse(html.UnescapeString(link[1]))
		if err != nil {
			continue
		}
		d := Demo{URL: u.String(), File: filepath.Base(u.Path)}
		if s := srvRe.FindStringSubmatch(card); s != nil {
			d.Server, _ = strconv.Atoi(s[1])
			d.Mode = strings.TrimSpace(html.UnescapeString(s[2]))
		}
		if s := mapRe.FindStringSubmatch(card); s != nil {
			d.Map = strings.TrimSpace(s[1])
		}
		if s := sizeRe.FindStringSubmatch(card); s != nil {
			v, _ := strconv.ParseFloat(strings.ReplaceAll(s[1], ",", "."), 64)
			switch s[2] {
			case "KB":
				v /= 1024
			case "GB":
				v *= 1024
			}
			d.SizeMB = v
		}
		out = append(out, d)
	}
	return out
}

// Wanted reports whether a demo passes the filters.
func (c Config) Wanted(d Demo) bool {
	minMB := c.MinMB
	if minMB <= 0 {
		minMB = 5
	}
	if d.SizeMB > 0 && d.SizeMB < float64(minMB) {
		return false
	}
	if len(c.Servers) > 0 && !slices.Contains(c.Servers, d.Server) {
		return false
	}
	if len(c.Modes) > 0 && !slices.ContainsFunc(c.Modes, func(m string) bool { return strings.EqualFold(m, d.Mode) }) {
		return false
	}
	if len(c.Maps) > 0 && !slices.ContainsFunc(c.Maps, func(m string) bool { return strings.EqualFold(strings.TrimSpace(m), d.Map) }) {
		return false
	}
	return true
}

// Interval is the time between checks.
func (c Config) Interval() time.Duration {
	if c.EveryMin <= 0 {
		return time.Hour
	}
	return time.Duration(max(c.EveryMin, 15)) * time.Minute
}

// seen remembers which files were handled, so they aren't downloaded again.
type seen map[string]time.Time

func seenPath() string { return appdata.Path("archive_seen.json") }

func loadSeen() seen {
	s := seen{}
	if b, err := os.ReadFile(seenPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	// Demos live 3 days on the server: older entries can go
	for f, t := range s {
		if time.Since(t) > 10*24*time.Hour {
			delete(s, f)
		}
	}
	return s
}

func (s seen) save() {
	b, _ := json.Marshal(s)
	tmp := seenPath() + ".tmp"
	if os.WriteFile(tmp, b, 0644) == nil {
		_ = os.Rename(tmp, seenPath())
	}
}

// lock makes sure only one process on this PC runs the archive at a time (the app
// and the bot can both have it enabled). A lock older than lockTTL is stale.
func lock() (func(), bool) {
	p := appdata.Path("archive.lock")
	if st, err := os.Stat(p); err == nil && time.Since(st.ModTime()) > lockTTL {
		_ = os.Remove(p)
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, false
	}
	fmt.Fprintf(f, "%d %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	f.Close()
	return func() { _ = os.Remove(p) }, true
}

// List reads every page of the demo list, stopping early once a whole page has
// been handled before (the list is newest first).
func List(ctx context.Context, c Config, already seen) ([]Demo, error) {
	base, err := url.Parse(c.url())
	if err != nil {
		return nil, err
	}
	var all []Demo
	for page := 1; page <= 50; page++ {
		u := *base
		q := u.Query()
		if page > 1 {
			q.Set("page", strconv.Itoa(page))
		}
		u.RawQuery = q.Encode()
		body, err := get(ctx, u.String())
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		demos := ParsePage(body, base)
		if len(demos) == 0 {
			break
		}
		all = append(all, demos...)
		allSeen := true
		for _, d := range demos {
			if _, ok := already[d.File]; !ok {
				allSeen = false
			}
		}
		if allSeen || !hasPage(body, page+1) {
			break
		}
		time.Sleep(500 * time.Millisecond) // be gentle with the server
	}
	return all, nil
}

func (c Config) url() string {
	if c.URL != "" {
		return c.URL
	}
	return DefaultURL
}

func hasPage(body string, n int) bool {
	for _, m := range nextRe.FindAllStringSubmatch(body, -1) {
		if v, _ := strconv.Atoi(m[1]); v == n {
			return true
		}
	}
	return false
}

func get(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", u, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return string(b), err
}

// download saves a demo to dst, checking it arrived whole.
func download(ctx context.Context, u, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descarga: %s", resp.Status)
	}
	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && resp.ContentLength > 0 && n != resp.ContentLength {
		err = fmt.Errorf("descarga incompleta (%d de %d bytes)", n, resp.ContentLength)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// ErrBusy means another process on this PC is running the archive right now.
var ErrBusy = errors.New("el archivo automático ya está corriendo en esta PC")

// Run does one check: lists the page, downloads and analyzes new wanted demos, keeps
// the flagged ones. onResult is called after each demo.
func Run(ctx context.Context, c Config, analyze Analyze, onResult func(Result)) error {
	if c.Dir == "" {
		return errors.New("elegí la carpeta donde guardar las demos")
	}
	if err := os.MkdirAll(c.Dir, 0755); err != nil {
		return err
	}
	runMu.Lock()
	defer runMu.Unlock()
	unlock, ok := lock()
	if !ok {
		return ErrBusy
	}
	defer unlock()

	already := loadSeen()
	demos, err := List(ctx, c, already)
	if err != nil {
		return err
	}
	// Oldest first: they are the closest to being deleted from the server
	slices.Reverse(demos)
	for _, d := range demos {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if _, done := already[d.File]; done || !c.Wanted(d) {
			continue
		}
		res := handle(ctx, c, d, analyze)
		if errors.Is(res.Err, context.Canceled) {
			return res.Err
		}
		// Failed downloads are retried next time; analyzed demos are done
		if res.Err == nil || !strings.HasPrefix(res.Err.Error(), "descarga") {
			already[d.File] = time.Now()
			already.save()
		}
		if res.Err == nil && res.Kept && res.Level != "" {
			recordKept(res)
		}
		if onResult != nil {
			onResult(res)
		}
	}
	return nil
}

func handle(ctx context.Context, c Config, d Demo, analyze Analyze) Result {
	res := Result{Demo: d}
	dst := filepath.Join(c.Dir, d.File)
	if _, err := os.Stat(dst); err == nil {
		res.Kept, res.Path = true, dst // already in the folder (downloaded by hand)
	} else if err := download(ctx, d.URL, dst); err != nil {
		res.Err = fmt.Errorf("descarga: %w", err)
		return res
	}
	rep, err := analyze(dst)
	if err != nil {
		if !res.Kept {
			_ = os.Remove(dst)
		}
		res.Err = err
		return res
	}
	res.Level = core.LevelClean
	seenFlag := map[string]bool{}
	for _, p := range rep.Players {
		res.Level = core.MaxLevel(res.Level, p.Level)
		for _, f := range p.Flags {
			if !seenFlag[f] {
				seenFlag[f] = true
				res.Flags = append(res.Flags, f)
			}
		}
	}
	keep := res.Level == core.LevelCheat || res.Level == core.LevelSuspect
	switch {
	case keep:
		res.Kept, res.Path = true, dst
	case !res.Kept:
		_ = os.Remove(dst)
	}
	return res
}

// KeptDemo is one demo the archive kept, for the list in the app.
type KeptDemo struct {
	File   string    `json:"file"`
	Path   string    `json:"path"`
	Level  string    `json:"level"`
	Flags  []string  `json:"flags,omitempty"`
	Server int       `json:"server"`
	Mode   string    `json:"mode"`
	Map    string    `json:"map"`
	SizeMB float64   `json:"sizeMB"`
	At     time.Time `json:"at"`
}

func keptPath() string { return appdata.Path("archive_kept.json") }

// recordKept remembers a kept demo (called under the run lock).
func recordKept(r Result) {
	list := loadKept()
	k := KeptDemo{File: r.Demo.File, Path: r.Path, Level: r.Level, Flags: r.Flags, Server: r.Demo.Server,
		Mode: r.Demo.Mode, Map: r.Demo.Map, SizeMB: r.Demo.SizeMB, At: time.Now()}
	list = slices.DeleteFunc(list, func(o KeptDemo) bool { return strings.EqualFold(o.Path, k.Path) })
	list = append(list, k)
	if len(list) > 2000 {
		list = list[len(list)-2000:]
	}
	b, _ := json.Marshal(list)
	tmp := keptPath() + ".tmp"
	if os.WriteFile(tmp, b, 0644) == nil {
		_ = os.Rename(tmp, keptPath())
	}
}

func loadKept() []KeptDemo {
	var list []KeptDemo
	if b, err := os.ReadFile(keptPath()); err == nil {
		_ = json.Unmarshal(b, &list)
	}
	return list
}

// Kept lists the demos the archive kept that are still on disk, newest first.
func Kept() []KeptDemo {
	var out []KeptDemo
	for _, k := range loadKept() {
		if _, err := os.Stat(k.Path); err == nil {
			out = append(out, k)
		}
	}
	slices.Reverse(out)
	return out
}
