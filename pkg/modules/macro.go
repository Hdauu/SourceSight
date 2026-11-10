package modules

import (
	"fmt"
	"strings"
	"sourcesight/pkg/core"
)

var semiAutoWeapons = []string{
	"tec9", "elite", "deagle", "revolver",
	"p250", "cz75a", "fiveseven", "taser",
}

func ProcessMacroTick(st *core.PlayerState, tick int, attackRisingEdge bool, weaponClass string, tickrate int) {
	if !attackRisingEdge {
		return
	}

	mac := &st.Macro

	isSemi := false
	lowerWep := strings.ToLower(weaponClass)
	for _, w := range semiAutoWeapons {
		if strings.Contains(lowerWep, w) {
			isSemi = true
			break
		}
	}

	if !isSemi {
		t := tick
		mac.LastAttackTick = &t
		return
	}

	if mac.LastAttackTick != nil {
		deltaTicks := tick - *mac.LastAttackTick
		deltaMs := float64(deltaTicks) * (1000.0 / float64(tickrate))
		mac.AttackIntervals = append(mac.AttackIntervals, deltaMs)

		if len(mac.AttackIntervals) >= core.MacroMinSamples {
			sig := core.StdDev(mac.AttackIntervals)
			if sig < core.MacroSigmaThreshold {
				shouldPush := true
				if len(mac.Incidents) > 0 {
					last := mac.Incidents[len(mac.Incidents)-1]
					if tick-last.Tick < tickrate*5 {
						shouldPush = false
					}
				}

				if shouldPush {
					mac.Incidents = append(mac.Incidents, core.TimelineIncident{
						Tick: tick,
						Type: "MACRO",
						Note: fmt.Sprintf("Flat cadence σ=%.2fms across %d shots on %s — SOFTWARE MACRO", sig, len(mac.AttackIntervals), weaponClass),
					})
				}
			}
		}
	}

	t := tick
	mac.LastAttackTick = &t
}
