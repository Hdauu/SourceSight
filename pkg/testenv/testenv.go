// Package testenv isolates tests from the user's real SourceSight data.
package testenv

import (
	"fmt"
	"os"
	"testing"

	"sourcesight/pkg/appdata"
)

// Run points the data directory at a fresh temporary folder, refuses to run if it
// doesn't take effect (the user's history and reoffender database must never be
// touched by tests), runs the tests and cleans up. Call it from TestMain.
func Run(m *testing.M) {
	dir, err := os.MkdirTemp("", "sourcesight-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "testenv:", err)
		os.Exit(1)
	}
	os.Setenv(appdata.DataDirEnv, dir)
	if got := appdata.Dir(); got != dir {
		fmt.Fprintf(os.Stderr, "testenv: data directory is %s, not the temporary %s; refusing to run\n", got, dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
