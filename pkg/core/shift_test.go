package core

import "testing"

const tr = 128

func unseenAt(name string, d float64) []NearEnemy {
	return []NearEnemy{{Name: name, Dist: d, Unseen: true}}
}

// Several presses while the same enemy stays close make one case.
func TestShiftGroupsPressesWhileEnemyIsClose(t *testing.T) {
	st := NewPlayerState("p", 1)
	tick := 1000
	step := func(held bool, near []NearEnemy) {
		ProcessShiftTick(st, tick, 3, held, near, tr)
		tick++
	}
	for range 3 { // three separate presses, enemy close the whole time
		for range 20 {
			step(true, unseenAt("e", 600))
		}
		for range 20 {
			step(false, unseenAt("e", 500))
		}
	}
	// Enemy leaves; the case closes after the gap
	for range 200 {
		step(false, nil)
	}
	cases := st.Shift.Cases
	if len(cases) != 1 {
		t.Fatalf("cases = %d, want 1", len(cases))
	}
	c := cases[0]
	if c.Presses != 3 || c.Round != 3 || c.Tick != 1000 || c.MinDist != 600 || len(c.Enemies) != 1 {
		t.Fatalf("case = %+v", c)
	}
	if c.HeldSecs != 0.5 { // 60 ticks at 128 tick, rounded to 0.1 s
		t.Fatalf("held = %v, want 0.5", c.HeldSecs)
	}
}

// Walking with no enemy close, or only with enemies the player can see, logs nothing.
func TestShiftNeedsAnUnseenEnemy(t *testing.T) {
	st := NewPlayerState("p", 1)
	for tick := 0; tick < 300; tick++ {
		near := []NearEnemy{{Name: "e", Dist: 400, Unseen: false}}
		if tick < 150 {
			near = nil
		}
		ProcessShiftTick(st, tick, 1, true, near, tr)
	}
	EndShift(st, tr)
	if n := len(st.Shift.Cases); n != 0 {
		t.Fatalf("cases = %d, want 0", n)
	}
}

// After the enemy has been gone for a while, a new press starts a new case.
func TestShiftNewCaseAfterEnemyLeaves(t *testing.T) {
	st := NewPlayerState("p", 1)
	tick := 0
	run := func(n int, held bool, near []NearEnemy) {
		for range n {
			ProcessShiftTick(st, tick, 1, held, near, tr)
			tick++
		}
	}
	run(30, true, unseenAt("a", 800))
	run(300, false, nil) // more than the 1 s gap
	run(30, true, unseenAt("b", 900))
	EndShift(st, tr) // death or round end closes the open one
	if n := len(st.Shift.Cases); n != 2 {
		t.Fatalf("cases = %d, want 2", n)
	}
	if st.Shift.Cases[1].Enemies[0] != "b" {
		t.Fatalf("second case = %+v", st.Shift.Cases[1])
	}
}
