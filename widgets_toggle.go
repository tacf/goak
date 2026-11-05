package goak

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
