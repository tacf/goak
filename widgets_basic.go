package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
)

func (c *Context) frameHeight() float32 { return c.lineHeight() + 2*c.Style.FramePadding.Y }

// RadioButton draws a radio button and returns true when clicked.
func (c *Context) RadioButton(label string, active bool) bool {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	fp := c.Style.FramePadding
	ts := c.TextSize(text)
	box := c.frameHeight()
	w := box
	if text != "" {
		w += c.Style.ItemInnerSpacing.X + ts.X
	}
	r := c.ItemAdd(Vec2{X: w, Y: box})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, id, 0)
	dl := c.cur.DL
	ctr := Vec2{X: r.Min.X + box/2, Y: r.Min.Y + box/2}
	col := c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive)
	dl.AddCircleFilled(ctr, box/2, c.col(col))
	dl.AddCircle(ctr, box/2, 1, c.col(ColBorder))
	if active {
		dl.AddCircleFilled(ctr, box*0.25, c.col(ColCheckMark))
	}
	if c.focusID == id {
		dl.AddRect(r.Min, r.Max, c.col(ColCheckMark), c.Style.FrameRounding, 1)
	}
	if text != "" {
		c.DrawText(Vec2{X: r.Min.X + box + c.Style.ItemInnerSpacing.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
	}
	return s.Pressed
}

// RadioButtonInt selects value v in *cur and returns true when it changed.
func (c *Context) RadioButtonInt(label string, cur *int, v int) bool {
	if c.RadioButton(label, cur != nil && *cur == v) {
		if cur != nil {
			*cur = v
		}
		return true
	}
	return false
}

// LabelText draws a value on the left and a label after it.
func (c *Context) LabelText(label, value string) {
	text := DisplayLabel(label)
	vs := c.TextSize(value)
	ls := c.TextSize(text)
	fp := c.Style.FramePadding
	vw := max(c.CalcItemWidth(), vs.X)
	r := c.ItemAdd(Vec2{X: vw + c.Style.ItemInnerSpacing.X + ls.X, Y: max(vs.Y, ls.Y) + 2*fp.Y})
	if c.cur.skip {
		return
	}
	c.DrawText(Vec2{X: r.Min.X, Y: r.Min.Y + fp.Y}, value, c.col(ColText))
	c.DrawText(Vec2{X: r.Min.X + vw + c.Style.ItemInnerSpacing.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
}

// BulletText draws a bullet followed by text.
func (c *Context) BulletText(s string) {
	ts := c.TextSize(s)
	lh := c.lineHeight()
	r := c.ItemAdd(Vec2{X: lh + ts.X, Y: ts.Y})
	if c.cur.skip {
		return
	}
	c.cur.DL.AddCircleFilled(Vec2{X: r.Min.X + lh*0.5, Y: r.Min.Y + lh*0.5}, lh*0.15, c.col(ColText))
	c.DrawText(Vec2{X: r.Min.X + lh, Y: r.Min.Y}, s, c.col(ColText))
}

// SeparatorText draws a rule with a label on its left.
func (c *Context) SeparatorText(label string) {
	w := c.cur
	text := DisplayLabel(label)
	ts := c.TextSize(text)
	width := max(w.contentRight-w.lineStartX(), 1)
	r := c.ItemAdd(Vec2{X: width, Y: ts.Y + c.Style.FramePadding.Y})
	if w.skip {
		return
	}
	y := r.Min.Y + ts.Y/2
	col := c.col(ColSeparator)
	gap := c.Style.ItemInnerSpacing.X
	if text == "" {
		w.DL.AddLine(Vec2{X: r.Min.X, Y: y}, Vec2{X: r.Max.X, Y: y}, col, 1)
		return
	}
	lead := float32(12)
	w.DL.AddLine(Vec2{X: r.Min.X, Y: y}, Vec2{X: r.Min.X + lead, Y: y}, col, 1)
	c.DrawText(Vec2{X: r.Min.X + lead + gap, Y: r.Min.Y}, text, c.col(ColText))
	x := r.Min.X + lead + 2*gap + ts.X
	if x < r.Max.X {
		w.DL.AddLine(Vec2{X: x, Y: y}, Vec2{X: r.Max.X, Y: y}, col, 1)
	}
}

// Image draws a texture region. Pass colors.White as tint for no tint.
func (c *Context) Image(tex draw.TextureID, size, uv0, uv1 Vec2, tint colors.Color) {
	r := c.ItemAdd(size)
	if c.cur.skip {
		return
	}
	c.cur.DL.AddImage(tex, r.Min, r.Max, [2]float32{uv0.X, uv0.Y}, [2]float32{uv1.X, uv1.Y}, tint)
}

// ImageButton is a button showing a texture region inside the frame padding.
func (c *Context) ImageButton(id string, tex draw.TextureID, size, uv0, uv1 Vec2, bg, tint colors.Color) bool {
	wid := c.ItemID(id)
	fp := c.Style.FramePadding
	r := c.ItemAdd(Vec2{X: size.X + 2*fp.X, Y: size.Y + 2*fp.Y})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, wid, 0)
	dl := c.cur.DL
	dl.AddRectFilled(r.Min, r.Max, c.col(c.frameColor(s, wid, ColButton, ColButtonHovered, ColButtonActive)), c.Style.FrameRounding)
	if bg.A != 0 {
		dl.AddRectFilled(r.Min.Add(fp), r.Max.Sub(fp), bg, 0)
	}
	dl.AddImage(tex, r.Min.Add(fp), r.Max.Sub(fp), [2]float32{uv0.X, uv0.Y}, [2]float32{uv1.X, uv1.Y}, tint)
	if c.focusID == wid {
		dl.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	}
	return s.Pressed
}

// ColorButton draws a color swatch button.
func (c *Context) ColorButton(id string, col colors.Color, size Vec2) bool {
	wid := c.ItemID(id)
	if size.X <= 0 {
		size.X = c.frameHeight()
	}
	if size.Y <= 0 {
		size.Y = c.frameHeight()
	}
	r := c.ItemAdd(size)
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, wid, 0)
	dl := c.cur.DL
	dl.AddRectFilled(r.Min, r.Max, col, c.Style.FrameRounding)
	bc := ColBorder
	if s.Hovered {
		bc = ColCheckMark
	}
	dl.AddRect(r.Min, r.Max, c.col(bc), c.Style.FrameRounding, 1)
	return s.Pressed
}

// ProgressBar draws frac in 0..1. A size component of 0 means the available
// width or the frame height. An empty overlay shows no text.
func (c *Context) ProgressBar(frac float32, size Vec2, overlay string) {
	if size.X <= 0 {
		size.X = max(c.ContentRegionAvail().X+size.X, 1)
	}
	if size.Y <= 0 {
		size.Y = c.frameHeight()
	}
	r := c.ItemAdd(size)
	if c.cur.skip {
		return
	}
	frac = min(max(frac, 0), 1)
	dl := c.cur.DL
	dl.AddRectFilled(r.Min, r.Max, c.col(ColFrameBg), c.Style.FrameRounding)
	if frac > 0 {
		dl.PushClipRect(Rect{Min: r.Min, Max: Vec2{X: r.Min.X + size.X*frac, Y: r.Max.Y}})
		dl.AddRectFilled(r.Min, r.Max, c.col(ColProgressBar), c.Style.FrameRounding)
		dl.PopClipRect()
	}
	dl.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	if overlay != "" {
		ts := c.TextSize(overlay)
		c.DrawText(Vec2{X: r.Min.X + (size.X-ts.X)/2, Y: r.Min.Y + (size.Y-ts.Y)/2}, overlay, c.col(ColText))
	}
}
