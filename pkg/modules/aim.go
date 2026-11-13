package modules

import (
	"fmt"
	"math"
	"sourcesight/pkg/core"
)

// AimTick is what the aim module needs about a player on one tick.
type AimTick struct {
	Pitch, Yaw float32
	Origin     core.Vec3
	// Commands is how many usercmds the server ran for the player this tick (tickbase
	// delta); -1 when unknown. More than one means commands arrived bunched after a
	// connection hiccup, and the view change is the sum of several of them.
	Commands int
}

// ProcessAimTick looks for snaps: view turns of more than core.AimSnapThresholdDeg in
// one tick. A turn measured while the connection hiccuped (commands bunched, the
// player warping, or a ping spike) isn't a real single-tick turn: it's counted in
// LagSnaps instead of SnapEvents.
func ProcessAimTick(st *core.PlayerState, tick int, t AimTick, tickrate int, isAlive bool) {
	aim := &st.Aim
	dtMs := 1000.0 / float64(tickrate)
	pitch, yaw := t.Pitch, t.Yaw

	if !isAlive {
		aim.PrevPitch = nil
		aim.PrevYaw = nil
		aim.PrevOrigin = nil
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

			if lagReason := snapLag(st, t, tickrate); lagReason != "" {
				aim.LagSnaps++
			} else {
				aim.SnapEvents = append(aim.SnapEvents, core.TimelineIncident{
					Tick: tick,
					Type: "AIMSNAP",
					Note: fmt.Sprintf("Angular snap: %.1f° in one tick (%.2fms) — exceeds physical mouse limit", delta, dtMs),
				})
			}
		}
	}

	p := pitch
	y := yaw
	o := t.Origin
	aim.PrevPitch = &p
	aim.PrevYaw = &y
	aim.PrevOrigin = &o
}

// snapLag says why a snap on this tick can't be trusted, or "" if it can.
func snapLag(st *core.PlayerState, t AimTick, tickrate int) string {
	if t.Commands > 1 {
		return "comandos amontonados"
	}
	if prev := st.Aim.PrevOrigin; prev != nil {
		// Moving faster than anyone can run, even bunny hopping: the player warped
		if core.Dist2D(*prev, t.Origin)*float64(tickrate) > core.AimWarpSpeed {
			return "warpeo"
		}
	}
	if st.Ping.Spiking() {
		return "pico de ping"
	}
	return ""
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
