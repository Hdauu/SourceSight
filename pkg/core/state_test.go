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
