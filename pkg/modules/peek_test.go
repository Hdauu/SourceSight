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

// A reaction measured during a ping spike is discarded, not counted as a prefire.
func TestPeekIgnoresPingSpike(t *testing.T) {
	st := core.NewPlayerState("p", 1)
	for range 5 {
		st.Ping.Sample(20)
	}
	st.Ping.Sample(200)
	UpdatePeekVisibility(st, 2, false, 100)
	UpdatePeekVisibility(st, 2, true, 101)
	RegisterPeekDamage(st, 2, "enemy", 102, 64) // 15 ms: would be a prefire
	if st.Peek.Engagements != 0 || st.Peek.Prefires != 0 || st.Peek.LagDiscarded != 1 {
		t.Fatalf("engagements=%d prefires=%d discarded=%d", st.Peek.Engagements, st.Peek.Prefires, st.Peek.LagDiscarded)
	}

	st.Ping.Sample(20) // spike over: measured again
	UpdatePeekVisibility(st, 2, false, 200)
	UpdatePeekVisibility(st, 2, true, 201)
	RegisterPeekDamage(st, 2, "enemy", 202, 64)
	if st.Peek.Prefires != 1 {
		t.Fatalf("prefires=%d after the spike", st.Peek.Prefires)
	}
}
