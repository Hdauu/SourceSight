package gui

import (
	_ "embed"
	"unsafe"
)

// The window icon is set at runtime so local builds show it without a resource
// compiler; release builds additionally embed it as the .exe icon (see winres/).

//go:embed assets/icon_16.png
var icon16 []byte

//go:embed assets/icon_32.png
var icon32 []byte

//go:embed assets/icon_48.png
var icon48 []byte

var (
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
)

const (
	wmSetIcon      = 0x0080
	iconSmall      = 0
	iconBig        = 1
	smCxSmIcon     = 49
	lrDefaultColor = 0x0000
)

// iconFromPNG builds an HICON from PNG bytes (supported by Windows Vista and later).
func iconFromPNG(png []byte, size int) uintptr {
	if len(png) == 0 {
		return 0
	}
	h, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&png[0])),
		uintptr(len(png)),
		1,          // fIcon
		0x00030000, // required version
		uintptr(size), uintptr(size),
		lrDefaultColor,
	)
	return h
}

// setWindowIcon assigns the SourceSight icon to the title bar and taskbar.
func setWindowIcon(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	small := icon16
	smallSize := 16
	if cx, _, _ := procGetSystemMetrics.Call(smCxSmIcon); cx > 20 {
		small, smallSize = icon32, 32 // scaled displays
	}
	if h := iconFromPNG(small, smallSize); h != 0 {
		procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, h)
	}
	if h := iconFromPNG(icon48, 48); h != 0 {
		procSendMessageW.Call(hwnd, wmSetIcon, iconBig, h)
	}
}
