package update

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v2.1.0", "v2.0.9", true},
		{"v2.0.10", "v2.0.9", true},
		{"v2.0.0", "v2.0.0", false},
		{"v1.9.9", "v2.0.0", false},
		{"v3.0.0-beta", "v2.5.1", true},
		{"2.1", "v2.0.3", true},
	}
	for _, c := range cases {
		if got := newer(c.latest, c.current); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}
