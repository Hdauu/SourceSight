//go:build !windows

package parser

import (
	"bufio"
	"io"
	"os"
)

// readBlock: read size, as on Windows (open_windows.go)
const readBlock = 8 << 20

type bufferedFile struct {
	*bufio.Reader
	f *os.File
}

func (b bufferedFile) Close() error { return b.f.Close() }

// openDemo reads the demo in 8 MB blocks (see open_windows.go).
func openDemo(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufferedFile{bufio.NewReaderSize(f, readBlock), f}, nil
}
