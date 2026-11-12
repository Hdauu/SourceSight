package launcher

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// NetconPort is the local TCP port of CS:GO's remote console (-netconport). Games
// started by SourceSight listen on it, so later "Watch" clicks reuse the open game.
const NetconPort = 2121

// ErrGameWithoutConsole means CS:GO is running but wasn't started with -netconport,
// so SourceSight can't send it the demo.
var ErrGameWithoutConsole = errors.New("CS:GO ya está abierto pero sin control remoto: cerralo una vez y tocá Ver de nuevo (o agregá -netconport 2121 a sus opciones de lanzamiento en Steam)")

const createNoWindow = 0x08000000

var (
	consoleMu  sync.Mutex
	lastPlayed string // demo reference last sent to the running game
)

// demoRef returns how +playdemo must name the demo: a path relative to the csgo
// folder. Demos elsewhere are reached through a folder junction inside csgo (works
// across drives, no copying, no admin rights), one per source folder.
func demoRef(csgoDir, demoPath string) (string, error) {
	name := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath))
	srcDir, err := filepath.Abs(filepath.Dir(demoPath))
	if err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(csgoDir, srcDir); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(filepath.Join(rel, name)), nil
	}

	sum := sha1.Sum([]byte(strings.ToLower(srcDir)))
	link := "sourcesight_" + hex.EncodeToString(sum[:4])
	linkPath := filepath.Join(csgoDir, link)
	if _, err := os.Stat(filepath.Join(linkPath, filepath.Base(demoPath))); err == nil {
		return link + "/" + name, nil
	}
	_ = os.Remove(linkPath) // stale junction pointing elsewhere
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, srcDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("no se pudo enlazar la carpeta de la demo: %v %s", err, strings.TrimSpace(string(out)))
	}
	return link + "/" + name, nil
}

// sendToRunningGame plays the demo in an open CS:GO started with -netconport.
// It returns false when no game console is listening.
func sendToRunningGame(ref string, t Target) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", NetconPort), 400*time.Millisecond)
	if err != nil {
		return false
	}
	go playOverConsole(conn, ref, t, false)
	return true
}

// launchAndPlay waits for a game just started by Steam to open its console and reach
// the main menu, then plays the demo. "+playdemo" on the command line doesn't work:
// it runs before the Panorama menu is ready and the game ignores it.
func launchAndPlay(ref string, t Target) {
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", NetconPort), time.Second)
		if err == nil {
			playOverConsole(conn, ref, t, true)
			return
		}
		time.Sleep(time.Second)
	}
}

// playOverConsole loads the demo (unless it's the one already playing) and seeks to
// tick once the map has loaded. Lines come from the game's console output.
func playOverConsole(conn net.Conn, ref string, t Target, freshLaunch bool) {
	defer conn.Close()
	consoleMu.Lock()
	defer consoleMu.Unlock()

	lines := make(chan string, 256)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(conn)
		for sc.Scan() {
			lines <- sc.Text()
		}
	}()
	waitFor := func(substr string, timeout time.Duration) bool {
		limit := time.After(timeout)
		for {
			select {
			case l, ok := <-lines:
				if !ok {
					return false
				}
				if strings.Contains(l, substr) {
					return true
				}
			case <-limit:
				return false
			}
		}
	}
	send := func(cmd string) error {
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := conn.Write([]byte(cmd + "\n"))
		return err
	}
	seek := func() {
		for i, cmd := range t.commands() {
			if i == 1 {
				// The camera commands must come after the seek has landed
				time.Sleep(400 * time.Millisecond)
			}
			if send(cmd) != nil {
				return
			}
		}
	}

	if freshLaunch {
		waitFor("-> CSGO_GAME_UI_STATE_MAINMENU", 45*time.Second)
		time.Sleep(2 * time.Second)
	} else if ref == lastPlayed {
		seek()
		return
	}

	if send(fmt.Sprintf(`playdemo "%s"`, ref)) != nil {
		return
	}
	lastPlayed = ref
	// Seeking during the loading screen is ignored: wait until the map is in game
	waitFor("-> CSGO_GAME_UI_STATE_INGAME", 90*time.Second)
	time.Sleep(1500 * time.Millisecond)
	seek()
}

// gameRunning reports whether csgo.exe is running.
func gameRunning() bool {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(snap)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	for err = windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		if strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), "csgo.exe") {
			return true
		}
	}
	return false
}

// Target is a moment to show: a tick and, optionally, the player whose eyes to watch
// it from and a short note printed in the game console.
type Target struct {
	Tick    int
	SteamID uint64
	Note    string
}

// steamID64Base turns a SteamID64 into the 32-bit account ID the game uses.
const steamID64Base = 76561197960265728

// commands are the console commands for a target: seek first, then the camera on the
// player in first person, then the note.
func (t Target) commands() []string {
	cmds := []string{fmt.Sprintf("demo_gototick %d", t.Tick)}
	if t.SteamID > steamID64Base {
		cmds = append(cmds,
			fmt.Sprintf("spec_player_by_accountid %d", t.SteamID-steamID64Base),
			"spec_mode 4")
	}
	if note := consoleSafe(t.Note); note != "" {
		cmds = append(cmds, `echo "SourceSight: `+note+`"`)
	}
	return cmds
}

// consoleSafe keeps text that can't break out of an echo: player names come from the
// demo and anyone can put ";" or quotes in theirs to run commands in the console.
func consoleSafe(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == ';' || r == '"' || r == '\'' || r == '\\' || r < ' ' || r == 0x7f:
			continue
		case b.Len() >= 120:
			return b.String()
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
