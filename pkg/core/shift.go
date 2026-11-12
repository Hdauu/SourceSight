package core

import (
	"slices"
	"strconv"
)

// SHIFT log: every time a player walks silently (holding +speed, shift by default)
// while an enemy they can't see is close. Walking is how you sneak up on someone, so
// doing it next to an enemy you supposedly don't know is there is worth a look; it is
// also completely normal play, so this log never changes a player's level or flags.
// It is a list for staff to review, with no criteria beyond distance and visibility.
//
// One case groups everything that happens while the same enemies stay close: every
// shift press is counted in it instead of producing a case of its own.

// IN_SPEED is the +speed button (walk, shift by default) in the usercmd buttons.
const IN_SPEED = 1 << 17

const (
	// ShiftRadiusDefault: an enemy within this many units counts as close (about the
	// distance at which running footsteps are heard). Configurable in the app.
	ShiftRadiusDefault = 1000.0
	// ShiftGapTicks: a case closes once no enemy has been close for this long, in
	// 64-tick ticks (1 s; scaled to the demo), so brief gaps don't split it.
	ShiftGapTicks = 64
)

// ShiftCase is one entry of the log.
type ShiftCase struct {
	Tick     int      `json:"tick"`    // first press with an unseen enemy close
	EndTick  int      `json:"endTick"` // last tick an enemy was close
	Round    int      `json:"round"`
	Player   string   `json:"player"`
	SteamID  uint64   `json:"steamId,string"`
	Enemies  []string `json:"enemies"`  // unseen enemies close while walking
	Presses  int      `json:"presses"`  // shift presses during the case
	HeldSecs float64  `json:"heldSecs"` // time walking during the case
	MinDist  int      `json:"minDist"`  // closest an unseen enemy got, in units
	GotoCmd  string   `json:"gotoCmd"`
}

// ShiftData is the per-player state of the log while a demo is analyzed.
type ShiftData struct {
	prevHeld    bool
	active      bool
	cur         ShiftCase
	heldTicks   int
	lastNear    int
	caseEnemies map[string]bool
	Cases       []ShiftCase
}

// NearEnemy is an enemy within the radius this tick.
type NearEnemy struct {
	Name   string
	Dist   float64
	Unseen bool // the player can't see them
}

// ProcessShiftTick updates the log for one living player. held: shift is down this
// tick; near: enemies within the radius (seen or not).
func ProcessShiftTick(st *PlayerState, tick, round int, held bool, near []NearEnemy, tickrate int) {
	sd := &st.Shift
	pressed := held && !sd.prevHeld
	sd.prevHeld = held

	unseen := 0
	for _, e := range near {
		if e.Unseen {
			unseen++
		}
	}

	if !sd.active {
		if !held || unseen == 0 {
			return
		}
		// A new case: walking with an unseen enemy close
		sd.active = true
		sd.cur = ShiftCase{Tick: tick, Round: round, Player: st.Name, SteamID: st.SteamID, Presses: 1, MinDist: 1 << 30}
		sd.heldTicks = 0
		sd.lastNear = tick
		sd.caseEnemies = map[string]bool{}
		pressed = false // already counted
	}

	// Still close: any enemy of the case, or any unseen enemy
	stillNear := unseen > 0
	for _, e := range near {
		if sd.caseEnemies[e.Name] {
			stillNear = true
		}
	}
	if stillNear {
		sd.lastNear = tick
	} else if tick-sd.lastNear > ShiftGapTicks*tickrate/64 {
		EndShift(st, tickrate)
		return
	}

	if held {
		sd.heldTicks++
		for _, e := range near {
			if !e.Unseen {
				continue
			}
			if !sd.caseEnemies[e.Name] {
				sd.caseEnemies[e.Name] = true
				sd.cur.Enemies = append(sd.cur.Enemies, e.Name)
			}
			sd.cur.MinDist = min(sd.cur.MinDist, int(e.Dist+0.5))
		}
	}
	if pressed && unseen > 0 {
		sd.cur.Presses++
	}
}

// EndShift closes the open case, if any: when the enemies are gone, the player dies,
// the round ends or the demo ends.
func EndShift(st *PlayerState, tickrate int) {
	sd := &st.Shift
	sd.prevHeld = false
	if !sd.active {
		return
	}
	sd.active = false
	c := sd.cur
	c.EndTick = sd.lastNear
	// The game lists players in random order: sort so the same demo always gives the same log
	slices.Sort(c.Enemies)
	c.HeldSecs = float64(int(float64(sd.heldTicks)/float64(tickrate)*10+0.5)) / 10
	c.GotoCmd = gotoCmd(c.Tick)
	sd.Cases = append(sd.Cases, c)
}

// gotoCmd is the demo_gototick command with the same lead-in as other incidents.
func gotoCmd(tick int) string {
	return "demo_gototick " + strconv.Itoa(max(128, tick-192))
}

// Active reports whether a case is open.
func (sd *ShiftData) Active() bool { return sd.active }
