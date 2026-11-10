package modules

import (
	"fmt"
	"sourcesight/pkg/core"
)

func ProcessBhopTick(st *core.PlayerState, tick int, onGround, isJumping, onLadder, inWater bool, speed float64) {
	b := &st.Bhop

	if speed > b.MaxSpeed {
		b.MaxSpeed = speed
	}

	// Count atomic jump hold duration
	if isJumping {
		b.JumpHeldCount++
	} else if b.JumpHeldCount > 0 {
		b.TotalPresses++
		if b.JumpHeldCount == 1 {
			b.OneTickHops++
		}
		b.JumpHeldCount = 0
	}

	// Landing detection: air -> ground
	if !b.WasOnGround && onGround {
		b.LandTick = tick
	}

	// Check if player stayed on the ground too long (walking/stopping breaks streak)
	if onGround && (tick-b.LandTick) > core.BhopMaxGroundWindow {
		b.CurrentStreak = 0
	}

	// Takeoff detection: ground -> air
	if b.WasOnGround && !onGround {
		// Filter ladders and water
		if !onLadder && !inWater {
			ticksOnGround := tick - b.LandTick
			if isJumping && ticksOnGround <= core.BhopMaxGroundWindow {
				b.TotalJumps++
				if ticksOnGround <= core.BhopPerfWindow {
					b.PerfectHops++
					b.CurrentStreak++
					if b.CurrentStreak > b.MaxStreak {
						b.MaxStreak = b.CurrentStreak
					}

					// High-confidence incident: streak >= 5 (P < 0.2%)
					if b.CurrentStreak >= core.BhopCheatStreak {
						b.Incidents = append(b.Incidents, core.TimelineIncident{
							Tick: tick,
							Type: "AUTOBUNNY",
							Note: fmt.Sprintf("Inhuman streak of %d consecutive perfect hops (tick 0-1) — P < 0.05%%", b.CurrentStreak),
						})
					}
				} else {
					b.CurrentStreak = 0
				}
			} else {
				b.CurrentStreak = 0
			}
		} else {
			b.CurrentStreak = 0
		}
	}

	b.WasOnGround = onGround
}
