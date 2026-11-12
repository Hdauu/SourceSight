package gui

import (
	"runtime"
	"sync"
	"unsafe"

	"sourcesight/pkg/applog"
)

// Global hotkeys to jump to the previous / next incident while the game has focus.
// Windows delivers them to SourceSight even when CS:GO is in front; SourceSight then
// sends the seek to the game through its remote console. Nothing is written to the
// game's binds or config.

var (
	procRegisterHotKey    = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey  = user32.NewProc("UnregisterHotKey")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procPostThreadMessage = user32.NewProc("PostThreadMessageW")
	procGetCurrentThread  = kernel32.NewProc("GetCurrentThreadId")
)

const (
	modAlt      = 0x1
	modControl  = 0x2
	modShift    = 0x4
	modNoRepeat = 0x4000
	vkPrior     = 0x21 // Page Up
	vkNext      = 0x22 // Page Down
	vkLeft      = 0x25
	vkRight     = 0x27
	wmHotkey    = 0x0312
	wmQuit      = 0x0012
)

type hotkey struct{ mod, vk uint32 }

// HotkeyPresets are the choices offered in Settings: previous, next and a label.
// Function keys are avoided on purpose: CS:GO binds several (F10 opens the quit
// dialog), and modifiers keep these from clashing with the game.
var HotkeyPresets = map[string][2]hotkey{
	"ctrl-alt-arrows":   {{modControl | modAlt, vkLeft}, {modControl | modAlt, vkRight}},
	"ctrl-shift-arrows": {{modControl | modShift, vkLeft}, {modControl | modShift, vkRight}},
	"ctrl-alt-pages":    {{modControl | modAlt, vkPrior}, {modControl | modAlt, vkNext}},
	"alt-pages":         {{modAlt, vkPrior}, {modAlt, vkNext}},
}

// DefaultHotkeys is used when the setting is empty.
const DefaultHotkeys = "ctrl-alt-arrows"

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

var (
	hkMu       sync.Mutex
	hkThreadID uintptr
	hkDone     chan struct{}
)

// setHotkeys (re)registers the hotkeys for a preset; "off" or an unknown preset
// removes them. onStep receives -1 or +1.
func setHotkeys(preset string, onStep func(int)) {
	hkMu.Lock()
	defer hkMu.Unlock()
	if hkThreadID != 0 {
		procPostThreadMessage.Call(hkThreadID, wmQuit, 0, 0)
		<-hkDone
		hkThreadID = 0
	}
	if preset == "" {
		preset = DefaultHotkeys
	}
	keys, ok := HotkeyPresets[preset]
	if !ok {
		return
	}
	started := make(chan uintptr)
	hkDone = make(chan struct{})
	go func() {
		// Hotkeys belong to the thread that registered them: keep this goroutine on it
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(hkDone)
		tid, _, _ := procGetCurrentThread.Call()
		ok := 0
		for i, k := range keys {
			if r, _, err := procRegisterHotKey.Call(0, uintptr(i+1), uintptr(k.mod|modNoRepeat), uintptr(k.vk)); r == 0 {
				applog.Printf("teclas: no se pudo registrar la tecla %d (¿la usa otro programa?): %v", i+1, err)
			} else {
				ok++
			}
		}
		started <- tid
		if ok == 0 {
			return
		}
		defer func() {
			procUnregisterHotKey.Call(0, 1)
			procUnregisterHotKey.Call(0, 2)
		}()
		var m msg
		for {
			r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if int32(r) <= 0 { // WM_QUIT or error
				return
			}
			if m.message == wmHotkey {
				if m.wParam == 1 {
					onStep(-1)
				} else {
					onStep(+1)
				}
			}
		}
	}()
	hkThreadID = <-started
}
