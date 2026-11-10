package gui

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
)

type openFileNameW struct {
	StructSize      uint32
	Hwnd            uintptr
	HInstance       uintptr
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustData        uintptr
	FnHook          uintptr
	TemplateName    *uint16
	PvReserved      unsafe.Pointer
	DwReserved      uint32
	FlagsEx         uint32
}

// OpenDemoDialog opens the native Windows file dialog to pick one or more .dem files.
// It returns nil when the user cancels.
func OpenDemoDialog(owner uintptr) ([]string, error) {
	fileBuf := make([]uint16, 64*1024)
	filter, _ := syscall.UTF16PtrFromString("Demos de Counter-Strike (*.dem)\x00*.dem\x00Todos los archivos (*.*)\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("SourceSight — Abrir demos")

	var ofn openFileNameW
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	ofn.Hwnd = owner // modal over the app window, so it never opens hidden behind it
	ofn.Filter = filter
	ofn.File = &fileBuf[0]
	ofn.MaxFile = uint32(len(fileBuf))
	ofn.Title = title
	ofn.Flags = 0x00080000 | 0x00001000 | 0x00000200 // OFN_EXPLORER | OFN_FILEMUSTEXIST | OFN_ALLOWMULTISELECT

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return nil, nil // user cancelled
	}
	return splitMultiSelect(fileBuf), nil
}

// splitMultiSelect decodes the OFN_ALLOWMULTISELECT buffer: either a single full
// path, or a directory followed by file names, each NUL-terminated, ending in "\0\0".
func splitMultiSelect(buf []uint16) []string {
	var parts []string
	start := 0
	for i, c := range buf {
		if c != 0 {
			continue
		}
		if i == start {
			break
		}
		parts = append(parts, syscall.UTF16ToString(buf[start:i]))
		start = i + 1
	}
	if len(parts) <= 1 {
		return parts
	}
	dir := parts[0]
	paths := make([]string, 0, len(parts)-1)
	for _, name := range parts[1:] {
		paths = append(paths, filepath.Join(dir, name))
	}
	return paths
}

// OpenCSGODialog opens the native Windows File Explorer dialog to pick csgo.exe or game directory.
func OpenCSGODialog(owner uintptr) (string, error) {
	fileBuf := make([]uint16, 2048)
	filter, _ := syscall.UTF16PtrFromString("Ejecutable de CS:GO (csgo.exe)\x00csgo.exe\x00Todos los archivos (*.*)\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("SourceSight — Seleccionar csgo.exe o carpeta del juego")

	var ofn openFileNameW
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	ofn.Hwnd = owner // modal over the app window, so it never opens hidden behind it
	ofn.Filter = filter
	ofn.File = &fileBuf[0]
	ofn.MaxFile = uint32(len(fileBuf))
	ofn.Title = title
	ofn.Flags = 0x00080000 | 0x00001000 // OFN_EXPLORER | OFN_FILEMUSTEXIST

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", nil // user cancelled
	}
	return syscall.UTF16ToString(fileBuf), nil
}
