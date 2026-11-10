package history

import (
	"math"
	"slices"
	"sort"
	"strings"

	"sourcesight/pkg/core"
)

// minJumpsForPerf excludes demos where a player barely jumped from the bhop average.
const minJumpsForPerf = 5

// PlayerRow is one line of the player list.
type PlayerRow struct {
	SteamID      string   `json:"steamId"`
	Level        string   `json:"level"`
	Name         string   `json:"name"`
	Demos        int      `json:"demos"`
	FlaggedDemos int      `json:"flaggedDemos"`
	FlagTypes    []string `json:"flagTypes,omitempty"`
	LastSeen     string   `json:"lastSeen"`
}

// Appearance is a player's line in one demo.
type Appearance struct {
	EntryID    string   `json:"entryId"`
	Level      string   `json:"level"`
	DemoName   string   `json:"demoName"`
	MapName    string   `json:"mapName"`
	AnalyzedAt string   `json:"analyzedAt"`
	Name       string   `json:"name"`
	Flags      []string `json:"flags,omitempty"`
	PerfRatio  float64  `json:"perfRatio"`
	TotalJumps int      `json:"totalJumps"`
	ReactionMs *int     `json:"reactionMs,omitempty"`
	Snaps      int      `json:"snaps"`
	Wallhack   int      `json:"wallhack"`
	Incidents  int      `json:"incidents"`
}

// Averages summarizes a group of player-demos. Pointers are nil when there is no data.
type Averages struct {
	PerfRatio      *float64 `json:"perfRatio,omitempty"`
	ReactionMs     *float64 `json:"reactionMs,omitempty"`
	SnapsPerDemo   float64  `json:"snapsPerDemo"`
	SilentPerDemo  float64  `json:"silentPerDemo"`
	WallPerDemo    float64  `json:"wallPerDemo"`
	IncidentsPerDm float64  `json:"incidentsPerDemo"`
}

// PlayerDetail aggregates one player across every stored demo.
type PlayerDetail struct {
	SteamID      string         `json:"steamId"`
	Level        string         `json:"level"`
	Name         string         `json:"name"`
	Names        []string       `json:"names"`
	Demos        int            `json:"demos"`
	FlaggedDemos int            `json:"flaggedDemos"`
	FlagCounts   map[string]int `json:"flagCounts"`
	FirstSeen    string         `json:"firstSeen"`
	LastSeen     string         `json:"lastSeen"`
	Totals       struct {
		Snaps     int `json:"snaps"`
		Silent    int `json:"silent"`
		Wallhack  int `json:"wallhack"`
		Macro     int `json:"macro"`
		Blind     int `json:"blind"`
		Smoke     int `json:"smoke"`
		Incidents int `json:"incidents"`
	} `json:"totals"`
	Averages    Averages     `json:"averages"`
	Global      Overview     `json:"global"`
	Appearances []Appearance `json:"appearances"`
}

// Overview is the "all demos" summary used as a baseline.
type Overview struct {
	Demos          int            `json:"demos"`
	FlaggedDemos   int            `json:"flaggedDemos"`
	Players        int            `json:"players"`
	FlaggedPlayers int            `json:"flaggedPlayers"`
	CheatPlayers   int            `json:"cheatPlayers"`
	SuspectPlayers int            `json:"suspectPlayers"`
	CheatDemos     int            `json:"cheatDemos"`
	SuspectDemos   int            `json:"suspectDemos"`
	Averages       Averages       `json:"averages"`
	FlagCounts     map[string]int `json:"flagCounts"`
}

type accum struct {
	demos                          int
	perfSum                        float64
	perfN                          int
	reactSum                       float64
	reactSamples                   int
	snaps, silent, wall, incidents int
}

func (a *accum) add(s PlayerSummary) {
	a.demos++
	if s.TotalJumps >= minJumpsForPerf {
		a.perfSum += s.PerfRatio
		a.perfN++
	}
	if s.TriggerAvgMs != nil && s.TriggerSamples > 0 {
		a.reactSum += float64(*s.TriggerAvgMs) * float64(s.TriggerSamples)
		a.reactSamples += s.TriggerSamples
	}
	a.snaps += s.SnapCount
	a.silent += s.SilentCount
	a.wall += s.WallTrackCount
	a.incidents += s.Incidents
}

