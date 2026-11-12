package modules

import (
	"fmt"

	"sourcesight/pkg/core"
)

// macroMaxGapMs: clicks further apart than this are separate bursts, not one spray of
// clicks; a pause between fights must not mix into the rhythm being measured.
const macroMaxGapMs = 500.0

// ProcessMacroTick measures the rhythm of attack presses (rising edges of IN_ATTACK).
// Only the last core.MacroMinSamples intervals of the current burst count: a burst
// ends after a pause longer than macroMaxGapMs or when the player switches to a weapon
// that isn't a semi-automatic pistol (semiAuto, decided by the caller from the type).
func ProcessMacroTick(st *core.PlayerState, tick int, attackRisingEdge bool, weaponName string, semiAuto bool, tickrate int) {
	if !attackRisingEdge {
		return
	}
	mac := &st.Macro
	t := tick
	if !semiAuto {
		mac.LastAttackTick = nil
		mac.AttackIntervals = mac.AttackIntervals[:0]
		return
	}
	if mac.LastAttackTick != nil {
		deltaMs := float64(tick-*mac.LastAttackTick) * 1000 / float64(tickrate)
		if deltaMs > macroMaxGapMs {
			mac.AttackIntervals = mac.AttackIntervals[:0] // new burst
		} else {
			mac.AttackIntervals = append(mac.AttackIntervals, deltaMs)
			if n := len(mac.AttackIntervals); n > core.MacroMinSamples {
				mac.AttackIntervals = mac.AttackIntervals[n-core.MacroMinSamples:]
			}
		}
	}
	mac.LastAttackTick = &t

	if len(mac.AttackIntervals) < core.MacroMinSamples {
		return
	}
	sig := core.StdDev(mac.AttackIntervals)
	if sig >= core.MacroSigmaThreshold {
		return
	}
	// One incident per burst: wait 5 s before reporting the same player again
	if n := len(mac.Incidents); n > 0 && tick-mac.Incidents[n-1].Tick < tickrate*5 {
		return
	}
	mac.Incidents = append(mac.Incidents, core.TimelineIncident{
		Tick: tick,
		Type: "MACRO",
		Note: fmt.Sprintf("Flat cadence σ=%.2fms across %d clicks on %s (avg %.0f ms apart) — SOFTWARE MACRO",
			sig, len(mac.AttackIntervals), weaponName, core.Mean(mac.AttackIntervals)),
	})
}
