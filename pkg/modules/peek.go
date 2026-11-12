package modules

import (
	"fmt"

	"sourcesight/pkg/core"
)

// UpdatePeekVisibility records whether an enemy is visible to the player this tick.
// A sighting starts when an enemy that was hidden on the previous tick becomes
// visible; maps are only written on those transitions, since this runs for every
// pair of players on every tick.
func UpdatePeekVisibility(st *core.PlayerState, enemyID uint64, visible bool, tick int) {
	pk := &st.Peek
	prev, known := pk.Visible[enemyID]
	if known && prev == visible {
		return
	}
	pk.Visible[enemyID] = visible
	if !visible {
		delete(pk.VisibleSince, enemyID)
		return
	}
	if known { // hidden -> visible: a fresh peek
		pk.VisibleSince[enemyID] = tick
		pk.Damaged[enemyID] = false
	}
}

// RegisterPeekDamage measures the reaction for the first bullet damage of a sighting.
func RegisterPeekDamage(st *core.PlayerState, enemyID uint64, enemyName string, tick, tickrate int) {
	pk := &st.Peek
	since, ok := pk.VisibleSince[enemyID]
	if !ok || pk.Damaged[enemyID] {
		return
	}
	pk.Damaged[enemyID] = true
	// A ping spike bunches commands together: the reaction would look faster than it was
	if st.Ping.Spiking() {
		pk.LagDiscarded++
		return
	}
	ms := float64(tick-since) * 1000 / float64(tickrate)
	if ms > core.PrefireMaxSightMs {
		return
	}
	pk.Engagements++
	pk.ReactionsMs = append(pk.ReactionsMs, ms)
	if ms < core.PrefireWindowMs {
		pk.Prefires++
		pk.Incidents = append(pk.Incidents, core.TimelineIncident{
			Tick: tick,
			Type: "PREFIRE",
			Note: fmt.Sprintf("Damaged %s %.0f ms after they became visible — faster than human reaction", enemyName, ms),
		})
	}
}

// ResetPeekRound clears sightings between rounds (positions reset on spawn).
func ResetPeekRound(st *core.PlayerState) {
	clear(st.Peek.VisibleSince)
	clear(st.Peek.Visible)
	clear(st.Peek.Damaged)
}
