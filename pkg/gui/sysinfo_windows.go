package gui

import (
	"runtime"
	"unsafe"
)

var procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// availableMemory returns free physical memory in bytes, or 0 if unknown.
func availableMemory() uint64 {
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	if r, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m))); r == 0 {
		return 0
	}
	return m.AvailPhys
}

// memoryPerAnalysis is the budget for one demo being analyzed: a 207 MB demo peaked
// at ~210 MB of heap, so 300 MB leaves room for garbage-collector overhead.
const memoryPerAnalysis = 300 << 20

// workerCount decides how many demos to analyze at once: one per CPU thread minus
// one (kept for the interface), capped by half of the free memory, never more than
// the number of demos.
func workerCount(demos int) int {
	n := max(1, runtime.NumCPU()-1)
	if free := availableMemory(); free > 0 {
		n = min(n, max(1, int(free/2/memoryPerAnalysis)))
	}
	return max(1, min(n, demos))
}
