package modules

import (
	"testing"

	"sourcesight/pkg/core"
)

// aimRun plays 400 ticks looking straight ahead, then one 90° turn on tick 400
// with the given lag conditions, and returns the state.
func aimRun(commands int, move float64, pingSpike bool) *core.PlayerState {
	st := core.NewPlayerState("p", 1)
	for range 5 {
		st.Ping.Sample(30)
	}
	pos := core.Vec3{}
	for tick := 0; tick < 400; tick++ {
		ProcessAimTick(st, tick, AimTick{Pitch: 10, Yaw: 0, Origin: pos, Commands: 1}, 128, true)
		pos.X += 1 // walking, about 128 u/s
	}
	if pingSpike {
		st.Ping.Sample(250)
	}
	pos.X += move
	ProcessAimTick(st, 400, AimTick{Pitch: 10, Yaw: 90, Origin: pos, Commands: commands}, 128, true)
	return st
}

func TestSnapCountsWhenReal(t *testing.T) {
	st := aimRun(1, 1, false)
	if len(st.Aim.SnapEvents) != 1 || st.Aim.LagSnaps != 0 {
		t.Fatalf("snaps=%d lag=%d, want a real snap", len(st.Aim.SnapEvents), st.Aim.LagSnaps)
	}
}

func TestSnapDuringLagIsNotCounted(t *testing.T) {
	cases := map[string]*core.PlayerState{
		"commands bunched": aimRun(5, 1, false),
		"warp":             aimRun(1, 40, false), // 40 units in one tick = 5120 u/s
		"ping spike":       aimRun(1, 1, true),
	}
	for name, st := range cases {
		if len(st.Aim.SnapEvents) != 0 || st.Aim.LagSnaps != 1 {
			t.Errorf("%s: snaps=%d lag=%d, want the snap discarded", name, len(st.Aim.SnapEvents), st.Aim.LagSnaps)
		}
	}
}
