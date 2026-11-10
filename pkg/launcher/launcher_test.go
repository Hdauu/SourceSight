package launcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSteamLaunchInfo(t *testing.T) {
	root := t.TempDir()
	steamapps := filepath.Join(root, "steamapps")
	game := filepath.Join(steamapps, "common", "csgo legacy")
	if err := os.MkdirAll(game, 0755); err != nil {
		t.Fatal(err)
	}
	steamExe := filepath.Join(root, "steam.exe")
	os.WriteFile(steamExe, []byte("x"), 0644)
	acf := "\"AppState\"\n{\n\t\"appid\"\t\t\"4465480\"\n\t\"LauncherPath\"\t\t\"" +
		filepath.ToSlash(steamExe) + "\"\n\t\"installdir\"\t\t\"csgo legacy\"\n}\n"
	os.WriteFile(filepath.Join(steamapps, "appmanifest_4465480.acf"), []byte(acf), 0644)
	os.WriteFile(filepath.Join(steamapps, "appmanifest_730.acf"), []byte("\"AppState\"\n{\n\t\"appid\"\t\t\"730\"\n\t\"installdir\"\t\t\"Counter-Strike Global Offensive\"\n}\n"), 0644)

	exe, id := steamLaunchInfo(game)
	if id != "4465480" {
		t.Errorf("appID = %s, want 4465480 (CS:GO legacy, not 730 = CS2)", id)
	}
	if exe == "" {
		t.Errorf("steam.exe not found")
	}
}