func (a *accum) averages() Averages {
	var out Averages
	if a.perfN > 0 {
		v := round1(a.perfSum / float64(a.perfN))
		out.PerfRatio = &v
	}
	if a.reactSamples > 0 {
		v := math.Round(a.reactSum / float64(a.reactSamples))
		out.ReactionMs = &v
	}
	if a.demos > 0 {
		d := float64(a.demos)
		out.SnapsPerDemo = round1(float64(a.snaps) / d)
		out.SilentPerDemo = round1(float64(a.silent) / d)
		out.WallPerDemo = round1(float64(a.wall) / d)
		out.IncidentsPerDm = round1(float64(a.incidents) / d)
	}
	return out
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

// GetOverview averages every player-demo in the history.
func GetOverview() Overview {
	entries := all()
	ov := Overview{Demos: len(entries), FlagCounts: map[string]int{}}
	var acc accum
	players := map[string]bool{}
	flagged := map[string]bool{}
	levels := map[string]string{}
	for _, e := range entries {
		if worthALook(e.Level) {
			ov.FlaggedDemos++
		}
		switch e.Level {
		case core.LevelCheat:
			ov.CheatDemos++
		case core.LevelSuspect:
			ov.SuspectDemos++
		}
		for _, s := range e.Summaries {
			acc.add(s)
			players[s.SteamID] = true
			levels[s.SteamID] = core.MaxLevel(levels[s.SteamID], s.Level)
			if worthALook(s.Level) {
				flagged[s.SteamID] = true
			}
			for _, f := range s.Flags {
				ov.FlagCounts[f]++
			}
		}
	}
	ov.Players = len(players)
	ov.FlaggedPlayers = len(flagged)
	for _, l := range levels {
		switch l {
		case core.LevelCheat:
			ov.CheatPlayers++
		case core.LevelSuspect:
			ov.SuspectPlayers++
		}
	}
	ov.Averages = acc.averages()
	return ov
}

// ListPlayers returns every player seen in the history. query filters by name, or
// by exact SteamID64 when steamID is non-empty.
func ListPlayers(query, steamID string) []PlayerRow {
	rows := map[string]*PlayerRow{}
	names := map[string][]string{}
	for _, e := range all() {
		for _, s := range e.Summaries {
			r, ok := rows[s.SteamID]
			if !ok {
				r = &PlayerRow{SteamID: s.SteamID}
				rows[s.SteamID] = r
			}
			r.Demos++
			r.Level = core.MaxLevel(r.Level, s.Level)
			if e.AnalyzedAt >= r.LastSeen {
				r.LastSeen = e.AnalyzedAt
				r.Name = s.Name
			}
			if worthALook(s.Level) {
				r.FlaggedDemos++
			}
			for _, f := range s.Flags {
				if !containsStr(r.FlagTypes, f) {
					r.FlagTypes = append(r.FlagTypes, f)
				}
			}
			if !containsStr(names[s.SteamID], s.Name) {
				names[s.SteamID] = append(names[s.SteamID], s.Name)
			}
		}
	}

	q := strings.ToLower(strings.TrimSpace(query))
	var out []PlayerRow
	for id, r := range rows {
		if steamID != "" && id != steamID {
			continue
		}
		if steamID == "" && q != "" && !anyContains(names[id], q) {
			continue
		}
		sort.Strings(r.FlagTypes)
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FlaggedDemos != out[j].FlaggedDemos {
			return out[i].FlaggedDemos > out[j].FlaggedDemos
		}
		if out[i].Demos != out[j].Demos {
			return out[i].Demos > out[j].Demos
		}
		return out[i].LastSeen > out[j].LastSeen
	})
	return out
}

// GetPlayer aggregates one player; nil if they appear in no stored demo.
func GetPlayer(steamID string) *PlayerDetail {
	d := &PlayerDetail{SteamID: steamID, FlagCounts: map[string]int{}}
	var acc accum
	for _, e := range all() {
		for _, s := range e.Summaries {
			if s.SteamID != steamID {
				continue
			}
			acc.add(s)
			d.Demos++
			d.Level = core.MaxLevel(d.Level, s.Level)
			if !containsStr(d.Names, s.Name) {
				d.Names = append(d.Names, s.Name)
			}
			if d.FirstSeen == "" || e.AnalyzedAt < d.FirstSeen {
				d.FirstSeen = e.AnalyzedAt
			}
			if e.AnalyzedAt >= d.LastSeen {
				d.LastSeen = e.AnalyzedAt
				d.Name = s.Name
			}
			if worthALook(s.Level) {
				d.FlaggedDemos++
			}
			for _, f := range s.Flags {
				d.FlagCounts[f]++
			}
			d.Totals.Snaps += s.SnapCount
			d.Totals.Silent += s.SilentCount
			d.Totals.Wallhack += s.WallTrackCount
			d.Totals.Macro += s.MacroHits
			d.Totals.Blind += s.FlashKills
			d.Totals.Smoke += s.SmokeKills
			d.Totals.Incidents += s.Incidents
			d.Appearances = append(d.Appearances, Appearance{
				EntryID:    e.ID,
				Level:      s.Level,
				DemoName:   e.DemoName,
				MapName:    e.MapName,
				AnalyzedAt: e.AnalyzedAt,
				Name:       s.Name,
				Flags:      s.Flags,
				PerfRatio:  s.PerfRatio,
				TotalJumps: s.TotalJumps,
				ReactionMs: s.TriggerAvgMs,
				Snaps:      s.SnapCount,
				Wallhack:   s.WallTrackCount,
				Incidents:  s.Incidents,
			})
		}
	}
	if d.Demos == 0 {
		return nil
	}
	sort.Slice(d.Appearances, func(i, j int) bool { return d.Appearances[i].AnalyzedAt > d.Appearances[j].AnalyzedAt })
	d.Averages = acc.averages()
	d.Global = GetOverview()
	return d
}

func containsStr(list []string, s string) bool { return slices.Contains(list, s) }

func anyContains(list []string, q string) bool {
	for _, v := range list {
		if strings.Contains(strings.ToLower(v), q) {
			return true
		}
	}
	return false
}

// worthALook is true for cheat and suspect levels: what "flagged" means in the
// history. Conduct issues (AFK, teamflash) are shown but not counted as flagged.
func worthALook(level string) bool {
	return level == core.LevelCheat || level == core.LevelSuspect
}
