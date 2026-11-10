package modules

import (
	"fmt"
	"math"
	"sourcesight/pkg/core"
)

// ProcessWallTrackTick — Module W: Occluded ESP / Wallhack crosshair tracking detector.
// Called every game tick for every living player.
// Detects sustained crosshair lock onto enemies that are behind solid geometry.
func ProcessWallTrackTick(
	st *core.PlayerState,
	tick int,
	eyePos core.Vec3,
	pitch, yaw float32,
	trackerSpeed float64,
	enemies []EnemyHitbox,
	isVisible func(enemyName string) bool, // evaluated lazily: only for enemies inside the crosshair cone
) {
	if !st.IsAlive || len(enemies) == 0 {
		st.Wallhack.TargetPlayer = ""
		st.Wallhack.LockTicks = 0
		return
	}

	if st.Aim.PrevYaw == nil || st.Aim.PrevPitch == nil {
		st.Wallhack.TargetPlayer = ""
		st.Wallhack.LockTicks = 0
		return
	}

	dYaw := math.Abs(core.NormalizeAngle(float64(yaw) - float64(*st.Aim.PrevYaw)))
	dPitch := math.Abs(float64(pitch) - float64(*st.Aim.PrevPitch))
	isAimMoving := (dYaw + dPitch) > 0.08

	forward := core.EyeAnglesForward(pitch, yaw)

	var closestEnemy string
	minAngle := 999.0
	var enemyMoving bool

	for _, enemy := range enemies {
		// Skip both-still pairs — static crosshair placement is not suspicious
		if enemy.Speed < 30.0 && trackerSpeed < 30.0 {
			continue
		}
		targetVec := core.Vec3{
			X: enemy.Pos.X - eyePos.X,
			Y: enemy.Pos.Y - eyePos.Y,
			Z: (enemy.Pos.Z + 36.0) - eyePos.Z, // aim at chest center
		}
		dist := core.Length(targetVec)
		if dist < core.WallTrackMinDist || dist > core.WallTrackMaxDist {
			continue
		}
		normTarget := core.Normalize(targetVec)
		angle := core.AngleBetween(forward, normTarget)
		if angle < core.WallTrackMaxAngleDeg && angle < minAngle && !isVisible(enemy.Name) {
			minAngle = angle
			closestEnemy = enemy.Name
			enemyMoving = enemy.Speed > 30.0
		}
	}

	if closestEnemy != "" && (isAimMoving || enemyMoving) {
		if st.Wallhack.TargetPlayer == closestEnemy {
			st.Wallhack.LockTicks++
			sweep := math.Abs(core.NormalizeAngle(float64(yaw) - st.Wallhack.LockStartYaw))
			if !st.Wallhack.LockReported && st.Wallhack.LockTicks >= core.WallTrackMinTicks && sweep >= core.WallTrackMinSweepDeg {
				st.Wallhack.LockReported = true
				st.Wallhack.IncidentCount++
				st.Wallhack.Incidents = append(st.Wallhack.Incidents, core.TimelineIncident{
					Tick: tick,
					Type: "WALLTRACK",
					Note: fmt.Sprintf(
						"Crosshair tracking %s through wall for %d ticks (%.1f\u00b0 off, %.0f\u00b0 sweep) \u2014 Closet ESP",
						closestEnemy, st.Wallhack.LockTicks, minAngle, sweep,
					),
				})
			}
		} else {
			st.Wallhack.TargetPlayer = closestEnemy
			st.Wallhack.LockTicks = 1
			st.Wallhack.LockStartYaw = float64(yaw)
			st.Wallhack.LockReported = false
		}
	} else {
		st.Wallhack.TargetPlayer = ""
		st.Wallhack.LockTicks = 0
	}
}
