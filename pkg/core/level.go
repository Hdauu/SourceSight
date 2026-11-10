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
	"ANTI-FLASH":      LevelSuspect,
	"SMOKE-VISION":    LevelSuspect,

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
