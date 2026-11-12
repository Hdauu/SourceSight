package modules

import (
	"testing"

	"sourcesight/pkg/core"
)

func TestSmokeKillNeedsALiveSmokeInBetween(t *testing.T) {
	eye, victim := core.Vec3{X: 0, Z: 64}, core.Vec3{X: 1000, Z: 36}
	smokeAt := func(x float64, start int) []SmokeVolume {
		return []SmokeVolume{{Pos: core.Vec3{X: x}, StartTick: start, ExpiryTick: start + 18*128}}
	}
	cases := []struct {
		name   string
		smokes []SmokeVolume
		tick   int
		want   int
	}{
		{"live smoke between them", smokeAt(500, 1000), 2000, 1},
		{"smoke already gone (thrown 60 s earlier)", smokeAt(500, 1000), 1000 + 60*128, 0},
		{"smoke not thrown yet", smokeAt(500, 5000), 2000, 0},
		{"smoke behind the victim", smokeAt(1400, 1000), 2000, 0},
	}
	for _, c := range cases {
		st := core.NewPlayerState("p", 1)
		CheckKillAnomalies(st, c.tick, 0, 0, eye, victim, "v", c.smokes, 1)
		if got := len(st.FlashSmoke.SmokeKills); got != c.want {
			t.Errorf("%s: %d smoke kills, want %d", c.name, got, c.want)
		}
	}
}
