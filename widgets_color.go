package goak

import (
	"fmt"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/colorx"
)

func clamp01(v float32) float32 { return min(max(v, 0), 1) }

func toU8(v float32) uint8 { return uint8(clamp01(v)*255 + 0.5) }

func rgbaColor(c [4]float32) colors.Color {
	return colors.RGBA(toU8(c[0]), toU8(c[1]), toU8(c[2]), toU8(c[3]))
}

// colorPickState keeps the hue and saturation between frames. They cannot be
// recovered from RGB when the color is gray or black.
type colorPickState struct {
	h, s, v float32
	rgb     [3]float32
	set     bool
}

// ColorEdit3 edits an RGB color as a row of 0..255 fields and a swatch that
// opens a ColorPicker. Components are in 0..1.
func (c *Context) ColorEdit3(label string, col *[3]float32) bool {
	if col == nil {
		return false
	}
	a := [4]float32{col[0], col[1], col[2], 1}
	changed := c.colorEdit(label, &a, false)
	col[0], col[1], col[2] = a[0], a[1], a[2]
	return changed
}

// ColorEdit4 is ColorEdit3 with an alpha component.
func (c *Context) ColorEdit4(label string, col *[4]float32) bool {
	if col == nil {
		return false
	}
	return c.colorEdit(label, col, true)
}

func (c *Context) colorEdit(label string, col *[4]float32, alpha bool) bool {
	n := 3
	if alpha {
		n = 4
	}
	text := DisplayLabel(label)
	gap := c.Style.ItemInnerSpacing.X
	sw := c.frameHeight()
	w := max((c.CalcItemWidth()-sw-gap*float32(n))/float32(n), 1)
	c.PushID(label)
	c.BeginGroup()
	changed := false
	for i := 0; i < n; i++ {
		if i > 0 {
			c.SameLine(0, gap)
		}
		v := int(toU8(col[i]))
		c.SetNextItemWidth(w)
		if c.DragInt(compID(i), &v, 0.5, 0, 255, "") {
			col[i], changed = float32(v)/255, true
		}
	}
	c.SameLine(0, gap)
	if c.ColorButton("##swatch", rgbaColor(*col), Vec2{X: sw, Y: sw}) {
		c.OpenPopup("picker")
	}
	if c.BeginPopup("picker") {
		if c.colorPicker("##pick", col, alpha) {
			changed = true
		}
		c.EndPopup()
	}
	if text != "" {
		c.SameLine(0, gap)
		c.labelAfter(text)
	}
	c.EndGroup()
	c.PopID()
	return changed
}

// ColorPicker3 is a saturation and value square with a hue bar.
func (c *Context) ColorPicker3(label string, col *[3]float32) bool {
	if col == nil {
		return false
	}
	a := [4]float32{col[0], col[1], col[2], 1}
	changed := c.colorPicker(label, &a, false)
	col[0], col[1], col[2] = a[0], a[1], a[2]
	return changed
}

// ColorPicker4 is ColorPicker3 with an alpha bar.
func (c *Context) ColorPicker4(label string, col *[4]float32) bool {
	if col == nil {
		return false
	}
	return c.colorPicker(label, col, true)
}

func opaque(r, g, b float32) colors.Color { return colors.RGB(toU8(r), toU8(g), toU8(b)) }

// bar runs a vertical drag bar and returns the 0..1 position from the top.
func (c *Context) bar(r Rect, id ID) (t float32, held bool) {
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	if c.activeID == id && s.Held {
		return clamp01((c.mouse().Y - r.Min.Y) / (r.Max.Y - r.Min.Y)), true
	}
	return 0, false
}

