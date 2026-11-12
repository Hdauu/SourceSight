package history

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"sourcesight/pkg/core"
)

// The player page compares a player against everyone else: their own hits must not
// be part of the "others" averages (they are part of the global ones).
func TestPlayerOthersExcludesThePlayer(t *testing.T) {
	rep := fakeReport(900)
	for i := range rep.Players {
		head := 2
		if i == 0 {
			head = 50 // the player we look at: everything to the head
		}
		rep.Players[i].Hits = &core.HitProfile{Zones: core.BodyZones{Head: core.ZoneCount{Hits: head}, Chest: core.ZoneCount{Hits: 10}}}
	}
	path := filepath.Join(t.TempDir(), "others.dem")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Save(path, rep, "others"); err != nil {
		t.Fatal(err)
	}
	d := GetPlayer(strconv.FormatUint(rep.Players[0].SteamID, 10))
	if d == nil || d.Others.Hits == nil || d.Global.Averages.Hits == nil {
		t.Fatalf("detail = %+v", d)
	}
	if got := d.Others.Hits.Zones.Head.Hits; got != 2*(len(rep.Players)-1) {
		t.Fatalf("others head hits = %d, want %d (without the player)", got, 2*(len(rep.Players)-1))
	}
	if d.Global.Averages.Hits.Zones.Head.Hits < 50 {
		t.Fatal("global averages should include the player")
	}
}
