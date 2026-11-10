package modules

import (
	"fmt"
	"sourcesight/pkg/core"
)

type SmokeVolume struct {
	Pos        core.Vec3
	StartTick  int
	ExpiryTick int
}

func CheckKillAnomalies(
	attacker *core.PlayerState,
	tick int,
	flashAlpha, flashDuration float32,
	attackerEye, victimPos core.Vec3,
	victimName string,
	activeSmokes []SmokeVolume,
	bulletCount int,
) {
	// Anti-Flash assist check
	if flashAlpha >= float32(core.FlashBlindAlpha) && flashDuration > float32(core.FlashDurationMin) {
		attacker.FlashSmoke.BlindKills = append(attacker.FlashSmoke.BlindKills, core.TimelineIncident{
			Tick: tick,
			Type: "BLIND_KILL",
			Note: fmt.Sprintf("Kill on %s while FULLY blinded (alpha=%.0f, duration=%.1fs) — ANTI-FLASH", victimName, flashAlpha, flashDuration),
		})
	}

	// Smoke penetration check
	fwd := core.Normalize(core.Vec3{
		X: victimPos.X - attackerEye.X,
		Y: victimPos.Y - attackerEye.Y,
		Z: victimPos.Z - attackerEye.Z,
	})

	throughSmoke := false
	for _, s := range activeSmokes {
		if core.RayIntersectsSphere(attackerEye, fwd, s.Pos, core.SmokeRadiusUnits) {
			throughSmoke = true
			break
		}
	}

	if throughSmoke && bulletCount <= 3 {
		attacker.FlashSmoke.SmokeKills = append(attacker.FlashSmoke.SmokeKills, core.TimelineIncident{
			Tick: tick,
			Type: "SMOKE_KILL",
			Note: fmt.Sprintf("%d-bullet kill through active smoke on %s (no prior LOS) — wallhack signature", bulletCount, victimName),
		})
	}
}
