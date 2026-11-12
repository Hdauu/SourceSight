package modules

import (
	"fmt"
	"sourcesight/pkg/core"
)

type EnemyHitbox struct {
	Name  string
	Team  int
	Mins  core.Vec3
	Maxs  core.Vec3
	Pos   core.Vec3
	Speed float64
}

func ProcessTriggerTick(st *core.PlayerState, tick int, pitch, yaw float32, eyePos core.Vec3, enemies []EnemyHitbox, attackRisingEdge bool, tickrate int) {
	tr := &st.Trigger
	fwd := core.EyeAnglesForward(pitch, yaw)

	crosshairOnEnemy := false

	for _, enemy := range enemies {
		if core.RayIntersectsAABB(eyePos, fwd, enemy.Mins, enemy.Maxs) {
			crosshairOnEnemy = true

			if tr.CrosshairOnEnemyTick == nil {
				t := tick
				tr.CrosshairOnEnemyTick = &t
			} else if attackRisingEdge {
				deltaTicks := tick - *tr.CrosshairOnEnemyTick
				deltaMs := float64(deltaTicks) * (1000.0 / float64(tickrate))

				if deltaMs > 0 && deltaMs <= 1500.0 && st.Ping.Spiking() {
					// Measured during a ping spike: bunched commands distort the timing
					tr.LagDiscarded++
				} else if deltaMs > 0 && deltaMs <= 1500.0 {
					tr.ReactionDeltas = append(tr.ReactionDeltas, deltaMs)

					// Only record in the Overwatch Incident Timeline if reaction is suspiciously inhuman (< 80ms)
					if deltaMs < core.TriggerBotMaxMs {
						tr.Incidents = append(tr.Incidents, core.TimelineIncident{
							Tick: tick,
							Type: "TRIGGERBOT",
							Note: fmt.Sprintf("[%s] Sub-%.0fms reaction (%.1fms) — cheat territory", enemy.Name, core.TriggerBotMaxMs, deltaMs),
						})
					}
				}
				tr.CrosshairOnEnemyTick = nil
			}
			break
		}
	}

	if !crosshairOnEnemy {
		tr.CrosshairOnEnemyTick = nil
	}
}
