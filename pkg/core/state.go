package core

import (
	"fmt"
	"math"
	"strings"
)

type TimelineIncident struct {
	Tick    int    `json:"tick"`
	Type    string `json:"type"`
	Player  string `json:"player"`
	SteamID uint64 `json:"steamId,string"`
	GotoCmd string `json:"gotoCmd"`
	Note    string `json:"note"`
}

type ChatMessageEntry struct {
	Tick          int    `json:"tick"`
	Round         int    `json:"round"`
	SenderName    string `json:"senderName"`
	SenderSteamID uint64 `json:"senderSteamId,string"`
	Text          string `json:"text"`
	IsTeamOnly    bool   `json:"isTeamOnly"`
	IsToxicity    bool   `json:"isToxicity"`
	IsGhosting    bool   `json:"isGhosting"`
	Category      string `json:"category"`
	GotoCmd       string `json:"gotoCmd"`
}

type BhopData struct {
	TotalJumps    int
	PerfectHops   int
	TotalPresses  int
	OneTickHops   int
	WasOnGround   bool
	LandTick      int
	JumpHeldCount int
	MaxSpeed      float64
	CurrentStreak int
	MaxStreak     int
	Incidents     []TimelineIncident
}

type TriggerData struct {
	CrosshairOnEnemyTick *int
	ReactionDeltas       []float64
	Incidents            []TimelineIncident
	LagDiscarded         int // reactions ignored because they happened during a ping spike
}

type AimData struct {
	PrevPitch    *float32
	PrevYaw      *float32
	PrevOrigin   *Vec3 // position last tick, to spot warping
	AliveTicks   int
	SnapEvents   []TimelineIncident
	LagSnaps     int // snaps measured during a connection hiccup, not counted (see modules.ProcessAimTick)
	TickBase     int // last tickbase seen; -1 or 0 when unknown
	SilentEvents []TimelineIncident
}

type FlashSmokeData struct {
	BlindKills []TimelineIncident
	SmokeKills []TimelineIncident
}

type MacroData struct {
	LastAttackTick  *int
	AttackIntervals []float64
	Incidents       []TimelineIncident
}

type GhostData struct {
	DeathTick *int
	EnemyName string
	EnemyPos  *Vec3
	DeathPos  *Vec3
	Incidents []TimelineIncident
}

type GriefData struct {
	TeamflashSecs float64
	StartPos      *Vec3
	RoundDisplace []float64
}

type WallhackData struct {
	TargetPlayer  string
	LockTicks     int
	LockStartYaw  float64 // crosshair yaw when the current lock began
	LockReported  bool
	IncidentCount int
	Incidents     []TimelineIncident
}

// PeekData tracks how fast a player damages enemies right after they become visible
// to them. Hitting someone before a human could react suggests knowing they were there.
type PeekData struct {
	VisibleSince map[uint64]int  // enemy SteamID -> tick they became visible
	Visible      map[uint64]bool // enemy SteamID -> visible last tick (absent = not seen yet)
	Damaged      map[uint64]bool // already damaged during this sighting
	Engagements  int
	Prefires     int
	ReactionsMs  []float64
	Incidents    []TimelineIncident
	LagDiscarded int // reactions ignored because they happened during a ping spike
}

type PlayerState struct {
	Name          string
	SteamID       uint64
	Team          int
	IsAlive       bool
	Origin        Vec3
	Pitch         float32
	Yaw           float32
	FlashAlpha    float32
	FlashDuration float32
	PrevAttack    bool

	LastFirePitch float32
	LastFireYaw   float32
	LastFireEye   Vec3
	LastFireTick  int

	Bhop       BhopData
	Trigger    TriggerData
	Aim        AimData
	FlashSmoke FlashSmokeData
	Macro      MacroData
	Ghost      GhostData
	Grief      GriefData
	Wallhack   WallhackData
	Peek       PeekData
	Hits       HitProfile
	Ping       PingData
	Shift      ShiftData
	Path       PlayerPath // minimap, see mapdata.go
}

func NewPlayerState(name string, steamID uint64) *PlayerState {
	return &PlayerState{
		Name:    name,
		SteamID: steamID,
		Bhop: BhopData{
			WasOnGround: true,
		},
		Peek: PeekData{
			VisibleSince: map[uint64]int{},
			Visible:      map[uint64]bool{},
			Damaged:      map[uint64]bool{},
		},
	}
}