func (c *Context) colorPicker(label string, col *[4]float32, alpha bool) bool {
	key := hashStr("#COLORPICK", c.GetID(label))
	st := StateGet(c.States, key, colorPickState{})
	cur := [3]float32{col[0], col[1], col[2]}
	if !st.set || st.rgb != cur {
		h, s, v := colorx.RGBToHSV(cur[0], cur[1], cur[2])
		if st.set && (s == 0 || v == 0) {
			h = st.h
		}
		st.h, st.s, st.v, st.rgb, st.set = h, s, v, cur, true
	}
	c.PushID(label)
	defer c.PopID()
	const size = float32(180)
	barW := float32(18)
	gap := c.Style.ItemInnerSpacing.X
	dl := c.cur.DL
	changed := false

	// Saturation and value square.
	sv := c.ItemAdd(Vec2{X: size, Y: size})
	if !c.cur.skip {
		if t, held := c.bar(sv, c.ItemID("##sv")); held {
			st.v = 1 - t
			st.s = clamp01((c.mouse().X - sv.Min.X) / size)
			changed = true
		}
		hr, hg, hb := colorx.HSVToRGB(st.h, 1, 1)
		hue := opaque(hr, hg, hb)
		clear := colors.RGBA(0, 0, 0, 0)
		dl.AddRectFilledMultiColor(sv.Min, sv.Max, colors.White, hue, hue, colors.White)
		dl.AddRectFilledMultiColor(sv.Min, sv.Max, clear, clear, colors.Black, colors.Black)
		dl.AddRect(sv.Min, sv.Max, c.col(ColBorder), 0, 1)
		p := Vec2{X: sv.Min.X + st.s*size, Y: sv.Min.Y + (1-st.v)*size}
		dl.AddCircle(p, 5, 2, colors.Black)
		dl.AddCircle(p, 5, 1, colors.White)
	}

	// Hue bar.
	c.SameLine(0, gap)
	hb := c.ItemAdd(Vec2{X: barW, Y: size})
	if !c.cur.skip {
		if t, held := c.bar(hb, c.ItemID("##hue")); held {
			st.h, changed = t, true
		}
		for k := 0; k < 6; k++ {
			r0, g0, b0 := colorx.HSVToRGB(float32(k)/6, 1, 1)
			r1, g1, b1 := colorx.HSVToRGB(float32(k+1)/6, 1, 1)
			a, b := opaque(r0, g0, b0), opaque(r1, g1, b1)
			y0, y1 := hb.Min.Y+size*float32(k)/6, hb.Min.Y+size*float32(k+1)/6
			dl.AddRectFilledMultiColor(Vec2{X: hb.Min.X, Y: y0}, Vec2{X: hb.Max.X, Y: y1}, a, a, b, b)
		}
		dl.AddRect(hb.Min, hb.Max, c.col(ColBorder), 0, 1)
		y := hb.Min.Y + st.h*size
		dl.AddRect(Vec2{X: hb.Min.X - 2, Y: y - 2}, Vec2{X: hb.Max.X + 2, Y: y + 2}, colors.White, 0, 1)
	}

	// Alpha bar.
	if alpha {
		c.SameLine(0, gap)
		ab := c.ItemAdd(Vec2{X: barW, Y: size})
		if !c.cur.skip {
			if t, held := c.bar(ab, c.ItemID("##alpha")); held {
				col[3], changed = 1-t, true
			}
			c.checkerboard(ab, 6)
			r, g, b := colorx.HSVToRGB(st.h, st.s, st.v)
			top := opaque(r, g, b)
			dl.AddRectFilledMultiColor(ab.Min, ab.Max, top, top, colors.RGBA(top.R, top.G, top.B, 0), colors.RGBA(top.R, top.G, top.B, 0))
			dl.AddRect(ab.Min, ab.Max, c.col(ColBorder), 0, 1)
			y := ab.Min.Y + (1-clamp01(col[3]))*size
			dl.AddRect(Vec2{X: ab.Min.X - 2, Y: y - 2}, Vec2{X: ab.Max.X + 2, Y: y + 2}, colors.White, 0, 1)
		}
	}

	if changed {
		col[0], col[1], col[2] = colorx.HSVToRGB(st.h, st.s, st.v)
		st.rgb = [3]float32{col[0], col[1], col[2]}
	}
	StateSet(c.States, key, st)
	c.Text(fmt.Sprintf("#%02X%02X%02X%02X", toU8(col[0]), toU8(col[1]), toU8(col[2]), toU8(col[3])))
	return changed
}

// checkerboard fills r with a gray check pattern, used behind transparency.
func (c *Context) checkerboard(r Rect, cell float32) {
	dl := c.cur.DL
	dl.PushClipRect(r.Intersect(c.cur.clip))
	light, dark := colors.RGB(200, 200, 200), colors.RGB(120, 120, 120)
	dl.AddRectFilled(r.Min, r.Max, light, 0)
	for y, row := r.Min.Y, 0; y < r.Max.Y; y, row = y+cell, row+1 {
		for x, col := r.Min.X, 0; x < r.Max.X; x, col = x+cell, col+1 {
			if (row+col)%2 == 1 {
				dl.AddRectFilled(Vec2{X: x, Y: y}, Vec2{X: x + cell, Y: y + cell}, dark, 0)
			}
		}
	}
	dl.PopClipRect()
}
