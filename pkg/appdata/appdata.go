// Package appdata resolves where SourceSight keeps its files: %APPDATA%\SourceSight.
// Older builds wrote next to the executable (or the working directory); those files
// are copied over the first time they are requested so nothing is lost on upgrade.
package appdata

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

// DataDirEnv names the environment variable that overrides the data directory.
const DataDirEnv = "SOURCESIGHT_DATA_DIR"

var (
	dirOnce sync.Once
	dir     string
)

// Dir returns the data directory, creating it if needed.
// SOURCESIGHT_DATA_DIR overrides the location (tests point it at a temporary folder).
func Dir() string {
	dirOnce.Do(func() {
		if custom := os.Getenv(DataDirEnv); custom != "" {
			dir = custom
		} else {
			base, err := os.UserConfigDir()
			if err != nil {
				base = exeDir()
			}
			dir = filepath.Join(base, "SourceSight")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			dir = exeDir()
		}
	})
	return dir
}

// Path returns the location of a data file, migrating a legacy copy if one exists.
func Path(name string) string {
	p := filepath.Join(Dir(), name)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	cwd, _ := os.Getwd()
	for _, legacyDir := range []string{exeDir(), cwd} {
		legacy := filepath.Join(legacyDir, name)
		if legacy != p && copyFile(legacy, p) == nil {
			break
		}
	}
	return p
}

// Sub returns a subdirectory of the data directory, creating it if needed.
func Sub(name string) string {
	p := filepath.Join(Dir(), name)
	_ = os.MkdirAll(p, 0755)
	return p
}

// ExportDir is where user-facing exports (PNG reports, evidence files) are written:
// Documents\SourceSight, falling back to the data directory.
func ExportDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "Documents", "SourceSight")
		if os.MkdirAll(p, 0755) == nil {
			return p
		}
	}
	return Dir()
}

func exeDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	return out.Close()
}
