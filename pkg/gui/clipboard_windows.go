package gui

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

// SetWindowsClipboard writes text directly to the native Windows OS clipboard.
func SetWindowsClipboard(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}
	bytesNeeded := uintptr(len(utf16) * 2)

	hMem, _, errAlloc := procGlobalAlloc.Call(gmemMoveable, bytesNeeded)
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc falló: %v", errAlloc)
	}

	ptr, _, errLock := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return fmt.Errorf("GlobalLock falló: %v", errLock)
	}

	src := unsafe.Pointer(&utf16[0])
	dst := unsafe.Pointer(ptr)
	copy((*[1 << 30]byte)(dst)[:bytesNeeded], (*[1 << 30]byte)(src)[:bytesNeeded])

	procGlobalUnlock.Call(hMem)

	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("no se pudo abrir el portapapeles de Windows")
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()
	ret, _, errSet := procSetClipboardData.Call(cfUnicodeText, hMem)
	if ret == 0 {
		return fmt.Errorf("SetClipboardData falló: %v", errSet)
	}

	return nil
}
