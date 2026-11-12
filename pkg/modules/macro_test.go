package modules

import (
	"testing"

	"sourcesight/pkg/core"
)

// click presses attack at each tick (128-tick demo) with a semi-automatic pistol.
func click(st *core.PlayerState, ticks []int, semi bool) {
	for _, t := range ticks {
		ProcessMacroTick(st, t, true, "Desert Eagle", semi, 128)
	}
}

func evenly(start, every, n int) []int {
	var out []int
	for i := range n {
		out = append(out, start+i*every)
	}
	return out
}

func TestMacroFlagsPerfectlyEvenClicks(t *testing.T) {
	st := core.NewPlayerState("p", 1)
	click(st, evenly(1000, 16, 11), true) // 10 intervals of exactly 125 ms
	if len(st.Macro.Incidents) != 1 {
		t.Fatalf("incidents = %d, want 1", len(st.Macro.Incidents))
	}
}

func TestMacroIgnoresHumanClicks(t *testing.T) {
	st := core.NewPlayerState("p", 1)
	// A fast human: 110–160 ms between clicks, never twice the same pattern
	ticks := []int{1000, 1016, 1030, 1049, 1063, 1081, 1095, 1113, 1128, 1145, 1160, 1177}
	click(st, ticks, true)
	if len(st.Macro.Incidents) != 0 {
		t.Fatalf("human clicking flagged: %+v", st.Macro.Incidents)
	}
}

func TestMacroBurstsAreSeparate(t *testing.T) {
	// Five even clicks, a pause, five more: never 10 intervals in one burst
	st := core.NewPlayerState("p", 1)
	click(st, evenly(1000, 16, 6), true)
	click(st, evenly(5000, 16, 6), true)
	if len(st.Macro.Incidents) != 0 {
		t.Fatalf("pause between fights joined two bursts: %+v", st.Macro.Incidents)
	}

	// An automatic weapon in between also starts over
	st = core.NewPlayerState("p", 1)
	click(st, evenly(1000, 16, 6), true)
	click(st, []int{1100}, false)
	click(st, evenly(1116, 16, 6), true)
	if len(st.Macro.Incidents) != 0 {
		t.Fatalf("weapon switch didn't reset the burst: %+v", st.Macro.Incidents)
	}
}