type PlayerStats struct {
	Name                string             `json:"name"`
	SteamID             uint64             `json:"steamId,string"`
	TotalJumps          int                `json:"totalJumps"`
	PerfectHops         int                `json:"perfectHops"`
	PerfRatio           float64            `json:"perfRatio"`
	TotalPresses        int                `json:"totalPresses"`
	OneTickHops         int                `json:"oneTickHops"`
	OneTickRatio        float64            `json:"oneTickRatio"`
	MaxStreak           int                `json:"maxStreak"`
	MaxSpeed            int                `json:"maxSpeed"`
	BhopVerdict         string             `json:"bhopVerdict"`
	TriggerAvgMs        *int               `json:"triggerAvgMs"`
	TriggerSigma        *float64           `json:"triggerSigma"`
	TriggerSamples      int                `json:"triggerSamples"`
	SnapCount           int                `json:"snapCount"`
	SilentCount         int                `json:"silentCount"`
	MacroHits           int                `json:"macroHits"`
	FlashKills          int                `json:"flashKills"`
	SmokeKills          int                `json:"smokeKills"`
	TeamflashSecs       float64            `json:"teamflashSecs"`
	GhostingCount       int                `json:"ghostingCount"`
	WallTrackCount      int                `json:"wallTrackCount"`
	PeekEngagements     int                `json:"peekEngagements"`
	Prefires            int                `json:"prefires"`
	PrefireRatio        float64            `json:"prefireRatio"`
	PeekReactionMs      *int               `json:"peekReactionMs"`         // median, from becoming visible to first damage
	LagDiscarded        int                `json:"lagDiscarded,omitempty"` // reactions ignored: measured during a ping spike
	LagSnaps            int                `json:"lagSnaps,omitempty"`     // snaps not counted: measured during a connection hiccup
	PingAvg             int                `json:"pingAvg,omitempty"`
	PingMax             int                `json:"pingMax,omitempty"`
	Hits                *HitProfile        `json:"hits,omitempty"`
	Notes               []string           `json:"notes,omitempty"`      // informative only, see notes.go
	ShiftCases          int                `json:"shiftCases,omitempty"` // SHIFT log entries (informative, see shift.go)
	PreviousInfractions int                `json:"previousInfractions"`
	PreviousNames       []string           `json:"previousNames"`
	ChatMessages        []ChatMessageEntry `json:"chatMessages"`
	IsAfk               bool               `json:"isAfk"`
	Flags               []string           `json:"flags"`
	Verdict             string             `json:"verdict"`
	Level               string             `json:"level"` // see level.go
	Incidents           []TimelineIncident `json:"incidents"`
}

