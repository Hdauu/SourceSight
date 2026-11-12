package core

import (
	"strings"
	"testing"
)

// typical is a normal player's demo: ~17% head, ~25% accuracy, 250 ms reaction.
func typical() NoteCounts {
	return NoteCounts{Hits: 60, HeadHits: 10, Shots: 240, Kills: 15, HSKills: 6,
		ReactSumMs: 250 * 12, ReactN: 12, Engagements: 30, Prefires: 2, PingSum: 30, PingN: 1, PingMax: 40}
}

func base() NoteCounts {
	var b NoteCounts
	for range 9 {
		b.Add(typical())
	}
	return b
}

func TestNotesQuietForNormalPlayer(t *testing.T) {
	if n := Notes(typical(), base(), "el resto"); len(n) != 0 {
		t.Fatalf("normal player got notes: %q", n)
	}
}

func TestNotesFlagOutliers(t *testing.T) {
	me := typical()
	me.HeadHits = 30                     // 50% head
	me.Shots = 100                       // 60% accuracy
	me.HSKills = 14                      // 14/15 kills HS
	me.ReactSumMs, me.ReactN = 90*12, 12 // 90 ms
	me.Prefires = 8                      // 27%: below the flag, well above the rest
	me.PingSum = 150                     // high ping
	me.LagDiscarded = 2
	notes := Notes(me, base(), "el resto")
	want := []string{"a la cabeza", "Acertó", "headshot", "en la mira", "dañó al rival", "ping alto", "Se descartaron 2"}
	if len(notes) != len(want) {
		t.Fatalf("got %d notes: %q", len(notes), notes)
	}
	for i, w := range want {
		if !strings.Contains(notes[i], w) {
			t.Errorf("note %d = %q, want it to mention %q", i, notes[i], w)
		}
	}
}

func TestNotesNeedEnoughData(t *testing.T) {
	// Every bullet to the head, but only 5 hits: says nothing
	me := NoteCounts{Hits: 5, HeadHits: 5, Shots: 10, Kills: 3, HSKills: 3}
	if n := Notes(me, base(), "el resto"); len(n) != 0 {
		t.Fatalf("too little data still produced notes: %q", n)
	}
	// Enough from the player, but no baseline to compare with
	if n := Notes(typical(), NoteCounts{}, "el resto"); len(n) != 0 {
		t.Fatalf("no baseline still produced notes: %q", n)
	}
}

func TestMatchNotesExcludesThePlayer(t *testing.T) {
	players := make([]PlayerStats, 10)
	for i := range players {
		players[i].Hits = &HitProfile{Shots: 240, Zones: BodyZones{Head: ZoneCount{Hits: 10}, Chest: ZoneCount{Hits: 50}}}
	}
	players[0].Hits.Zones.Head.Hits = 40 // 40 of 90 to the head
	MatchNotes(players)
	if len(players[0].Notes) == 0 || !strings.Contains(players[0].Notes[0], "cabeza") {
		t.Fatalf("outlier notes = %q", players[0].Notes)
	}
	for _, p := range players[1:] {
		if len(p.Notes) != 0 {
			t.Fatalf("normal player got %q", p.Notes)
		}
	}
}
