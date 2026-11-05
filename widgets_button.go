package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
)

// Dir is the direction of an ArrowButton.
type Dir uint8

// Arrow directions.
const (
	DirLeft Dir = iota
	DirRight
	DirUp
	DirDown
)

// SmallButton is a Button without vertical frame padding.
func (c *Context) SmallButton(label string) bool {
	old := c.Style.FramePadding.Y
	c.Style.FramePadding.Y = 0
	defer func() { c.Style.FramePadding.Y = old }()
	return c.Button(label)
}

// InvisibleButton is a clickable area of the given size with no drawing.
func (c *Context) InvisibleButton(id string, size Vec2) bool {
	wid := c.ItemID(id)
	r := c.ItemAdd(size)
	if c.cur.skip {
		return false
	}
	return c.ButtonBehaviorEx(r, wid, 0).Pressed
}

// ArrowButton is a square button with an arrow.
func (c *Context) ArrowButton(id string, dir Dir) bool {
	wid := c.ItemID(id)
	h := c.frameHeight()
	r := c.ItemAdd(Vec2{X: h, Y: h})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, wid, 0)
	dl := c.cur.DL
	col := c.frameColor(s, wid, ColButton, ColButtonHovered, ColButtonActive)
	dl.AddRectFilled(r.Min, r.Max, c.col(col), c.Style.FrameRounding)
	if c.focusID == wid {
		dl.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	}
	c.drawArrow(r, dir, c.col(ColText))
	return s.Pressed
}

// drawArrow draws a triangle pointing in dir centered in r.
func (c *Context) drawArrow(r Rect, dir Dir, col colors.Color) {
	ctr := Vec2{X: (r.Min.X + r.Max.X) / 2, Y: (r.Min.Y + r.Max.Y) / 2}
	h := min(r.Max.X-r.Min.X, r.Max.Y-r.Min.Y) * 0.25
	dl := c.cur.DL
	switch dir {
	case DirLeft:
		dl.AddTriangleFilled(Vec2{X: ctr.X - h, Y: ctr.Y}, Vec2{X: ctr.X + h, Y: ctr.Y - h}, Vec2{X: ctr.X + h, Y: ctr.Y + h}, col)
	case DirRight:
		dl.AddTriangleFilled(Vec2{X: ctr.X + h, Y: ctr.Y}, Vec2{X: ctr.X - h, Y: ctr.Y + h}, Vec2{X: ctr.X - h, Y: ctr.Y - h}, col)
	case DirUp:
		dl.AddTriangleFilled(Vec2{X: ctr.X, Y: ctr.Y - h}, Vec2{X: ctr.X - h, Y: ctr.Y + h}, Vec2{X: ctr.X + h, Y: ctr.Y + h}, col)
	default:
		dl.AddTriangleFilled(Vec2{X: ctr.X, Y: ctr.Y + h}, Vec2{X: ctr.X + h, Y: ctr.Y - h}, Vec2{X: ctr.X - h, Y: ctr.Y - h}, col)
	}
}

// Button draws a button and returns true on the frame it is clicked.
func (c *Context) Button(label string) bool {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	fp := c.Style.FramePadding
	ts := c.TextSize(text)
	r := c.ItemAdd(Vec2{X: ts.X + 2*fp.X, Y: ts.Y + 2*fp.Y})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, id, 0)
	dl := c.cur.DL
	col := c.frameColor(s, id, ColButton, ColButtonHovered, ColButtonActive)
	dl.AddRectFilled(r.Min, r.Max, c.col(col), c.Style.FrameRounding)
	if c.focusID == id {
		dl.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	}
	c.DrawText(r.Min.Add(fp), text, c.col(ColText))
	return s.Pressed
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
