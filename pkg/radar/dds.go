package radar

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
)

// decodeDDS reads the radar textures CS:GO ships in resource/overviews: DXT1 or DXT5
// compressed (most maps) or uncompressed 32-bit BGRA. Only the top mipmap is read.
func decodeDDS(b []byte) (*image.NRGBA, error) {
	if len(b) < 128 || string(b[:4]) != "DDS " {
		return nil, errors.New("no es un archivo DDS")
	}
	h := int(binary.LittleEndian.Uint32(b[12:]))
	w := int(binary.LittleEndian.Uint32(b[16:]))
	if w <= 0 || h <= 0 || w > 8192 || h > 8192 {
		return nil, fmt.Errorf("tamaño inválido %dx%d", w, h)
	}
	pfFlags := binary.LittleEndian.Uint32(b[80:])
	fourCC := string(b[84:88])
	data := b[128:]
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	switch {
	case pfFlags&0x4 != 0 && fourCC == "DXT1":
		if len(data) < blocks(w, h)*8 {
			return nil, errors.New("DDS DXT1 incompleto")
		}
		forBlocks(w, h, func(i, bx, by int) {
			blk := data[i*8:]
			pal := colorPalette(blk, true)
			idx := binary.LittleEndian.Uint32(blk[4:])
			put(img, bx, by, func(p int) color.NRGBA { return pal[(idx>>(2*p))&3] })
		})
	case pfFlags&0x4 != 0 && fourCC == "DXT5":
		if len(data) < blocks(w, h)*16 {
			return nil, errors.New("DDS DXT5 incompleto")
		}
		forBlocks(w, h, func(i, bx, by int) {
			blk := data[i*16:]
			alpha := alphaPalette(blk[0], blk[1])
			var abits uint64
			for k := 0; k < 6; k++ {
				abits |= uint64(blk[2+k]) << (8 * k)
			}
			pal := colorPalette(blk[8:], false)
			idx := binary.LittleEndian.Uint32(blk[12:])
			put(img, bx, by, func(p int) color.NRGBA {
				c := pal[(idx>>(2*p))&3]
				c.A = alpha[(abits>>(3*p))&7]
				return c
			})
		})
	case pfFlags&0x40 != 0 && binary.LittleEndian.Uint32(b[88:]) == 32:
		if len(data) < w*h*4 {
			return nil, errors.New("DDS sin comprimir incompleto")
		}
		for p := 0; p < w*h; p++ {
			s := data[p*4:]
			a := s[3]
			if pfFlags&0x1 == 0 { // no alpha channel
				a = 255
			}
			img.Pix[p*4], img.Pix[p*4+1], img.Pix[p*4+2], img.Pix[p*4+3] = s[2], s[1], s[0], a
		}
	default:
		return nil, fmt.Errorf("formato DDS no soportado (%q)", fourCC)
	}
	return img, nil
}

func blocks(w, h int) int { return ((w + 3) / 4) * ((h + 3) / 4) }

func forBlocks(w, h int, fn func(i, bx, by int)) {
	i := 0
	for by := 0; by < h; by += 4 {
		for bx := 0; bx < w; bx += 4 {
			fn(i, bx, by)
			i++
		}
	}
}

// put writes the 16 pixels of a 4×4 block, clipping at the image edge.
func put(img *image.NRGBA, bx, by int, px func(p int) color.NRGBA) {
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if bx+x < img.Rect.Dx() && by+y < img.Rect.Dy() {
				img.SetNRGBA(bx+x, by+y, px(y*4+x))
			}
		}
	}
}

func rgb565(v uint16) color.NRGBA {
	r, g, b := uint8(v>>11&31), uint8(v>>5&63), uint8(v&31)
	return color.NRGBA{R: r<<3 | r>>2, G: g<<2 | g>>4, B: b<<3 | b>>2, A: 255}
}

// colorPalette builds the 4 colors of a DXT color block. DXT1 blocks with c0 <= c1
// use 3 colors plus transparent black.
func colorPalette(blk []byte, dxt1 bool) [4]color.NRGBA {
	v0 := binary.LittleEndian.Uint16(blk)
	v1 := binary.LittleEndian.Uint16(blk[2:])
	c0, c1 := rgb565(v0), rgb565(v1)
	mix := func(a, b uint8, wa, wb, d int) uint8 { return uint8((int(a)*wa + int(b)*wb) / d) }
	var p [4]color.NRGBA
	p[0], p[1] = c0, c1
	if !dxt1 || v0 > v1 {
		p[2] = color.NRGBA{mix(c0.R, c1.R, 2, 1, 3), mix(c0.G, c1.G, 2, 1, 3), mix(c0.B, c1.B, 2, 1, 3), 255}
		p[3] = color.NRGBA{mix(c0.R, c1.R, 1, 2, 3), mix(c0.G, c1.G, 1, 2, 3), mix(c0.B, c1.B, 1, 2, 3), 255}
	} else {
		p[2] = color.NRGBA{mix(c0.R, c1.R, 1, 1, 2), mix(c0.G, c1.G, 1, 1, 2), mix(c0.B, c1.B, 1, 1, 2), 255}
		p[3] = color.NRGBA{}
	}
	return p
}

// alphaPalette builds the 8 alpha values of a DXT5 alpha block.
func alphaPalette(a0, a1 uint8) [8]uint8 {
	var a [8]uint8
	a[0], a[1] = a0, a1
	if a0 > a1 {
		for i := 1; i < 7; i++ {
			a[i+1] = uint8((int(a0)*(7-i) + int(a1)*i) / 7)
		}
	} else {
		for i := 1; i < 5; i++ {
			a[i+1] = uint8((int(a0)*(5-i) + int(a1)*i) / 5)
		}
		a[6], a[7] = 0, 255
	}
	return a
}
