package launcher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"sourcesight/pkg/appdata"
)

type AppConfig struct {
	CSGOPath       string  `json:"csgoPath"`               // Path to 'csgo' directory or game root
	PrerollSeconds float64 `json:"prerollSeconds"`         // Seconds to rewind before incident
	SidebarWidth   int     `json:"sidebarWidth,omitempty"` // px; 0 = default
}

var (
	cfgLock sync.Mutex
	cfgFile = appdata.Path("sourcesight_config.json")
)

func DefaultConfig() AppConfig {
	return AppConfig{
		CSGOPath:       "",
		PrerollSeconds: 2.5,
	}
}

func LoadConfig() AppConfig {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return DefaultConfig()
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	if cfg.PrerollSeconds <= 0 {
		cfg.PrerollSeconds = 2.5
	}
	return cfg
}

func SaveConfig(cfg AppConfig) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgFile, data, 0644)
}

// CleanCSGODir returns the path ending in the 'csgo' folder where demos live.
func CleanCSGODir(inputPath string) string {
	clean := filepath.Clean(inputPath)
	if stringsHasSuffixFold(clean, "csgo.exe") {
		clean = filepath.Dir(clean)
	}
	// If path doesn't end with "csgo", check if a subfolder "csgo" exists
	if !stringsHasSuffixFold(clean, "csgo") {
		sub := filepath.Join(clean, "csgo")
		if info, err := os.Stat(sub); err == nil && info.IsDir() {
			return sub
		}
	}
	return clean
}

func stringsHasSuffixFold(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	end := s[len(s)-len(suffix):]
	return filepath.Clean(end) == filepath.Clean(suffix) ||
		filepath.Base(s) == suffix
}
