package parser

import (
	"bufio"
	"errors"
	"io"
	"os"
	"syscall"
	"unsafe"
)

// Reading a demo in 8 MB blocks, bypassing the Windows file cache.
//
// The parser asks for 128 KB at a time. With several demos analyzed at once on a hard
// disk, every 128 KB read of one demo sends the head to another file: measured on a
// laptop HDD with 11 demos at once, 30 MB/s in 128 KB reads against 105 MB/s in 8 MB
// reads. Bypassing the cache (FILE_FLAG_NO_BUFFERING) makes each read really one
// 8 MB disk read; on an SSD it costs nothing.
const (
	readBlock          = 8 << 20
	sectorAlign        = 4096
	fileFlagNoBuffer   = 0x20000000
	fileFlagSequential = 0x08000000
)

type blockReader struct {
	path    string
	h       syscall.Handle
	buf     []byte // aligned to sectorAlign, as unbuffered reads require
	r, w    int
	started bool
	// fallback is used when unbuffered reads aren't supported (some network drives)
	fallback io.ReadCloser
}

func openDemo(path string) (io.ReadCloser, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ, syscall.FILE_SHARE_READ, nil,
		syscall.OPEN_EXISTING, fileFlagNoBuffer|fileFlagSequential, 0)
	if err != nil {
		return openBuffered(path)
	}
	raw := make([]byte, readBlock+sectorAlign)
	off := (sectorAlign - int(uintptr(unsafe.Pointer(&raw[0]))%sectorAlign)) % sectorAlign
	return &blockReader{path: path, h: h, buf: raw[off : off+readBlock]}, nil
}

func (b *blockReader) Read(p []byte) (int, error) {
	if b.fallback != nil {
		return b.fallback.Read(p)
	}
	if b.r == b.w {
		var n uint32
		err := syscall.ReadFile(b.h, b.buf, &n, nil)
		if err != nil && !errors.Is(err, syscall.ERROR_HANDLE_EOF) {
			if !b.started {
				// Unbuffered reads refused before anything was read: read normally instead
				syscall.CloseHandle(b.h)
				b.h = syscall.InvalidHandle
				fb, ferr := openBuffered(b.path)
				if ferr != nil {
					return 0, ferr
				}
				b.fallback = fb
				return fb.Read(p)
			}
			return 0, err
		}
		b.started = true
		if n == 0 {
			return 0, io.EOF
		}
		b.r, b.w = 0, int(n)
	}
	n := copy(p, b.buf[b.r:b.w])
	b.r += n
	return n, nil
}

func (b *blockReader) Close() error {
	if b.fallback != nil {
		return b.fallback.Close()
	}
	if b.h != syscall.InvalidHandle {
		return syscall.CloseHandle(b.h)
	}
	return nil
}

type bufferedFile struct {
	*bufio.Reader
	f *os.File
}

func (b bufferedFile) Close() error { return b.f.Close() }

func openBuffered(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufferedFile{bufio.NewReaderSize(f, readBlock), f}, nil
}
