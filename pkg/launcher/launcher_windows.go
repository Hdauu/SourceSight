package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// FindSteamPath returns the root directory of the Steam installation.
func FindSteamPath() (string, error) {
	// Try HKCU
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		p, _, err := k.GetStringValue("SteamPath")
		if err == nil && p != "" {
			return filepath.Clean(p), nil
		}
	}

	// Try HKLM 64-bit
	k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		p, _, err := k.GetStringValue("InstallPath")
		if err == nil && p != "" {
			return filepath.Clean(p), nil
		}
	}

	// Fallback standard locations
	candidates := []string{
		`C:\Program Files (x86)\Steam`,
		`C:\Program Files\Steam`,
		`D:\Steam`,
		`E:\Steam`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("no se encontró la instalación de Steam en el registro")
}

// FindCSGODirectory locates the 'csgo' folder where demos must reside for +playdemo.
func FindCSGODirectory() (string, error) {
	// 1. Check user configured path first
	cfg := LoadConfig()
	if cfg.CSGOPath != "" {
		cleaned := CleanCSGODir(cfg.CSGOPath)
		if info, err := os.Stat(cleaned); err == nil && info.IsDir() {
			return cleaned, nil
		}
	}

	steamPath, err := FindSteamPath()
	if err != nil {
		return "", err
	}

	// Candidates list: Valve legacy folder, standard CSGO, and CS2
	subnames := []string{
		filepath.Join("common", "csgo legacy", "csgo"),
		filepath.Join("common", "Counter-Strike Global Offensive", "csgo"),
		filepath.Join("common", "Counter-Strike Global Offensive", "game", "csgo"),
		filepath.Join("common", "Counter-Strike 2", "game", "csgo"),
	}

	// 2. Default main library
	for _, sub := range subnames {
		candidate := filepath.Join(steamPath, "steamapps", sub)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			cfg.CSGOPath = candidate
			_ = SaveConfig(cfg)
			return candidate, nil
		}
	}

	// 3. Parse libraryfolders.vdf
	vdfPath := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")
	if f, err := os.Open(vdfPath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, `"path"`) {
				parts := strings.Split(line, `"path"`)
				if len(parts) > 1 {
					libPath := strings.Trim(parts[1], `"`+" \t")
					libPath = strings.ReplaceAll(libPath, `\\`, `\`)
					for _, sub := range subnames {
						candidate := filepath.Join(libPath, "steamapps", sub)
						if info, err := os.Stat(candidate); err == nil && info.IsDir() {
							cfg.CSGOPath = candidate
							_ = SaveConfig(cfg)
							return candidate, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no se encontró la carpeta 'csgo'. Por favor configúrala en Configuración")
}

// LaunchInGame copies/links the demo into CS:GO's folder if needed, and launches +playdemo.
func LaunchInGame(demoPath string, tick int) error {
	csgoDir, err := FindCSGODirectory()
	if err != nil {
		return err
	}

	baseName := filepath.Base(demoPath)
	targetDemo := filepath.Join(csgoDir, baseName)

	// If demo is not already inside csgoDir, link or copy it
	if _, err := os.Stat(targetDemo); os.IsNotExist(err) {
		// Try hardlink first (instant, 0 bytes extra disk space)
		if errLink := os.Link(demoPath, targetDemo); errLink != nil {
			// Fallback: copy file
			data, errRead := os.ReadFile(demoPath)
			if errRead != nil {
				return fmt.Errorf("error leyendo demo: %v", errRead)
			}
			if errWrite := os.WriteFile(targetDemo, data, 0644); errWrite != nil {
				return fmt.Errorf("error copiando demo a csgo: %v", errWrite)
			}
		}
	}

	cfg := LoadConfig()
	prerollTicks := int(cfg.PrerollSeconds * 64.0)
	if prerollTicks < 64 {
		prerollTicks = 192
	}
	preroll := tick - prerollTicks
	if preroll < 128 {
		preroll = 128
	}

	nameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Priority 1: Direct executable launch if csgo.exe exists in parent
	parentDir := filepath.Dir(csgoDir)
	csgoExe := filepath.Join(parentDir, "csgo.exe")
	if info, err := os.Stat(csgoExe); err == nil && !info.IsDir() {
		cmd := exec.Command(csgoExe, "-steam", "-novid", "+playdemo", nameNoExt, "+demo_gototick", fmt.Sprintf("%d", preroll))
		cmd.Dir = parentDir
		return cmd.Start()
	}

	// Priority 2: Steam protocol URL
	steamURL := fmt.Sprintf("steam://run/730//+playdemo %s +demo_gototick %d", nameNoExt, preroll)
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", steamURL)
	return cmd.Start()
}
