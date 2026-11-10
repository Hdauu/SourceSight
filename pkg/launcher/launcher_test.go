package launcher

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDemoRefJunction(t *testing.T) {
	csgo := t.TempDir()
	demos := t.TempDir()
	demo := filepath.Join(demos, "s1-test de_mirage.dem")
	os.WriteFile(demo, []byte("demo"), 0644)

	ref, err := demoRef(csgo, demo)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(ref) != "" || filepath.Base(ref) != "s1-test de_mirage" {
		t.Fatalf("ref = %q", ref)
	}
	// The demo must be readable through the junction, as the game will read it
	data, err := os.ReadFile(filepath.Join(csgo, filepath.FromSlash(ref)+".dem"))
	if err != nil || string(data) != "demo" {
		t.Fatalf("demo not reachable through %q: %v", ref, err)
	}
	// Second call reuses the junction
	if ref2, err := demoRef(csgo, demo); err != nil || ref2 != ref {
		t.Fatalf("second call: %q %v", ref2, err)
	}
	// Demos already inside csgo need no link
	inside := filepath.Join(csgo, "replays", "x.dem")
	os.MkdirAll(filepath.Dir(inside), 0755)
	os.WriteFile(inside, []byte("x"), 0644)
	if ref, _ := demoRef(csgo, inside); ref != "replays/x" {
		t.Fatalf("inside ref = %q", ref)
	}
}

// fakeGame mimics CS:GO's -netconport console: it records commands and announces
// playback after a playdemo.
func fakeGame(t *testing.T) (got chan string) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", NetconPort))
	if err != nil {
		t.Skipf("port %d busy: %v", NetconPort, err)
	}
	t.Cleanup(func() { ln.Close() })
	got = make(chan string, 10)
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				sc := bufio.NewScanner(c)
				for sc.Scan() {
					got <- sc.Text()
					if strings.HasPrefix(sc.Text(), "playdemo") {
						c.Write([]byte("Playing demo from x.dem.\r\n"))
					}
				}
			}()
		}
	}()
	return got
}

func TestSendToRunningGame(t *testing.T) {
	got := fakeGame(t)
	lastPlayed = ""
	if !sendToRunningGame("sourcesight_ab/demo1", 5000) {
		t.Fatal("did not connect to the console")
	}
	want := []string{`playdemo "sourcesight_ab/demo1"`, "demo_gototick 5000"}
	for _, w := range want {
		select {
		case line := <-got:
			if line != w {
				t.Fatalf("got %q, want %q", line, w)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for %q", w)
		}
	}
	// Same demo again: only seek, no reload
	sendToRunningGame("sourcesight_ab/demo1", 9000)
	select {
	case line := <-got:
		if line != "demo_gototick 9000" {
			t.Fatalf("same demo should only seek, got %q", line)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for seek")
	}
}
