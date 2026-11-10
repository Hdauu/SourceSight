package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
)

type InfractionRecord struct {
	Date      string   `json:"date"`
	MapName   string   `json:"mapName"`
	DemoName  string   `json:"demoName"`
	Flags     []string `json:"flags"`
	PerfRatio float64  `json:"perfRatio"`
}

type PlayerProfile struct {
	SteamID      uint64             `json:"steamId"`
	LastSeenName string             `json:"lastSeenName"`
	KnownNames   []string           `json:"knownNames"`
	FirstSeen    string             `json:"firstSeen"`
	LastSeen     string             `json:"lastSeen"`
	MatchCount   int                `json:"matchCount"`
	Infractions  []InfractionRecord `json:"infractions"`
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

// LookupPlayer returns the known previous aliases and previous infractions for a SteamID.
func (pdb *PlayerDatabase) LookupPlayer(steamID uint64) ([]string, []InfractionRecord) {
	if steamID == 0 {
		return nil, nil
	}
	pdb.mu.RLock()
	defer pdb.mu.RUnlock()

	p, exists := pdb.Players[steamID]
	if !exists {
		return nil, nil
	}
	return p.KnownNames, p.Infractions
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
	return &cp
}

// RecordPlayerMatch updates the database with the current match outcome and saves to disk.
func (pdb *PlayerDatabase) RecordPlayerMatch(
	steamID uint64,
	currentName string,
	mapName, demoName string,
	flags []string,
	perfRatio float64,
) {
	if steamID == 0 {
		return
	}
	pdb.mu.Lock()
	defer pdb.mu.Unlock()

	p, exists := pdb.Players[steamID]
	nowStr := time.Now().Format("2006-01-02 15:04:05")

	if !exists {
		p = &PlayerProfile{
			SteamID:      steamID,
			LastSeenName: currentName,
			KnownNames:   []string{currentName},
			FirstSeen:    nowStr,
			LastSeen:     nowStr,
			MatchCount:   1,
		}
		pdb.Players[steamID] = p
	} else {
		p.LastSeen = nowStr
		p.LastSeenName = currentName
		p.MatchCount++

		// Track new alias if changed
		hasName := false
		for _, n := range p.KnownNames {
			if n == currentName {
				hasName = true
				break
			}
		}
		if !hasName && currentName != "" {
			p.KnownNames = append(p.KnownNames, currentName)
		}
	}

	// Record infraction if flags detected
	if len(flags) > 0 {
		p.Infractions = append(p.Infractions, InfractionRecord{
			Date:      time.Now().Format("2006-01-02"),
			MapName:   mapName,
			DemoName:  filepath.Base(demoName),
			Flags:     flags,
			PerfRatio: perfRatio,
		})
	}

	// Auto-flush to disk
	pdb.saveUnlocked()
}

func (pdb *PlayerDatabase) saveUnlocked() {
	if data, err := json.MarshalIndent(pdb.Players, "", "  "); err == nil {
		_ = os.WriteFile(pdb.filePath, data, 0644)
	}
}
