package core

// Levels grade how serious a player's detections are. The app never says a player is
// a confirmed cheater: that is a staff decision recorded separately (db.Confirmation).
const (
	LevelClean   = "limpio"     // nothing worth a look
	LevelConduct = "conducta"   // not cheating, but staff-relevant (AFK, teamflash)
	LevelSuspect = "sospechoso" // weak indicators that need a look at the demo
	LevelCheat   = "cheat"      // crossed a strong detection threshold
)

var flagLevels = map[string]string{
	"AUTOBUNNY":  LevelCheat,
	"TRIGGERBOT": LevelCheat,
	"SILENT AIM": LevelCheat,
	"AIMLOCK":    LevelCheat,
	"WALLHACK":   LevelCheat,
	"MACRO":      LevelCheat,

	"SOSPECHOSO-BHOP": LevelSuspect,
	"SOSPECHOSO-AIM":  LevelSuspect,
	"ANTI-FLASH":      LevelSuspect,
	"SMOKE-VISION":    LevelSuspect,
	"PREFIRE":         LevelSuspect,

	"AFK/GRIEF": LevelConduct,
	"TEAMFLASH": LevelConduct,
	"GHOSTING":  LevelConduct,
}

var levelRank = map[string]int{LevelClean: 0, LevelConduct: 1, LevelSuspect: 2, LevelCheat: 3}

// FlagLevel returns the level of a single detection.
func FlagLevel(flag string) string {
	if l, ok := flagLevels[flag]; ok {
		return l
	}
	return LevelSuspect // unknown flags deserve a look
}

// LevelOf returns the most serious level among a player's detections.
func LevelOf(flags []string) string {
	level := LevelClean
	for _, f := range flags {
		level = MaxLevel(level, FlagLevel(f))
	}
	return level
}

// MaxLevel returns the more serious of two levels.
func MaxLevel(a, b string) string {
	if levelRank[b] > levelRank[a] {
		return b
	}
	return a
}

// FillLevels sets Level on every player of a report; reports cached by older
// versions don't carry it.
func (r *DemoReport) FillLevels() {
	for i := range r.Players {
		r.Players[i].Level = LevelOf(r.Players[i].Flags)
	}
}

// KnownFlags lists every detection the analyzer can emit.
func KnownFlags() []string {
	out := make([]string, 0, len(flagLevels))
	for f := range flagLevels {
		out = append(out, f)
	}
	return out
}

// AnalysisVersion changes whenever detectors change in a way that makes older
// reports incomplete. Stored with each history entry, so the app can offer to
// re-analyze demos analyzed by an older version.
//
//	1: up to v2.2.0
//	2: prefire detection (v2.3.0)
//	3: stricter cheat level for bhop streaks and few aim snaps (v2.3.0)
//	4: hits per body zone and weapon class, ping, reactions during ping spikes ignored;
//	   macro measured per burst on every semi-auto pistol; smoke kills only through a smoke
//	   that exists (18 s, same round) and lies between killer and victim, bullets per
//	   round, wallhack and revenge windows scaled to the demo tickrate
//	5: SHIFT log (walking near unseen enemies); chat spam window scaled to the tickrate
//	6: positions for the minimap (kills and once-per-second paths)
//	7: aim snaps during connection hiccups (bunched commands, warps, ping spikes) not counted;
//	   every player who played is listed
const AnalysisVersion = 7
