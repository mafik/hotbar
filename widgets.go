package main

import (
	"fmt"
	"image/png"
	"os"
	"reflect"
	"unicode/utf8"

	"github.com/nfnt/resize"

	"github.com/veandco/go-sdl2/sdl"
)

var glyphCache = make(map[rune]*sdl.Texture)
var iconCache = make(map[string]*sdl.Texture)

func MeasureText(text string) int {
	return utf8.RuneCountInString(text) * T.FontHeight / 2
}

func FillText(x int, text string, color []uint8) int {
	bytes := []byte(text)

	dst := sdl.Rect{
		X: int32(x),
		Y: int32(BarHeight/2 - LineHeight/2),
		W: int32(T.FontHeight / 2),
		H: int32(LineHeight)}

	for {
		r, size := utf8.DecodeRune(bytes)
		if size == 0 {
			break
		}
		bytes = bytes[size:]
		if r == utf8.RuneError {
			continue
		}
		texture, ok := glyphCache[r]
		if !ok {
			white := sdl.Color{R: 255, G: 255, B: 255, A: 255}
			surface, err := Font.RenderUTF8Blended(string(r), white)
			if err != nil {
				break
			}
			texture, err = R.CreateTextureFromSurface(surface)
			if err != nil {
				break
			}
			glyphCache[r] = texture
			surface.Free()
		}
		if color != nil {
			texture.SetColorMod(color[0], color[1], color[2])
		}
		R.Copy(texture, nil, &dst)
		dst.X += int32(T.FontHeight / 2)
	}
	return int(dst.X) - x
}

func GetIconTexture(path string) *sdl.Texture {
	t, ok := iconCache[path]
	if !ok {
		f, err := os.Open(path)
		if err != nil {
			fmt.Println("Couldn't load", path, ":", err)
			return nil
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			fmt.Println("Couldn't decode", path, ":", err)
			return nil
		}
		img = resize.Thumbnail(uint(T.IconSize), uint(T.IconSize), img, resize.NearestNeighbor)
		t, err = R.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, int32(T.IconSize), int32(T.IconSize))
		t.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			fmt.Println("Couldn't create SDL texture:", err)
			return nil
		}
		data, _, err := t.Lock(nil)
		if err != nil {
			fmt.Println("Couldn't lock SDL texture:", err)
		}
		for x := 0; x < T.IconSize; x++ {
			for y := 0; y < T.IconSize; y++ {
				base := (y*T.IconSize + x) * 4
				r, g, b, a := img.At(x, y).RGBA()
				if a != 0 {
					data[base] = byte(b * 255 / a)
					data[base+1] = byte(g * 255 / a)
					data[base+2] = byte(r * 255 / a)
					data[base+3] = byte(a / 256)
				}
			}
		}
		t.Unlock()
		iconCache[path] = t
	}
	return t
}

type Icon string

func Width(class string, blocks ...interface{}) int {
	c := getClass(class)
	w := c.Padding
	for _, block := range blocks {
		switch x := block.(type) {
		case Icon:
			w += T.IconSize + c.Padding
		case string:
			w += MeasureText(x) + c.Padding
		default:
			panic("Unsupported type: " + reflect.TypeOf(x).String())
		}
	}
	return w
}

func getClass(class string) Class {
	c, ok := T.Classes[class]
	if !ok {
		c = T.Default
	}
	if c.Foreground == nil {
		c.Foreground = T.Default.Foreground
	}
	if c.Background == nil {
		c.Background = T.Default.Background
		c.Fill = T.Default.Fill
	}
	if c.Fill == nil {
		c.Fill = c.Background
	}
	return c
}

func Draw(x int, class string, fill float32, blocks ...interface{}) {
	c := getClass(class)

	rect := sdl.Rect{
		X: int32(x),
		Y: BarHeight - int32(BarHeight*fill),
		W: int32(Width(class, blocks...)),
		H: int32(BarHeight * fill),
	}
	R.SetDrawColorArray(c.Fill...)
	R.FillRect(&rect)

	rect.H = rect.Y
	rect.Y = 0
	R.SetDrawColorArray(c.Background...)
	R.FillRect(&rect)

	x += c.Padding
	for _, block := range blocks {
		switch b := block.(type) {
		case Icon:
			path := Dir + "/" + T.IconDir + "/" + string(b) + ".png"
			t := GetIconTexture(path)
			if t == nil {
				fmt.Println("Couldn't find icon " + string(b) + " at " + path)
				continue
			}
			if T.TintIcons {
				t.SetColorMod(c.Foreground[0], c.Foreground[1], c.Foreground[2])
			}
			dst := sdl.Rect{X: int32(x), Y: BarHeight/2 - int32(T.IconSize/2), W: int32(T.IconSize), H: int32(T.IconSize)}
			R.Copy(t, nil, &dst)
			x += T.IconSize + c.Padding
		case string:
			FillText(x, b, c.Foreground)
			x += MeasureText(b) + c.Padding
		default:
			panic("Unsupported type: " + reflect.TypeOf(x).String())
		}
	}
}
