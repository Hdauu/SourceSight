package gui

import (
	"testing"
	"time"

	"sourcesight/pkg/core"
)

func TestPlaylistFor(t *testing.T) {
	rep := &core.DemoReport{
		Players: []core.PlayerStats{
			{Name: "a", SteamID: 1, Incidents: []core.TimelineIncident{{Tick: 900, Type: "PREFIRE"}, {Tick: 100, Type: "AUTOBUNNY"}}},
			{Name: "b", SteamID: 2},
		},
		Timeline: []core.TimelineIncident{{Tick: 100, Player: "a", SteamID: 1, Type: "AUTOBUNNY"}, {Tick: 500, Player: "c", SteamID: 3, Type: "MACRO"}, {Tick: 900, Player: "a", SteamID: 1, Type: "PREFIRE"}},
	}
	// One player's incidents, in tick order
	got := playlistFor(rep, 1)
	if len(got) != 2 || got[0].tick != 100 || got[1].tick != 900 || got[1].label != "PREFIRE de a" {
		t.Fatalf("player playlist = %+v", got)
	}
	// A player without incidents, or no player: the whole demo
	if got := playlistFor(rep, 2); len(got) != 3 || got[1].steamID != 3 {
		t.Fatalf("demo playlist = %+v", got)
	}
}

// Registering, switching and removing the hotkeys must not hang or leak the thread.
func TestHotkeysRegisterAndRelease(t *testing.T) {
	done := make(chan struct{})
	go func() {
		setHotkeys("ctrl-alt-pages", func(int) {})
		setHotkeys("ctrl-shift-arrows", func(int) {})
		setHotkeys("off", func(int) {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("setHotkeys hung")
	}
}
