package core

import (
	"encoding/json"
	"strings"
	"testing"
)

// SteamID64 values exceed 2^53, so they must reach the UI as strings or JavaScript
// silently rounds them (wrong Steam profile, wrong ban target).
func TestSteamIDSerializedAsString(t *testing.T) {
	b, _ := json.Marshal(PlayerStats{SteamID: 76561198000000123})
	if !strings.Contains(string(b), `"steamId":"76561198000000123"`) {
		t.Fatalf("SteamID not serialized as string: %s", b)
	}
	var back PlayerStats
	if err := json.Unmarshal(b, &back); err != nil || back.SteamID != 76561198000000123 {
		t.Fatalf("round trip failed: %v %d", err, back.SteamID)
	}
}

func TestSegmentIntersectsSphere(t *testing.T) {
	a, b := Vec3{X: 0}, Vec3{X: 1000}
	cases := []struct {
		name   string
		center Vec3
		want   bool
	}{
		{"between", Vec3{X: 500, Y: 100}, true},
		{"beside", Vec3{X: 500, Y: 200}, false},
		{"behind the target", Vec3{X: 1300}, false},
		{"behind the shooter", Vec3{X: -300}, false},
		{"touching the target", Vec3{X: 1100}, true},
	}
	for _, c := range cases {
		if got := SegmentIntersectsSphere(a, b, c.center, 144); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}
