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

// sendToRunningGame plays the demo in a CS:GO started with -netconport. When the
// same demo is already playing it only seeks; otherwise it loads the demo, waits for
// playback to start (or a timeout) and then seeks.
func sendToRunningGame(ref string, tick int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", NetconPort), 400*time.Millisecond)
	if err != nil {
		return false
	}
	go func() {
		defer conn.Close()
		consoleMu.Lock()
		defer consoleMu.Unlock()

		send := func(cmd string) error {
			_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			_, err := conn.Write([]byte(cmd + "\n"))
			return err
		}
		gotoCmd := fmt.Sprintf("demo_gototick %d", tick)
		if ref == lastPlayed {
			_ = send(gotoCmd)
			return
		}
		if send(fmt.Sprintf(`playdemo "%s"`, ref)) != nil {
			return
		}
		lastPlayed = ref
		waitForPlayback(conn, 15*time.Second)
		time.Sleep(time.Second) // let the demo settle before seeking
		_ = send(gotoCmd)
	}()
	return true
}

// waitForPlayback reads console output until the demo starts playing or timeout.
func waitForPlayback(conn net.Conn, timeout time.Duration) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		if strings.Contains(strings.ToLower(sc.Text()), "playing demo") {
			return
		}
	}
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
