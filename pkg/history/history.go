// Package history remembers every analyzed demo and caches its report, so a demo
// can be reopened instantly, duplicates can be detected and players can be followed
// across every match.
package history

import (
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
	"sourcesight/pkg/core"
)

// maxEntries bounds the cache; the oldest reports are dropped beyond this.
const maxEntries = 300

// PlayerSummary is the per-player slice of a report kept in the index, enough to
// aggregate players across demos without decoding every cached report.
type PlayerSummary struct {
	SteamID         string   `json:"steamId"`
	Name            string   `json:"name"`
	Flags           []string `json:"flags,omitempty"`
	Level           string   `json:"level"`
	TotalJumps      int      `json:"totalJumps"`
	PerfRatio       float64  `json:"perfRatio"`
	MaxStreak       int      `json:"maxStreak"`
	TriggerAvgMs    *int     `json:"triggerAvgMs,omitempty"`
	TriggerSamples  int      `json:"triggerSamples"`
	SnapCount       int      `json:"snapCount"`
	SilentCount     int      `json:"silentCount"`
	WallTrackCount  int      `json:"wallTrackCount"`
	MacroHits       int      `json:"macroHits"`
	FlashKills      int      `json:"flashKills"`
	SmokeKills      int      `json:"smokeKills"`
	PeekEngagements int      `json:"peekEngagements,omitempty"`
	Prefires        int      `json:"prefires,omitempty"`
	Incidents       int      `json:"incidents"`
}

type Entry struct {
	ID         string   `json:"id"`
	DemoPath   string   `json:"demoPath"`
	DemoName   string   `json:"demoName"`
	SizeBytes  int64    `json:"sizeBytes"`
	FileSHA256 string   `json:"fileSha256,omitempty"`
	MatchID    string   `json:"matchId,omitempty"`
	AnalyzedAt string   `json:"analyzedAt"`
	MapName    string   `json:"mapName"`
	Server     string   `json:"server"`
	Tickrate   int      `json:"tickrate"`
	TotalTicks int      `json:"totalTicks"`
	Rounds     int      `json:"rounds"`
	Minutes    int      `json:"minutes"`
	Players    int      `json:"players"`
	Flagged    []string `json:"flagged"`
	// Level is the most serious player level in the demo (core.Level*)
	Level string `json:"level"`
	// FlagTypes lists the distinct detections in the demo (AIMLOCK, WALLHACK, ...)
	FlagTypes []string        `json:"flagTypes,omitempty"`
	Summaries []PlayerSummary `json:"summaries,omitempty"`
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
	sum := sha1.Sum(fmt.Appendf(nil, "%s|%d|%d", strings.ToLower(abs), st.Size(), st.ModTime().UnixNano()))
	return hex.EncodeToString(sum[:8]), nil
}

// FileSHA256 hashes a demo file; identical copies under other names share it.
func FileSHA256(path string) (string, error) {
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

// load reads the index once and fills in data older versions didn't store
// (fingerprint, round count, player summaries) from the cached reports.
func load() {
	if loaded {
		return
	}
	loaded = true
	if data, err := os.ReadFile(indexPath()); err == nil {
		_ = json.Unmarshal(data, &items)
	}
	migrated := false
	for i := range items {
		if items[i].MatchID != "" && items[i].Summaries != nil {
			continue
		}
		rep, err := readReport(items[i].ID)
		if err != nil {
			continue
		}
		if rep.Header.MatchID == "" {
			rep.Header.MatchID = rep.Fingerprint()
		}
		fillFromReport(&items[i], rep)
		migrated = true
	}
	for i := range items {
		derive(&items[i])
	}
	if migrated {
		_ = persist()
	}
}

func persist() error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath(), data, 0644)
}

func fillFromReport(e *Entry, rep *core.DemoReport) {
	e.MatchID = rep.Header.MatchID
	e.MapName = rep.Header.MapName
	e.Server = rep.Header.ServerName
	e.Tickrate = rep.Header.Tickrate
	e.TotalTicks = rep.Header.TotalTicks
	e.Rounds = len(rep.Header.RoundStarts)
	e.Players = len(rep.Players)
	if rep.Header.Tickrate > 0 {
		e.Minutes = rep.Header.TotalTicks / (rep.Header.Tickrate * 60)
	}
	e.Flagged = nil
	e.Summaries = make([]PlayerSummary, 0, len(rep.Players))
	for _, p := range rep.Players {
		if len(p.Flags) > 0 {
			e.Flagged = append(e.Flagged, p.Name)
		}
		if p.SteamID == 0 {
			continue
		}
		e.Summaries = append(e.Summaries, PlayerSummary{
			SteamID:         strconv.FormatUint(p.SteamID, 10),
			Name:            p.Name,
			Flags:           p.Flags,
			TotalJumps:      p.TotalJumps,
			PerfRatio:       math.Round(p.PerfRatio*10) / 10,
			MaxStreak:       p.MaxStreak,
			TriggerAvgMs:    p.TriggerAvgMs,
			TriggerSamples:  p.TriggerSamples,
			SnapCount:       p.SnapCount,
			SilentCount:     p.SilentCount,
			WallTrackCount:  p.WallTrackCount,
			MacroHits:       p.MacroHits,
			FlashKills:      p.FlashKills,
			SmokeKills:      p.SmokeKills,
			PeekEngagements: p.PeekEngagements,
			Prefires:        p.Prefires,
			Incidents:       len(p.Incidents),
		})
	}
	derive(e)
}

