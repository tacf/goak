package sdl3

import (
	"math"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak"
)

// snap returns the integer nearest v and whether v is within pxEps of it.
func snap(v float32) (float32, bool) {
	n := math.Round(float64(v))
	return float32(n), math.Abs(float64(v)-n) < pxEps
}

// roundPx returns the nearest whole pixel edge with ties rounding down. A
// pixel is covered when its center (k+0.5) lies inside the rect, so this
// matches the coverage of the geometry path.
func roundPx(v float32) float32 { return float32(math.Ceil(float64(v) - 0.5)) }

// quadRun tracks the state of consecutive quads drawn with the same method.
type quadRun struct {
	color uint32
	have  bool
}

// renderQuads draws a Cmd of canonical quads. Pixel aligned solid rects use
// RenderFillRects and pixel aligned 1:1 texture copies use RenderTexture.
// Everything else goes through the geometry path in order.
func (b *Backend) renderQuads(dl *goak.DrawList, idx []uint32, tex *sdl.Texture, sx, sy float32) error {
	var firstErr error
	note := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	vs := dl.Vertices
	var texW, texH float32
	if tex != nil {
		texW, texH = float32(tex.W), float32(tex.H)
	}
	var fillColor, modColor uint32
	var haveFill, haveMod bool
	blendOff := false // opaque fills skip blending
	defer func() {
		if blendOff {
			note(b.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND))
		}
	}()
	b.rects = b.rects[:0]
	flushRects := func() {
		if len(b.rects) == 0 {
			return
		}
		note(b.renderer.RenderFillRects(b.rects))
		b.rects = b.rects[:0]
	}
	slowFrom := -1 // start of a pending run of geometry quads in idx
	resetMod := func() {
		if haveMod {
			note(tex.SetColorMod(255, 255, 255))
			note(tex.SetAlphaMod(255))
			haveMod = false
		}
	}
	flushSlow := func(end int) {
		if slowFrom >= 0 {
			resetMod()
			note(b.drawGeometry(tex, idx[slowFrom:end]))
			slowFrom = -1
		}
	}
	for q := 0; q < len(idx); q += 6 {
		i0 := int(idx[q])
		if i0+3 >= len(vs) {
			flushRects()
			flushSlow(q)
			note(errIndexRange)
			continue
		}
		v0, v2 := &vs[i0], &vs[i0+2]
		x0, ok0 := snap(v0.Pos.X * sx)
		y0, ok1 := snap(v0.Pos.Y * sy)
		x1, ok2 := snap(v2.Pos.X * sx)
		y1, ok3 := snap(v2.Pos.Y * sy)
		fast := ok0 && ok1 && ok2 && ok3 && x1 > x0 && y1 > y0
		if tex == nil && !fast {
			// Solid rects snap to the pixel grid. The software rasterizer
			// drops rows of fractional geometry quads, leaving seams
			// between rects that share an edge.
			x0, y0, x1, y1 = roundPx(v0.Pos.X*sx), roundPx(v0.Pos.Y*sy), roundPx(v2.Pos.X*sx), roundPx(v2.Pos.Y*sy)
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if y1 <= y0 {
				y1 = y0 + 1
			}
			fast = true
		}
		var src sdl.FRect
		if fast && tex != nil {
			u0, a0 := snap(v0.UV[0] * texW)
			t0, a1 := snap(v0.UV[1] * texH)
			u1, a2 := snap(v2.UV[0] * texW)
			t1, a3 := snap(v2.UV[1] * texH)
			fast = a0 && a1 && a2 && a3 && u1-u0 == x1-x0 && t1-t0 == y1-y0
			src = sdl.FRect{X: u0, Y: t0, W: u1 - u0, H: t1 - t0}
		}
		if !fast {
			b.nSlow++
			flushRects()
			if slowFrom < 0 {
				slowFrom = q
			}
			continue
		}
		flushSlow(q)
		dst := sdl.FRect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
		col := v0.Color
		if tex == nil {
			if !haveFill || col != fillColor {
				flushRects()
				c := goak.Unpack(col)
				if opaque := c.A == 255; opaque != blendOff {
					mode := sdl.BLENDMODE_BLEND
					if opaque {
						mode = sdl.BLENDMODE_NONE
					}
					note(b.renderer.SetDrawBlendMode(mode))
					blendOff = opaque
				}
				note(b.renderer.SetDrawColor(c.R, c.G, c.B, c.A))
				fillColor, haveFill = col, true
			}
			b.rects = append(b.rects, dst)
			b.nFill++
			continue
		}
		if !haveMod || col != modColor {
			c := goak.Unpack(col)
			note(tex.SetColorMod(c.R, c.G, c.B))
			note(tex.SetAlphaMod(c.A))
			modColor, haveMod = col, true
		}
		note(b.renderer.RenderTexture(tex, &src, &dst))
		b.nCopy++
	}
	flushRects()
	flushSlow(len(idx))
	resetMod()
	return firstErr
}