func (ps *PlayerState) ComputeStats(tickrate int) PlayerStats {
	perfRatio := 0.0
	if ps.Bhop.TotalJumps > 0 {
		perfRatio = (float64(ps.Bhop.PerfectHops) / float64(ps.Bhop.TotalJumps)) * 100.0
	}
	oneTickRatio := 0.0
	if ps.Bhop.TotalPresses > 0 {
		oneTickRatio = (float64(ps.Bhop.OneTickHops) / float64(ps.Bhop.TotalPresses)) * 100.0
	}

	var flags []string

	// Cheat only on the ratio over enough jumps; a streak alone or a high ratio over few
	// jumps can happen to a good legit bhopper, so those are only suspicious.
	bhopCheat := perfRatio >= BhopCheatPerfRatio && ps.Bhop.TotalJumps >= BhopCheatMinJumps
	bhopSusp := !bhopCheat && ps.Bhop.TotalJumps >= BhopSuspMinJumps &&
		(perfRatio >= BhopSuspPerfRatio || ps.Bhop.MaxStreak >= BhopCheatStreak)

	bhopVerdict := "Limpio"
	if bhopCheat {
		bhopVerdict = "AUTOBUNNY"
		flags = append(flags, "AUTOBUNNY")
	} else if bhopSusp {
		bhopVerdict = "Sospechoso"
		flags = append(flags, "SOSPECHOSO-BHOP")
	}

	// Module B: Triggerbot
	var trigAvg *int
	var trigSig *float64
	trigSamples := len(ps.Trigger.ReactionDeltas)
	if trigSamples >= TriggerMinEvents {
		avg := Mean(ps.Trigger.ReactionDeltas)
		sig := StdDev(ps.Trigger.ReactionDeltas)
		roundedAvg := int(math.Round(avg))
		roundedSig := math.Round(sig*10) / 10
		trigAvg = &roundedAvg
		trigSig = &roundedSig

		if avg < TriggerBotMaxMs && sig < TriggerBotSigma {
			flags = append(flags, "TRIGGERBOT")
		}
	}

	// Module C: Aimlock & Silent Aim
	if len(ps.Aim.SilentEvents) >= 3 {
		flags = append(flags, "SILENT AIM")
	}
	// Spinbots produce hundreds of snaps; a handful in bursts is worth a look but can be
	// a brief toggle or odd flicks, so it is only suspicious.
	if len(ps.Aim.SnapEvents) >= AimMinSnapsToFlag && maxSnapBurst(ps.Aim.SnapEvents, tickrate) >= AimBurstMinSnaps {
		if len(ps.Aim.SnapEvents) >= AimCheatMinSnaps {
			flags = append(flags, "AIMLOCK")
		} else {
			flags = append(flags, "SOSPECHOSO-AIM")
		}
	}

	// Module E: Macro
	if len(ps.Macro.Incidents) >= 1 {
		flags = append(flags, "MACRO")
	}

	// Module D: Flash & Smoke (Must be statistically overwhelming proof to convict)
	if len(ps.FlashSmoke.BlindKills) >= 4 {
		flags = append(flags, "ANTI-FLASH")
	}
	if len(ps.FlashSmoke.SmokeKills) >= 5 {
		flags = append(flags, "SMOKE-VISION")
	}

	// Module G: Ghosting
	if len(ps.Ghost.Incidents) >= 2 {
		flags = append(flags, "GHOSTING")
	}

	// Prefire: damaging enemies faster than a human reacts once they become visible
	var peekMedian *int
	prefireRatio := 0.0
	if n := ps.Peek.Engagements; n > 0 {
		m := int(math.Round(Median(ps.Peek.ReactionsMs)))
		peekMedian = &m
		prefireRatio = 100 * float64(ps.Peek.Prefires) / float64(n)
	}
	prefireFlag := ps.Peek.Engagements >= PrefireMinEngagements &&
		prefireRatio >= PrefireMinRatio && *peekMedian <= int(PrefireMaxMedianMs)
	if prefireFlag {
		flags = append(flags, "PREFIRE")
	}

	// Module W: Wallhack / Occluded ESP Tracking
	if ps.Wallhack.IncidentCount >= 3 {
		flags = append(flags, "WALLHACK")
	}

	// Module I: Griefing & AFK
	afkCount := 0
	maxAfkStreak := 0
	for _, d := range ps.Grief.RoundDisplace {
		if d < AfkDisplacementUnits {
			afkCount++
			if afkCount > maxAfkStreak {
				maxAfkStreak = afkCount
			}
		} else {
			afkCount = 0
		}
	}
	isAfk := maxAfkStreak >= AfkRoundsThreshold
	if isAfk {
		flags = append(flags, "AFK/GRIEF")
	}
	if ps.Grief.TeamflashSecs >= 35.0 {
		flags = append(flags, "TEAMFLASH")
	}

	// Final verdict string
	verdict := "Limpio"
	if len(flags) > 0 {
		verdict = strings.Join(flags, " + ")
	} else if bhopSusp {
		verdict = "Sospechoso"
	}

	// Consolidate incidents: only include trigger incidents if player has triggerbot anomalies
	var incs []TimelineIncident
	incs = append(incs, ps.Bhop.Incidents...)
	if trigAvg != nil && (*trigAvg < TriggerBotMaxMs || (trigSig != nil && *trigSig < TriggerBotSigma && *trigAvg < 120)) {
		incs = append(incs, ps.Trigger.Incidents...)
	}
	incs = append(incs, ps.Aim.SnapEvents...)
	incs = append(incs, ps.Aim.SilentEvents...)
	incs = append(incs, ps.FlashSmoke.BlindKills...)
	incs = append(incs, ps.FlashSmoke.SmokeKills...)
	incs = append(incs, ps.Macro.Incidents...)
	incs = append(incs, ps.Ghost.Incidents...)
	incs = append(incs, ps.Wallhack.Incidents...)
	if prefireFlag {
		incs = append(incs, ps.Peek.Incidents...)
	}

	// Pre-roll calculation for timeline gotoCmd
	for i := range incs {
		prerollTick := incs[i].Tick - 192
		if prerollTick < 128 {
			prerollTick = 128
		}
		incs[i].GotoCmd = fmt.Sprintf("demo_gototick %d", prerollTick)
		incs[i].Player = ps.Name
		incs[i].SteamID = ps.SteamID
	}

	return PlayerStats{
		Name:            ps.Name,
		SteamID:         ps.SteamID,
		TotalJumps:      ps.Bhop.TotalJumps,
		PerfectHops:     ps.Bhop.PerfectHops,
		PerfRatio:       perfRatio,
		TotalPresses:    ps.Bhop.TotalPresses,
		OneTickHops:     ps.Bhop.OneTickHops,
		OneTickRatio:    oneTickRatio,
		MaxStreak:       ps.Bhop.MaxStreak,
		MaxSpeed:        int(math.Round(ps.Bhop.MaxSpeed)),
		BhopVerdict:     bhopVerdict,
		TriggerAvgMs:    trigAvg,
		TriggerSigma:    trigSig,
		TriggerSamples:  trigSamples,
		SnapCount:       len(ps.Aim.SnapEvents),
		SilentCount:     len(ps.Aim.SilentEvents),
		MacroHits:       len(ps.Macro.Incidents),
		FlashKills:      len(ps.FlashSmoke.BlindKills),
		SmokeKills:      len(ps.FlashSmoke.SmokeKills),
		TeamflashSecs:   math.Round(ps.Grief.TeamflashSecs*10) / 10,
		GhostingCount:   len(ps.Ghost.Incidents),
		WallTrackCount:  ps.Wallhack.IncidentCount,
		PeekEngagements: ps.Peek.Engagements,
		Prefires:        ps.Peek.Prefires,
		PrefireRatio:    math.Round(prefireRatio*10) / 10,
		PeekReactionMs:  peekMedian,
		LagDiscarded:    ps.Peek.LagDiscarded + ps.Trigger.LagDiscarded,
		LagSnaps:        ps.Aim.LagSnaps,
		PingAvg:         ps.Ping.Average(),
		PingMax:         ps.Ping.Max,
		Hits:            hitsOrNil(&ps.Hits),
		ShiftCases:      len(ps.Shift.Cases),
		IsAfk:           isAfk,
		Flags:           flags,
		Verdict:         verdict,
		Level:           LevelOf(flags),
		Incidents:       incs,
	}
}

