package db

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
	"sourcesight/pkg/core"
)

type InfractionRecord struct {
	Date      string   `json:"date"`
	MapName   string   `json:"mapName"`
	DemoName  string   `json:"demoName"`
	Flags     []string `json:"flags"`
	PerfRatio float64  `json:"perfRatio"`
	// MatchID links the record to core.DemoReport.Header.MatchID; empty for records
	// written before matches were fingerprinted.
	MatchID string `json:"matchId,omitempty"`
}

type PlayerProfile struct {
	SteamID      uint64             `json:"steamId"`
	LastSeenName string             `json:"lastSeenName"`
	KnownNames   []string           `json:"knownNames"`
	FirstSeen    string             `json:"firstSeen"`
	LastSeen     string             `json:"lastSeen"`
	MatchCount   int                `json:"matchCount"`
	Infractions  []InfractionRecord `json:"infractions"`
	// MatchIDs lists the fingerprinted matches already counted in MatchCount, so the
	// same match is never counted twice and can be removed again.
	MatchIDs []string `json:"matchIds,omitempty"`
	// Confirmed is set by a staff member after reviewing the demo; the app never sets
	// it on its own.
	Confirmed *Confirmation `json:"confirmed,omitempty"`
}

// Confirmation records that staff confirmed a player as a cheater.
type Confirmation struct {
	Date string `json:"date"`
	Note string `json:"note,omitempty"`
	Name string `json:"name"` // player name when confirmed
}

type PlayerDatabase struct {
	mu       sync.RWMutex
	filePath string
	Players  map[uint64]*PlayerProfile `json:"players"`
}

var (
	globalDB   *PlayerDatabase
	globalOnce sync.Once
)

// GetDatabase initializes and returns the singleton player history database.
func GetDatabase() *PlayerDatabase {
	globalOnce.Do(func() {
		dbFile := appdata.Path("sourcesight_players.json")

		pdb := &PlayerDatabase{
			filePath: dbFile,
			Players:  make(map[uint64]*PlayerProfile),
		}

		if data, err := os.ReadFile(dbFile); err == nil {
			_ = json.Unmarshal(data, &pdb.Players)
		}
		globalDB = pdb
	})
	return globalDB
}

// LookupPlayer returns a player's known aliases and their infractions from other
// matches (records of excludeMatch are left out, so re-analyzing a demo doesn't make
// a player a reoffender of themselves).
func (pdb *PlayerDatabase) LookupPlayer(steamID uint64, excludeMatch string) ([]string, []InfractionRecord) {
	if steamID == 0 {
		return nil, nil
	}
	pdb.mu.RLock()
	defer pdb.mu.RUnlock()

	p, exists := pdb.Players[steamID]
	if !exists {
		return nil, nil
	}
	var infractions []InfractionRecord
	for _, inf := range p.Infractions {
		if excludeMatch == "" || inf.MatchID != excludeMatch {
			infractions = append(infractions, inf)
		}
	}
	return p.KnownNames, infractions
}

// GetProfile returns a copy of a player's stored profile, or nil if unknown.
func (pdb *PlayerDatabase) GetProfile(steamID uint64) *PlayerProfile {
	pdb.mu.RLock()
	defer pdb.mu.RUnlock()
	p, ok := pdb.Players[steamID]
	if !ok {
		return nil
	}
	cp := *p
	cp.KnownNames = append([]string(nil), p.KnownNames...)
	cp.Infractions = append([]InfractionRecord(nil), p.Infractions...)
	cp.MatchIDs = append([]string(nil), p.MatchIDs...)
	return &cp
}

// HasMatch reports whether a match was already recorded.
func (pdb *PlayerDatabase) HasMatch(matchID string) bool {
	pdb.mu.RLock()
	defer pdb.mu.RUnlock()
	for _, p := range pdb.Players {
		if contains(p.MatchIDs, matchID) {
			return true
		}
	}
	return false
}

