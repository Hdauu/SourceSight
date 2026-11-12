package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestParallelMatchesSequential parses real demos one by one and then all at once,
// and requires identical reports: any state shared between analyses would show up
// as a difference here (and as a data race under -race).
//
// Needs demos: SOURCESIGHT_TEST_DEMOS=D:\demos go test ./pkg/parser -run Parallel -v
func TestParallelMatchesSequential(t *testing.T) {
	dir := os.Getenv("SOURCESIGHT_TEST_DEMOS")
	if dir == "" {
		t.Skip("SOURCESIGHT_TEST_DEMOS not set")
	}
	demos, _ := filepath.Glob(filepath.Join(dir, "*.dem"))
	if len(demos) > 12 {
		demos = demos[:12]
	}
	if len(demos) < 2 {
		t.Skip("need at least 2 demos")
	}

	encode := func(path string) string {
		rep, err := ParseDemo(path, ParserOptions{})
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(path), err)
			return ""
		}
		rep.Header.Date = ""
		b, _ := json.Marshal(rep)
		return string(b)
	}

	start := time.Now()
	sequential := make([]string, len(demos))
	for i, d := range demos {
		sequential[i] = encode(d)
	}
	seqTime := time.Since(start)

	start = time.Now()
	parallel := make([]string, len(demos))
	var wg sync.WaitGroup
	for i, d := range demos {
		wg.Go(func() { parallel[i] = encode(d) })
	}
	wg.Wait()
	parTime := time.Since(start)

	for i := range demos {
		if sequential[i] != parallel[i] {
			t.Errorf("%s: parallel report differs from sequential", filepath.Base(demos[i]))
		}
	}
	t.Logf("%d demos: secuencial %.1fs, en paralelo %.1fs (%.1fx)", len(demos), seqTime.Seconds(), parTime.Seconds(), seqTime.Seconds()/parTime.Seconds())
}
