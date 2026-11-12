package gui

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"sourcesight/pkg/history"
)

// fakePage stands in for the webview: it records every call made into the page and
// answers duplicate questions like a user would.
type fakePage struct {
	mu     sync.Mutex
	calls  []string
	answer string // option picked for every showPrompt
}

func (f *fakePage) Dispatch(fn func()) { fn() }

func (f *fakePage) Eval(js string) {
	f.mu.Lock()
	f.calls = append(f.calls, js)
	f.mu.Unlock()
	if strings.HasPrefix(js, "showPrompt(") {
		var args []json.RawMessage
		_ = json.Unmarshal([]byte("["+strings.TrimSuffix(strings.TrimPrefix(js, "showPrompt("), ");")+"]"), &args)
		var id string
		_ = json.Unmarshal(args[0], &id)
		go func() {
			for {
				if ch, ok := promptWaiters.Load(id); ok {
					ch.(chan string) <- f.answer
					return
				}
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
}

// last returns the arguments of the last call to fn, decoded into out.
func (f *fakePage) last(fn string, out ...any) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.calls) - 1; i >= 0; i-- {
		if strings.HasPrefix(f.calls[i], fn+"(") {
			var args []json.RawMessage
			body := strings.TrimSuffix(strings.TrimPrefix(f.calls[i], fn+"("), ");")
			if json.Unmarshal([]byte("["+body+"]"), &args) != nil {
				return false
			}
			for j := range out {
				_ = json.Unmarshal(args[j], out[j])
			}
			return true
		}
	}
	return false
}

func (f *fakePage) count(fn string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, fn+"(") {
			n++
		}
	}
	return n
}

func testDemos(t *testing.T, n int) []string {
	t.Helper()
	dir := os.Getenv("SOURCESIGHT_TEST_DEMOS")
	if dir == "" {
		t.Skip("SOURCESIGHT_TEST_DEMOS not set")
	}
	demos, _ := filepath.Glob(filepath.Join(dir, "*.dem"))
	if len(demos) < n {
		t.Skipf("need %d demos", n)
	}
	return demos[:n]
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
}

// TestBatchEndToEnd runs the real engine on real demos plus a renamed copy of one of
// them: every demo must be stored once and the copy must trigger the duplicate
// question (answered "skip").
//
// SOURCESIGHT_TEST_DEMOS=<carpeta-con-demos> go test ./pkg/gui -run Batch -v
func TestBatchEndToEnd(t *testing.T) {
	demos := testDemos(t, 4)
	dupe := filepath.Join(t.TempDir(), "copia renombrada.dem")
	copyFile(t, demos[0], dupe)

	page := &fakePage{answer: "skip"}
	a := &app{host: page}
	a.processQueue(append(append([]string{}, demos...), dupe), false)

	var rows []batchRow
	var info batchInfo
	if !page.last("batchDone", &rows, &info) {
		t.Fatalf("no batchDone; calls: %d", len(page.calls))
	}
	if page.count("batchStart") != 1 || page.count("batchUpdate") == 0 {
		t.Errorf("batchStart=%d batchUpdate=%d", page.count("batchStart"), page.count("batchUpdate"))
	}
	if page.count("showPrompt") != 1 {
		t.Errorf("want exactly one duplicate question, got %d", page.count("showPrompt"))
	}
	states := map[string]int{}
	for _, r := range rows {
		states[r.State]++
		if r.State == stDone && (r.HistoryID == "" || r.Level == "") {
			t.Errorf("%s: done without history id or level: %+v", r.Name, r)
		}
	}
	if states[stDone] != len(demos) || states[stSkipped] != 1 {
		t.Errorf("states = %v, want %d done and 1 skipped", states, len(demos))
	}
	if got := len(history.List()); got != len(demos) {
		t.Errorf("history has %d entries, want %d", got, len(demos))
	}
	t.Logf("%d demos + 1 copia en %.1fs con %d en paralelo", len(demos), info.Elapsed, info.Workers)
}

// TestBatchCancel stops a batch right after it starts: it must end promptly with
// every row in a final state.
func TestBatchCancel(t *testing.T) {
	demos := testDemos(t, 4)

	page := &fakePage{answer: "skip"}
	a := &app{host: page}
	done := make(chan struct{})
	go func() {
		a.processQueue(demos, true)
		close(done)
	}()
	for currentBatch.Load() == nil {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
	currentBatch.Load().stop()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("batch did not stop after cancel")
	}
	var rows []batchRow
	var info batchInfo
	page.last("batchDone", &rows, &info)
	if !info.Canceled {
		t.Error("batchDone not marked as canceled")
	}
	for _, r := range rows {
		switch r.State {
		case stCanceled, stDone, stCached:
		default:
			t.Errorf("%s left in state %q", r.Name, r.State)
		}
	}
}
