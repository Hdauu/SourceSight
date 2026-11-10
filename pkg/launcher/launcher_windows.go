package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
// LaunchInGame shows the demo in CS:GO at the given tick (demo ticks; tickrate turns
// the configured pre-roll seconds into ticks). An open game started by SourceSight is
// reused through its remote console; otherwise the game is started through Steam.
func LaunchInGame(demoPath string, tick, tickrate int) error {
	csgoDir, err := FindCSGODirectory()
	if err != nil {
		return err
	}
	ref, err := demoRef(csgoDir, demoPath)
	if err != nil {
		return err
	}

	if tickrate <= 0 {
		tickrate = 64
	}
	cfg := LoadConfig()
	prerollTicks := int(cfg.PrerollSeconds * float64(tickrate))
	if prerollTicks <= 0 {
		prerollTicks = 3 * tickrate
	}
	preroll := max(tick-prerollTicks, 1)

	if sendToRunningGame(ref, preroll) {
		return nil
	}
	if gameRunning() {
		return ErrGameWithoutConsole
	}

	// The demo is not passed on the command line ("+playdemo" there is ignored by the
	// Panorama menu); launchAndPlay sends it through the console once the menu is up.
	gameArgs := []string{"-novid", "-netconport", fmt.Sprint(NetconPort)}

	// The game must be started by Steam: running csgo.exe directly fails with
	// "Failed to connect with local Steam Client process". CS:GO Legacy is its own
	// Steam app (4465480), not 730 (now CS2), so the app ID comes from Steam's manifest.
	steamExe, appID := steamLaunchInfo(filepath.Dir(csgoDir))
	var start *exec.Cmd
	if steamExe != "" {
		start = exec.Command(steamExe, append([]string{"-applaunch", appID}, gameArgs...)...)
	} else {
		steamURL := fmt.Sprintf("steam://run/%s//%s", appID, strings.Join(gameArgs, " "))
		start = exec.Command("rundll32", "url.dll,FileProtocolHandler", steamURL)
	}
	if err := start.Start(); err != nil {
		return err
	}
	go launchAndPlay(ref, preroll)
	return nil
}

var acfField = regexp.MustCompile(`"(\w+)"\s+"([^"]*)"`)

// steamLaunchInfo finds steam.exe and the Steam app ID of the game installed in
// gameRoot (…\steamapps\common\<installdir>) by reading Steam's app manifests.
func steamLaunchInfo(gameRoot string) (steamExe, appID string) {
	installDir := filepath.Base(gameRoot)
	steamapps := filepath.Dir(filepath.Dir(gameRoot))

	manifests, _ := filepath.Glob(filepath.Join(steamapps, "appmanifest_*.acf"))
	for _, m := range manifests {
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		fields := map[string]string{}
		for _, f := range acfField.FindAllStringSubmatch(string(data), 40) {
			if _, seen := fields[f[1]]; !seen {
				fields[f[1]] = f[2]
			}
		}
		if strings.EqualFold(fields["installdir"], installDir) && fields["appid"] != "" {
			appID = fields["appid"]
			steamExe = strings.ReplaceAll(fields["LauncherPath"], `\\`, `\`) // .acf escapes backslashes
			break
		}
	}

	if appID == "" {
		appID = "730"
		if strings.Contains(strings.ToLower(installDir), "legacy") {
			appID = "4465480" // Counter-Strike: Global Offensive (legacy)
		}
	}
	if steamExe == "" || !fileExists(steamExe) {
		steamExe = ""
		if p := filepath.Join(filepath.Dir(steamapps), "steam.exe"); fileExists(p) {
			steamExe = p
		}
	}
	return steamExe, appID
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
