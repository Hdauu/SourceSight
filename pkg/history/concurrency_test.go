package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
)

func fakeReport(i int) *core.DemoReport {
	rep := &core.DemoReport{Header: core.DemoHeader{
		MapName:     "de_test",
		Tickrate:    128,
		TotalTicks:  100000 + i*1000,
		RoundStarts: []int{100, 2000, 4000},
	}}
	for p := 0; p < 10; p++ {
		ps := core.PlayerStats{SteamID: uint64(76561198000000000 + i*100 + p), Name: fmt.Sprintf("p%d_%d", i, p)}
		if p == 0 {
			ps.Flags = []string{"AIMLOCK"}
			ps.Level = core.LevelCheat
		}
		rep.Players = append(rep.Players, ps)
	}
	rep.Header.MatchID = rep.Fingerprint()
	return rep
}

// TestConcurrentSaveAndRead stores reports and records them in the reoffender
// database from many goroutines while others read. Meant for `go test -race`, which
// CI runs: the batch engine relies on these being safe to call concurrently.
func TestConcurrentSaveAndRead(t *testing.T) {
	dir := t.TempDir()
	const n = 16
	before, ovBefore := len(List()), GetOverview()
	var wg sync.WaitGroup
	for i := range n {
		path := filepath.Join(dir, fmt.Sprintf("demo%d.dem", i))
		if err := os.WriteFile(path, []byte(fmt.Sprint(i)), 0644); err != nil {
			t.Fatal(err)
		}
		wg.Go(func() {
			rep := fakeReport(i)
			if _, err := Save(path, rep, fmt.Sprint(i)); err != nil {
				t.Error(err)
			}
			db.GetDatabase().RecordReport(rep)
		})
		wg.Go(func() {
			_ = List()
			_ = GetOverview()
			_ = ListPlayers("p", "")
			_ = FindSimilar(fakeReport(i), "")
		})
	}
	wg.Wait()

	if got := len(List()); got != before+n {
		t.Fatalf("history has %d entries, want %d", got, before+n)
	}
	if ov := GetOverview(); ov.Players != ovBefore.Players+n*10 || ov.CheatPlayers != ovBefore.CheatPlayers+n {
		t.Fatalf("overview players=%d cheat=%d, want %d and %d", ov.Players, ov.CheatPlayers, ovBefore.Players+n*10, ovBefore.CheatPlayers+n)
	}
}
