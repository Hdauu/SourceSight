// Package history remembers every analyzed demo and caches its report, so a demo
// can be reopened instantly without re-parsing (and without recording its players
// in the reoffender database a second time).
package history

import (
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
	"sourcesight/pkg/core"
)

// maxEntries bounds the cache; the oldest reports are dropped beyond this.
const maxEntries = 300

type Entry struct {
	ID         string   `json:"id"`
	DemoPath   string   `json:"demoPath"`
	DemoName   string   `json:"demoName"`
	SizeBytes  int64    `json:"sizeBytes"`
	AnalyzedAt string   `json:"analyzedAt"`
	MapName    string   `json:"mapName"`
	Server     string   `json:"server"`
	Tickrate   int      `json:"tickrate"`
	Minutes    int      `json:"minutes"`
	Players    int      `json:"players"`
	Flagged    []string `json:"flagged"`
}

var (
	mu     sync.Mutex
	loaded bool
	items  []Entry
)

func indexPath() string { return appdata.Path("sourcesight_history.json") }

func reportPath(id string) string {
	return filepath.Join(appdata.Sub("reportes"), id+".json.gz")
}

// ID identifies a demo file by path, size and modification time: a re-recorded or
// edited file gets a new ID, the same file always maps to the same cached report.
func ID(demoPath string) (string, error) {
	abs, err := filepath.Abs(demoPath)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%d|%d", strings.ToLower(abs), st.Size(), st.ModTime().UnixNano())))
	return hex.EncodeToString(sum[:8]), nil
}

func load() {
	if loaded {
		return
	}
	loaded = true
	if data, err := os.ReadFile(indexPath()); err == nil {
		_ = json.Unmarshal(data, &items)
	}
}

func persist() error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath(), data, 0644)
}

// List returns all entries, most recent first.
func List() []Entry {
	mu.Lock()
	defer mu.Unlock()
	load()
	out := append([]Entry(nil), items...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].AnalyzedAt > out[j].AnalyzedAt })
	return out
}

// Lookup returns the cached report for a demo file if it was analyzed before.
func Lookup(demoPath string) (*core.DemoReport, *Entry) {
	id, err := ID(demoPath)
	if err != nil {
		return nil, nil
	}
	rep, entry, err := Open(id)
	if err != nil {
		return nil, nil
	}
	return rep, entry
}

// Open loads a cached report by entry ID.
func Open(id string) (*core.DemoReport, *Entry, error) {
	mu.Lock()
	load()
	var entry *Entry
	for i := range items {
		if items[i].ID == id {
			e := items[i]
			entry = &e
			break
		}
	}
	mu.Unlock()
	if entry == nil {
		return nil, nil, fmt.Errorf("la demo no está en el historial")
	}

	f, err := os.Open(reportPath(id))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, err
	}
	defer zr.Close()
	var rep core.DemoReport
	if err := json.NewDecoder(zr).Decode(&rep); err != nil {
		return nil, nil, err
	}
	return &rep, entry, nil
}

// Save stores a fresh report and returns its history entry.
func Save(demoPath string, rep *core.DemoReport) (*Entry, error) {
	id, err := ID(demoPath)
	if err != nil {
		return nil, err
	}
	abs, _ := filepath.Abs(demoPath)
	st, _ := os.Stat(abs)

	f, err := os.Create(reportPath(id))
	if err != nil {
		return nil, err
	}
	zw := gzip.NewWriter(f)
	err = json.NewEncoder(zw).Encode(rep)
	if cerr := zw.Close(); err == nil {
		err = cerr
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}

	minutes := 0
	if rep.Header.Tickrate > 0 {
		minutes = rep.Header.TotalTicks / (rep.Header.Tickrate * 60)
	}
	entry := Entry{
		ID:         id,
		DemoPath:   abs,
		DemoName:   filepath.Base(abs),
		AnalyzedAt: time.Now().Format("2006-01-02 15:04:05"),
		MapName:    rep.Header.MapName,
		Server:     rep.Header.ServerName,
		Tickrate:   rep.Header.Tickrate,
		Minutes:    minutes,
		Players:    len(rep.Players),
	}
	if st != nil {
		entry.SizeBytes = st.Size()
	}
	for _, p := range rep.Players {
		if len(p.Flags) > 0 {
			entry.Flagged = append(entry.Flagged, p.Name)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	load()
	kept := items[:0]
	for _, it := range items {
		if it.ID != id {
			kept = append(kept, it)
		}
	}
	items = append(kept, entry)
	if len(items) > maxEntries {
		sort.SliceStable(items, func(i, j int) bool { return items[i].AnalyzedAt > items[j].AnalyzedAt })
		for _, old := range items[maxEntries:] {
			_ = os.Remove(reportPath(old.ID))
		}
		items = items[:maxEntries]
	}
	return &entry, persist()
}

// Remove forgets an entry and deletes its cached report.
func Remove(id string) error {
	mu.Lock()
	defer mu.Unlock()
	load()
	kept := items[:0]
	for _, it := range items {
		if it.ID != id {
			kept = append(kept, it)
		}
	}
	items = kept
	_ = os.Remove(reportPath(id))
	return persist()
}
