package modules

import (
	"fmt"
	"math"
	"sourcesight/pkg/core"
)

func ProcessAimTick(st *core.PlayerState, tick int, pitch, yaw float32, tickrate int, isAlive bool) {
	aim := &st.Aim
	dtMs := 1000.0 / float64(tickrate)

	if !isAlive {
		aim.PrevPitch = nil
		aim.PrevYaw = nil
		aim.AliveTicks = 0
		return
	}

	aim.AliveTicks++

	if aim.PrevPitch != nil && aim.PrevYaw != nil {
		rawDeltaYaw := float64(yaw - *aim.PrevYaw)
		rawDeltaPitch := float64(pitch - *aim.PrevPitch)

		normYaw := core.NormalizeAngle(rawDeltaYaw)
		delta := math.Sqrt(normYaw*normYaw + rawDeltaPitch*rawDeltaPitch)

		// Warmup is defined in 64-tick ticks; pitch exactly 0 is a server-forced view angle (spawn/teleport)
		warmup := core.AimAliveWarmup * tickrate / 64
		if aim.AliveTicks > warmup && pitch != 0 &&
			delta > core.AimSnapThresholdDeg &&
			delta < core.AimWrapThreshold {

			aim.SnapEvents = append(aim.SnapEvents, core.TimelineIncident{
				Tick: tick,
				Type: "AIMSNAP",
				Note: fmt.Sprintf("Angular snap: %.1f° in one tick (%.2fms) — exceeds physical mouse limit", delta, dtMs),
			})
		}
	}

	p := pitch
	y := yaw
	aim.PrevPitch = &p
	aim.PrevYaw = &y
}

func CheckSilentAim(st *core.PlayerState, tick int, victimPos core.Vec3, victimName string, dmg int) {
	if dmg <= 0 || st.LastFireTick == 0 || (tick-st.LastFireTick) > 16 {
		return
	}

	serverVec := core.Normalize(core.Vec3{
		X: victimPos.X - st.LastFireEye.X,
		Y: victimPos.Y - st.LastFireEye.Y,
		Z: victimPos.Z - st.LastFireEye.Z,
	})

	clientVec := core.EyeAnglesForward(st.LastFirePitch, st.LastFireYaw)
	divDeg := core.AngleBetween(clientVec, serverVec)

	// In Source Engine, jump/running spray inaccuracy with SMGs/shotguns can be up to 35 degrees.
	// True Silent Aim (angle packet divergence) typically exceeds 35-40 degrees.
	if divDeg > 35.0 {
		st.Aim.SilentEvents = append(st.Aim.SilentEvents, core.TimelineIncident{
			Tick: tick,
			Type: "SILENT_AIM",
			Note: fmt.Sprintf("Client view %.1f° divergence from registered hit on %s (dmg: %d) — SILENT AIM", divDeg, victimName, dmg),
		})
	}
}
