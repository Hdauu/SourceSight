package core

import (
	"slices"
	"testing"
)

func statsFor(mutate func(ps *PlayerState)) PlayerStats {
	ps := NewPlayerState("p", 1)
	mutate(ps)
	return ps.ComputeStats(128)
}

func snaps(n int) []TimelineIncident {
	out := make([]TimelineIncident, n)
	for i := range out {
		out[i] = TimelineIncident{Tick: 1000 + i*3} // all inside one second at 128 tick
	}
	return out
}

func TestDetectionRules(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(ps *PlayerState)
		flag   string
		level  string
	}{
		{"autobunny by ratio", func(ps *PlayerState) { ps.Bhop.TotalJumps, ps.Bhop.PerfectHops = 50, 45 }, "AUTOBUNNY", LevelCheat},
		{"streak alone is only suspicious (fofinha: 51%, streak 6)", func(ps *PlayerState) {
			ps.Bhop.TotalJumps, ps.Bhop.PerfectHops, ps.Bhop.MaxStreak = 80, 41, 6
		}, "SOSPECHOSO-BHOP", LevelSuspect},
		{"high ratio over few jumps is nothing (ZywOo: 60% in 15)", func(ps *PlayerState) { ps.Bhop.TotalJumps, ps.Bhop.PerfectHops = 15, 9 }, "", LevelClean},
		{"spinbot", func(ps *PlayerState) { ps.Aim.SnapEvents = snaps(200) }, "AIMLOCK", LevelCheat},
		{"few snaps in a burst is only suspicious (3ly0-: 17)", func(ps *PlayerState) { ps.Aim.SnapEvents = snaps(17) }, "SOSPECHOSO-AIM", LevelSuspect},
	}
	for _, c := range cases {
		s := statsFor(c.mutate)
		if c.flag == "" && len(s.Flags) > 0 || c.flag != "" && !slices.Contains(s.Flags, c.flag) {
			t.Errorf("%s: flags %v, want %q", c.name, s.Flags, c.flag)
		}
		if s.Level != c.level {
			t.Errorf("%s: level %s, want %s", c.name, s.Level, c.level)
		}
	}
}
