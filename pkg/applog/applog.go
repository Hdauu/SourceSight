// Package applog appends diagnostic lines to %APPDATA%\SourceSight\sourcesight.log
// so problems on a user's machine can be inspected after the fact.
package applog

import (
	"fmt"
	"os"
	"sync"
	"time"

	"sourcesight/pkg/appdata"
)

const maxSize = 1 << 20 // rotate at 1 MB

var mu sync.Mutex

// Path is the log file location.
func Path() string { return appdata.Path("sourcesight.log") }

// Printf writes one timestamped line.
func Printf(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	p := Path()
	if st, err := os.Stat(p); err == nil && st.Size() > maxSize {
		_ = os.Rename(p, p+".1")
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), fmt.Sprintf(format, args...))
}
