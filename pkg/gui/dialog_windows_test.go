package gui

import (
	"syscall"
	"testing"
)

func utf16z(parts ...string) []uint16 {
	var out []uint16
	for _, p := range parts {
		s, _ := syscall.UTF16FromString(p)
		out = append(out, s...)
	}
	return append(out, 0, 0, 0)
}

func TestSplitMultiSelect(t *testing.T) {
	single := splitMultiSelect(utf16z(`C:\demos\a.dem`))
	if len(single) != 1 || single[0] != `C:\demos\a.dem` {
		t.Fatalf("single: %v", single)
	}
	multi := splitMultiSelect(utf16z(`C:\demos`, "a.dem", "b.dem"))
	if len(multi) != 2 || multi[0] != `C:\demos\a.dem` || multi[1] != `C:\demos\b.dem` {
		t.Fatalf("multi: %v", multi)
	}
}
