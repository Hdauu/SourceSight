// Command icongen draws the SourceSight icon at the sizes Windows uses and writes
// PNGs that go-winres packs into the executables' .ico resource.
//
//	go run ./tools/icongen -out cmd/sourcesight/winres
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
)

var (
	bg     = color.RGBA{0x1b, 0x1c, 0x1f, 0xff}
	edge   = color.RGBA{0x2e, 0x30, 0x35, 0xff}
	accent = color.RGBA{0xe0, 0x7b, 0x2a, 0xff}
	light  = color.RGBA{0xec, 0xea, 0xe6, 0xff}
)

func draw(size int) *gg.Context {
	s := float64(size)
	dc := gg.NewContext(size, size)

	// Rounded tile
	pad := s * 0.04
	dc.DrawRoundedRectangle(pad, pad, s-2*pad, s-2*pad, s*0.2)
	dc.SetColor(bg)
	dc.FillPreserve()
	dc.SetColor(edge)
	dc.SetLineWidth(math.Max(1, s*0.02))
	dc.Stroke()

	c := s / 2
	r := s * 0.27
	lw := math.Max(1.5, s*0.075)

	// Scope ring, open at the four ticks so it reads as a sight, not a target
	dc.SetColor(light)
	dc.SetLineWidth(lw)
	dc.SetLineCapButt()
	gap := 0.32
	for q := 0; q < 4; q++ {
		start := float64(q)*math.Pi/2 + gap
		end := float64(q+1)*math.Pi/2 - gap
		dc.DrawArc(c, c, r, start, end)
		dc.Stroke()
	}

	// Ticks
	dc.SetColor(light)
	dc.SetLineCapRound()
	tIn, tOut := r*0.62, r*1.32
	for _, a := range []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2} {
		dc.DrawLine(c+math.Cos(a)*tIn, c+math.Sin(a)*tIn, c+math.Cos(a)*tOut, c+math.Sin(a)*tOut)
		dc.Stroke()
	}

	// Center dot
	dc.SetColor(accent)
	dc.DrawCircle(c, c, math.Max(1.5, s*0.07))
	dc.Fill()
	return dc
}

func main() {
	out := flag.String("out", ".", "output directory")
	ico := flag.String("ico", "", "also write a multi-size .ico to this path")
	flag.Parse()
	if err := os.MkdirAll(*out, 0755); err != nil {
		panic(err)
	}
	for _, size := range []int{16, 24, 32, 48, 64, 128, 256} {
		p := filepath.Join(*out, fmt.Sprintf("icon_%d.png", size))
		if err := draw(size).SavePNG(p); err != nil {
			panic(err)
		}
	}
	fmt.Println("icons written to", *out)

	if *ico != "" {
		if err := writeICO(*ico, []int{16, 24, 32, 48, 64, 128, 256}); err != nil {
			panic(err)
		}
		fmt.Println("ico written to", *ico)
	}
}

// writeICO packs PNG-encoded images into an .ico (supported since Windows Vista).
func writeICO(path string, sizes []int) error {
	var images [][]byte
	for _, size := range sizes {
		var buf bytes.Buffer
		if err := draw(size).EncodePNG(&buf); err != nil {
			return err
		}
		images = append(images, buf.Bytes())
	}

	var out bytes.Buffer
	w := func(v any) { _ = binary.Write(&out, binary.LittleEndian, v) }
	w(uint16(0))          // reserved
	w(uint16(1))          // type: icon
	w(uint16(len(sizes))) // image count
	offset := 6 + 16*len(sizes)
	for i, size := range sizes {
		dim := uint8(size)
		if size >= 256 {
			dim = 0 // 0 means 256
		}
		w(dim)                    // width
		w(dim)                    // height
		w(uint8(0))               // palette
		w(uint8(0))               // reserved
		w(uint16(1))              // color planes
		w(uint16(32))             // bits per pixel
		w(uint32(len(images[i]))) // data size
		w(uint32(offset))         // data offset
		offset += len(images[i])
	}
	for _, img := range images {
		out.Write(img)
	}
	return os.WriteFile(path, out.Bytes(), 0644)
}
