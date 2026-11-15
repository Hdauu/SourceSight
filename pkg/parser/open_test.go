package parser

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// openDemo must return exactly the file's bytes, whatever the size of the reads
// (unbuffered reads come in 8 MB blocks; the last one is short).
func TestOpenDemoReadsTheWholeFile(t *testing.T) {
	dir := t.TempDir()
	for _, size := range []int{0, 1, 4095, 4096, readBlock - 1, readBlock, readBlock + 12345} {
		want := make([]byte, size)
		for i := range want {
			want[i] = byte(i * 7)
		}
		path := filepath.Join(dir, "d.dem")
		if err := os.WriteFile(path, want, 0644); err != nil {
			t.Fatal(err)
		}
		f, err := openDemo(path)
		if err != nil {
			t.Fatal(err)
		}
		// Odd read sizes, like the parser's
		var got bytes.Buffer
		buf := make([]byte, 131072+17)
		for {
			n, err := f.Read(buf)
			got.Write(buf[:n])
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		f.Close()
		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("size %d: read %d bytes, different from the file", size, got.Len())
		}
	}
}

// ParseDemoHashed gives the same report as ParseDemo and the same hash as hashing
// the file on its own.
//
// Needs demos: SOURCESIGHT_TEST_DEMOS=<carpeta-con-demos> go test ./pkg/parser -run Hashed -v
func TestParseDemoHashed(t *testing.T) {
	dir := os.Getenv("SOURCESIGHT_TEST_DEMOS")
	if dir == "" {
		t.Skip("SOURCESIGHT_TEST_DEMOS not set")
	}
	demos, _ := filepath.Glob(filepath.Join(dir, "*.dem"))
	if len(demos) == 0 {
		t.Skip("no demos")
	}
	path := demos[0]
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)

	rep, sha, err := ParseDemoHashed(path, ParserOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if sha != hex.EncodeToString(sum[:]) {
		t.Errorf("hash %s, want %s", sha, hex.EncodeToString(sum[:]))
	}
	plain, err := ParseDemo(path, ParserOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rep.Header.Date, plain.Header.Date = "", ""
	a, _ := json.Marshal(rep)
	b, _ := json.Marshal(plain)
	if !bytes.Equal(a, b) {
		t.Error("report differs from ParseDemo")
	}
}