// List returns all entries, most recent first, without player summaries.
func List() []Entry {
	mu.Lock()
	defer mu.Unlock()
	load()
	out := make([]Entry, len(items))
	for i, it := range items {
		it.Summaries = nil
		out[i] = it
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].AnalyzedAt > out[j].AnalyzedAt })
	return out
}

// all returns a copy of every entry including summaries. Caller must not hold mu.
func all() []Entry {
	mu.Lock()
	defer mu.Unlock()
	load()
	return append([]Entry(nil), items...)
}

// Lookup returns the cached report for a demo file if that exact file was analyzed.
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

// FindByFileHash returns an entry for a byte-identical demo, possibly stored under
// another name or folder.
func FindByFileHash(sha string) *Entry {
	if sha == "" {
		return nil
	}
	for _, e := range all() {
		if e.FileSHA256 == sha {
			e.Summaries = nil
			return &e
		}
	}
	return nil
}

// FindSimilar returns stored entries that look like the same match as rep: the same
// fingerprint, or the same map and round count with nearly the same length and
// players (a second recording or a trimmed copy of the match).
func FindSimilar(rep *core.DemoReport, excludeID string) []Entry {
	ids := map[string]bool{}
	for _, p := range rep.Players {
		if p.SteamID != 0 {
			ids[strconv.FormatUint(p.SteamID, 10)] = true
		}
	}
	rounds := len(rep.Header.RoundStarts)

	var out []Entry
	for _, e := range all() {
		if e.ID == excludeID {
			continue
		}
		same := e.MatchID != "" && e.MatchID == rep.Header.MatchID
		if !same && strings.EqualFold(e.MapName, rep.Header.MapName) && e.Rounds == rounds && rounds > 0 {
			ticksClose := math.Abs(float64(e.TotalTicks-rep.Header.TotalTicks)) <= 0.02*float64(rep.Header.TotalTicks)
			same = ticksClose && overlap(ids, e.Summaries) >= 0.8
		}
		if same {
			e.Summaries = nil
			out = append(out, e)
		}
	}
	return out
}

// overlap is the Jaccard similarity between two player sets.
func overlap(ids map[string]bool, summaries []PlayerSummary) float64 {
	if len(ids) == 0 || len(summaries) == 0 {
		return 0
	}
	inter := 0
	for _, s := range summaries {
		if ids[s.SteamID] {
			inter++
		}
	}
	union := len(ids) + len(summaries) - inter
	return float64(inter) / float64(union)
}

// Open loads a cached report by entry ID.
func Open(id string) (*core.DemoReport, *Entry, error) {
	var entry *Entry
	for _, e := range all() {
		if e.ID == id {
			e.Summaries = nil
			entry = &e
			break
		}
	}
	if entry == nil {
		return nil, nil, fmt.Errorf("la demo no está en el historial")
	}
	rep, err := readReport(id)
	if err != nil {
		return nil, nil, err
	}
	if rep.Header.MatchID == "" {
		rep.Header.MatchID = entry.MatchID
	}
	rep.FillLevels()
	return rep, entry, nil
}

func readReport(id string) (*core.DemoReport, error) {
	f, err := os.Open(reportPath(id))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var rep core.DemoReport
	if err := json.NewDecoder(zr).Decode(&rep); err != nil {
		return nil, err
	}
	return &rep, nil
}

// Save stores a fresh report and returns its history entry. fileSHA may be empty.
func Save(demoPath string, rep *core.DemoReport, fileSHA string) (*Entry, error) {
	id, err := ID(demoPath)
	if err != nil {
		return nil, err
	}
	abs, _ := filepath.Abs(demoPath)

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

	entry := Entry{
		ID:         id,
		DemoPath:   abs,
		DemoName:   filepath.Base(abs),
		FileSHA256: fileSHA,
		AnalyzedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	if st, err := os.Stat(abs); err == nil {
		entry.SizeBytes = st.Size()
	}
	fillFromReport(&entry, rep)

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
	out := entry
	out.Summaries = nil
	return &out, persist()
}

// Remove forgets an entry and deletes its cached report. It returns the removed
// entry's MatchID so the caller can also take it out of the player database.
func Remove(id string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	load()
	matchID := ""
	kept := items[:0]
	for _, it := range items {
		if it.ID == id {
			matchID = it.MatchID
			continue
		}
		kept = append(kept, it)
	}
	items = kept
	_ = os.Remove(reportPath(id))
	return matchID, persist()
}

// flagTypes returns the distinct detections across a demo's players, sorted.
func flagTypes(summaries []PlayerSummary) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range summaries {
		for _, f := range s.Flags {
			if !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	sort.Strings(out)
	return out
}

// MatchInUse reports whether any stored entry still belongs to the match.
func MatchInUse(matchID string) bool {
	for _, e := range all() {
		if e.MatchID == matchID {
			return true
		}
	}
	return false
}

// derive recomputes the values that follow from the player summaries.
func derive(e *Entry) {
	e.Level = core.LevelClean
	for i := range e.Summaries {
		e.Summaries[i].Level = core.LevelOf(e.Summaries[i].Flags)
		e.Level = core.MaxLevel(e.Level, e.Summaries[i].Level)
	}
	e.FlagTypes = flagTypes(e.Summaries)
}
