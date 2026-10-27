package main

// The tray icon, drawn at start: a gold coin with a dark rim, like the add-on's own icon. Windows wants an
// .ico, so on Windows the PNG goes inside one.

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"runtime"
)

func coinPNG(size int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	c := float64(size-1) / 2
	gold := color.NRGBA{0xf0, 0xc3, 0x5a, 0xff}
	deep := color.NRGBA{0xb8, 0x86, 0x0b, 0xff}
	rim := color.NRGBA{0x3b, 0x2a, 0x17, 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			d := math.Hypot(float64(x)-c, float64(y)-c) / c
			switch {
			case d > 1:
				continue
			case d > 0.86:
				img.Set(x, y, rim)
			case d > 0.62:
				img.Set(x, y, deep)
			default:
				img.Set(x, y, gold)
			}
		}
	}
	// A small dark bar across the middle, like the stamp on a coin.
	for y := int(c) - size/16; y <= int(c)+size/16; y++ {
		for x := int(c) - size/5; x <= int(c)+size/5; x++ {
			img.Set(x, y, rim)
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

func trayIcon() []byte {
	p := coinPNG(32)
	if runtime.GOOS != "windows" {
		return p
	}
	var b bytes.Buffer
	// ICONDIR, then one ICONDIRENTRY pointing at the PNG.
	_ = binary.Write(&b, binary.LittleEndian, []uint16{0, 1, 1})
	b.Write([]byte{32, 32, 0, 0})
	_ = binary.Write(&b, binary.LittleEndian, []uint16{1, 32})
	_ = binary.Write(&b, binary.LittleEndian, []uint32{uint32(len(p)), 22})
	b.Write(p)
	return b.Bytes()
}
