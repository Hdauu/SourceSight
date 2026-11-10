package modules

import (
	"testing"

	"sourcesight/pkg/core"
)

func TestPeekReaction(t *testing.T) {
	st := core.NewPlayerState("a", 1)
	const enemy = 2
	// Visible from the first observation: not a fresh peek, damage doesn't count
	UpdatePeekVisibility(st, enemy, true, 100)
	RegisterPeekDamage(st, enemy, "b", 105, 128)
	if st.Peek.Engagements != 0 {
		t.Fatalf("damage without a peek counted: %+v", st.Peek)
	}
	// Hidden, then peeks at tick 200; damage 8 ticks later = 62.5 ms -> prefire
	UpdatePeekVisibility(st, enemy, false, 150)
	UpdatePeekVisibility(st, enemy, true, 200)
	RegisterPeekDamage(st, enemy, "b", 208, 128)
	RegisterPeekDamage(st, enemy, "b", 210, 128) // second hit of the same sighting ignored
	// Next peek, damage after 40 ticks = 312 ms -> human reaction
	UpdatePeekVisibility(st, enemy, false, 300)
	UpdatePeekVisibility(st, enemy, true, 320)
	RegisterPeekDamage(st, enemy, "b", 360, 128)
	if st.Peek.Engagements != 2 || st.Peek.Prefires != 1 {
		t.Fatalf("want 2 engagements, 1 prefire; got %+v", st.Peek)
	}
}