type DemoHeader struct {
	MapName    string `json:"mapName"`
	ServerName string `json:"serverName"`
	TotalTicks int    `json:"totalTicks"`
	Tickrate   int    `json:"tickrate"`
	Date       string `json:"date"`
	DemoFile   string `json:"demoFile"`
	// RoundStarts holds the tick each round began, in order (round N starts at index N-1)
	RoundStarts []int `json:"roundStarts"`
	// MatchID fingerprints the match itself (see MatchFingerprint), so the same match
	// is recognized even when the file was renamed or copied.
	MatchID string `json:"matchId"`
}

type AdminEvent struct {
	Tick       int    `json:"tick"`
	Round      int    `json:"round"`
	AdminName  string `json:"adminName"`
	Command    string `json:"command"`
	TargetName string `json:"targetName"`
	IsRevenge  bool   `json:"isRevenge"`
	Note       string `json:"note"`
	GotoCmd    string `json:"gotoCmd"`
}

type DemoReport struct {
	Header   DemoHeader         `json:"header"`
	Players  []PlayerStats      `json:"players"`
	AdminLog []AdminEvent       `json:"adminLog"`
	ShiftLog []ShiftCase        `json:"shiftLog"`      // walking near unseen enemies, see shift.go (null: none; missing: older analysis)
	Map      *MapData           `json:"map,omitempty"` // positions for the minimap, see mapdata.go
	ChatLog  []ChatMessageEntry `json:"chatLog"`
	Timeline []TimelineIncident `json:"timeline"`
}

// maxSnapBurst returns the largest number of snaps inside any 1-second window.
// Snap events are appended in tick order, so a two-pointer sweep suffices.
func maxSnapBurst(snaps []TimelineIncident, tickrate int) int {
	best := 0
	lo := 0
	for hi := range snaps {
		for snaps[hi].Tick-snaps[lo].Tick > tickrate {
			lo++
		}
		if n := hi - lo + 1; n > best {
			best = n
		}
	}
	return best
}

// hitsOrNil keeps reports of players who never fired free of an empty hit profile.
func hitsOrNil(h *HitProfile) *HitProfile {
	if h.Empty() {
		return nil
	}
	c := *h
	return &c
}
