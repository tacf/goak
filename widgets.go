package goak

import "fmt"

// Text draws a possibly multi-line label.
func (c *Context) Text(s string) {
	sz := c.TextSize(s)
	r := c.ItemAdd(sz)
	if c.cur.skip {
		return
	}
	c.DrawText(r.Min, s, c.col(ColText))
}

// Textf is Text with formatting.
func (c *Context) Textf(format string, a ...any) { c.Text(fmt.Sprintf(format, a...)) }

func (c *Context) frameColor(s ButtonState, id ID, base, hov, act Col) Col {
	switch {
	case s.Held && c.activeID == id:
		return act
	case s.Hovered:
		return hov
	}
	return base
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

// Checkbox draws a check box and toggles *v on click. It returns true when
// the value changed.
func (c *Context) Checkbox(label string, v *bool) bool {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	fp := c.Style.FramePadding
	ts := c.TextSize(text)
	box := c.lineHeight() + 2*fp.Y
	w := box
	if text != "" {
		w += c.Style.ItemInnerSpacing.X + ts.X
	}
	r := c.ItemAdd(Vec2{X: w, Y: box})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, id, 0)
	changed := false
	if s.Pressed && v != nil {
		*v = !*v
		changed = true
	}
	dl := c.cur.DL
	br := Rect{Min: r.Min, Max: Vec2{X: r.Min.X + box, Y: r.Min.Y + box}}
	col := c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive)
	dl.AddRectFilled(br.Min, br.Max, c.col(col), c.Style.FrameRounding)
	dl.AddRect(br.Min, br.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	if c.focusID == id {
		dl.AddRect(r.Min, r.Max, c.col(ColCheckMark), c.Style.FrameRounding, 1)
	}
	if v != nil && *v {
		m := box * 0.25
		dl.AddRectFilled(Vec2{X: br.Min.X + m, Y: br.Min.Y + m}, Vec2{X: br.Max.X - m, Y: br.Max.Y - m}, c.col(ColCheckMark), c.Style.FrameRounding)
	}
	if text != "" {
		c.DrawText(Vec2{X: br.Max.X + c.Style.ItemInnerSpacing.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
	}
	return changed
}