// RecordReport counts every player of a match and stores their infractions. A match
// that was already recorded (same MatchID) is ignored, so it can never count twice.
func (pdb *PlayerDatabase) RecordReport(rep *core.DemoReport) {
	matchID := rep.Header.MatchID
	pdb.mu.Lock()
	defer pdb.mu.Unlock()

	if matchID != "" {
		for _, p := range pdb.Players {
			if contains(p.MatchIDs, matchID) {
				return
			}
		}
	}

	now := time.Now()
	for _, ps := range rep.Players {
		if ps.SteamID == 0 {
			continue
		}
		p, exists := pdb.Players[ps.SteamID]
		if !exists {
			p = &PlayerProfile{SteamID: ps.SteamID, FirstSeen: now.Format("2006-01-02 15:04:05")}
			pdb.Players[ps.SteamID] = p
		}
		p.LastSeen = now.Format("2006-01-02 15:04:05")
		p.LastSeenName = ps.Name
		p.MatchCount++
		if matchID != "" {
			p.MatchIDs = append(p.MatchIDs, matchID)
		}
		if ps.Name != "" && !contains(p.KnownNames, ps.Name) {
			p.KnownNames = append(p.KnownNames, ps.Name)
		}
		if len(ps.Flags) > 0 {
			p.Infractions = append(p.Infractions, InfractionRecord{
				Date:      now.Format("2006-01-02"),
				MapName:   rep.Header.MapName,
				DemoName:  rep.Header.DemoFile,
				Flags:     ps.Flags,
				PerfRatio: ps.PerfRatio,
				MatchID:   matchID,
			})
		}
	}
	pdb.saveUnlocked()
}

// RemoveMatch undoes RecordReport for a match: it is subtracted from each player's
// match count and its infractions are dropped. Used when a duplicate replaces it.
func (pdb *PlayerDatabase) RemoveMatch(matchID string) {
	if matchID == "" {
		return
	}
	pdb.mu.Lock()
	defer pdb.mu.Unlock()

	changed := false
	for id, p := range pdb.Players {
		if !contains(p.MatchIDs, matchID) {
			continue
		}
		changed = true
		p.MatchIDs = remove(p.MatchIDs, matchID)
		if p.MatchCount > 0 {
			p.MatchCount--
		}
		kept := p.Infractions[:0]
		for _, inf := range p.Infractions {
			if inf.MatchID != matchID {
				kept = append(kept, inf)
			}
		}
		p.Infractions = kept
		if p.MatchCount == 0 && len(p.Infractions) == 0 && p.Confirmed == nil {
			delete(pdb.Players, id)
		}
	}
	if changed {
		pdb.saveUnlocked()
	}
}

func (pdb *PlayerDatabase) saveUnlocked() {
	if data, err := json.MarshalIndent(pdb.Players, "", "  "); err == nil {
		_ = os.WriteFile(pdb.filePath, data, 0644)
	}
}

func contains(list []string, s string) bool { return slices.Contains(list, s) }

func remove(list []string, s string) []string {
	out := list[:0]
	for _, v := range list {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// SetConfirmed marks (or, with confirmed=false, unmarks) a player as a confirmed
// cheater. Players not yet in the database are created so the mark is kept.
func (pdb *PlayerDatabase) SetConfirmed(steamID uint64, name, note string, confirmed bool) {
	if steamID == 0 {
		return
	}
	pdb.mu.Lock()
	defer pdb.mu.Unlock()
	p, ok := pdb.Players[steamID]
	if !ok {
		if !confirmed {
			return
		}
		p = &PlayerProfile{SteamID: steamID, LastSeenName: name, KnownNames: []string{name}}
		pdb.Players[steamID] = p
	}
	if confirmed {
		p.Confirmed = &Confirmation{Date: time.Now().Format("2006-01-02"), Note: note, Name: name}
	} else {
		p.Confirmed = nil
	}
	pdb.saveUnlocked()
}

// Confirmations returns every confirmed player keyed by SteamID64 as a string (so
// JavaScript doesn't round it).
func (pdb *PlayerDatabase) Confirmations() map[string]Confirmation {
	pdb.mu.RLock()
	defer pdb.mu.RUnlock()
	out := map[string]Confirmation{}
	for id, p := range pdb.Players {
		if p.Confirmed != nil {
			out[strconv.FormatUint(id, 10)] = *p.Confirmed
		}
	}
	return out
}
