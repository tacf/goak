package goak

func (w *Window) resetLayout(start Vec2, right float32) {
	w.contentMin = start
	w.cursor = start
	w.maxPos = start
	w.contentRight = right
	w.indent = 0
	w.prevLineEnd = start
	w.prevLineH, w.currLineH = 0, 0
	w.onSameLine = false
	w.itemWidths = w.itemWidths[:0]
	w.nextWidthSet = false
}

func (w *Window) lineStartX() float32 { return w.contentMin.X + w.indent }

// ItemAdd reserves a rect of the given size at the cursor and advances the
// cursor to the next line. It also grows the content extent.
func (c *Context) ItemAdd(size Vec2) Rect {
	w := c.cur
	pos := w.cursor
	r := Rect{Min: pos, Max: pos.Add(size)}
	lineH := max(w.currLineH, size.Y)
	w.prevLineEnd = Vec2{X: r.Max.X, Y: pos.Y}
	w.prevLineH = lineH
	w.currLineH = 0
	w.cursor = Vec2{X: w.lineStartX(), Y: pos.Y + lineH + c.Style.ItemSpacing.Y}
	w.maxPos = w.maxPos.Max(r.Max)
	w.onSameLine = false
	w.nextWidthSet = false
	c.lastItem = r
	return r
}

// Dummy reserves empty space.
func (c *Context) Dummy(size Vec2) { c.ItemAdd(size) }

// SameLine places the next item on the previous line. A zero offset uses
// spacing after the previous item, a negative spacing means ItemSpacing.X.
// A nonzero offset is measured from the line start.
func (c *Context) SameLine(offset, spacing float32) {
	w := c.cur
	if spacing < 0 {
		spacing = c.Style.ItemSpacing.X
	}
	x := w.prevLineEnd.X + spacing
	if offset != 0 {
		x = w.lineStartX() + offset
	}
	w.cursor = Vec2{X: x, Y: w.prevLineEnd.Y}
	w.currLineH = w.prevLineH
	w.onSameLine = true
}

// NewLine ends a SameLine run or inserts one blank line.
func (c *Context) NewLine() {
	if c.cur.onSameLine {
		c.ItemAdd(Vec2{})
		return
	}
	c.ItemAdd(Vec2{Y: c.lineHeight()})
}

// Spacing inserts vertical item spacing.
func (c *Context) Spacing() { c.ItemAdd(Vec2{}) }

// Separator draws a horizontal rule across the content width.
func (c *Context) Separator() {
	w := c.cur
	r := c.ItemAdd(Vec2{X: max(w.contentRight-w.lineStartX(), 1), Y: 1})
	if !w.skip {
		w.DL.AddLine(Vec2{X: r.Min.X, Y: r.Min.Y + 0.5}, Vec2{X: r.Max.X, Y: r.Min.Y + 0.5}, c.col(ColSeparator), 1)
	}
}

// Indent moves the line start right. Zero means Style.IndentSpacing.
func (c *Context) Indent(amount float32) {
	if amount == 0 {
		amount = c.Style.IndentSpacing
	}
	w := c.cur
	w.indent += amount
	w.cursor.X = w.lineStartX()
}

// Unindent undoes Indent.
func (c *Context) Unindent(amount float32) {
	if amount == 0 {
		amount = c.Style.IndentSpacing
	}
	c.Indent(-amount)
}

// CursorScreenPos is the cursor in screen coordinates.
func (c *Context) CursorScreenPos() Vec2 { return c.cur.cursor }

// SetCursorScreenPos moves the cursor.
func (c *Context) SetCursorScreenPos(p Vec2) { c.cur.cursor = p }

// ContentRegionAvail is the space from the cursor to the content edges.
func (c *Context) ContentRegionAvail() Vec2 {
	w := c.cur
	bottom := w.Pos.Y + w.Size.Y
	if w.Child {
		bottom -= w.childPad.Y
	} else {
		bottom -= c.Style.WindowPadding.Y
	}
	return Vec2{X: max(w.contentRight-w.cursor.X, 0), Y: max(bottom-w.cursor.Y, 0)}
}

// SetNextItemWidth sets the width for the next item only.
func (c *Context) SetNextItemWidth(v float32) {
	c.cur.nextWidth, c.cur.nextWidthSet = v, true
}

// PushItemWidth pushes a default item width. Negative means the available
// width minus that amount.
func (c *Context) PushItemWidth(v float32) { c.cur.itemWidths = append(c.cur.itemWidths, v) }

// PopItemWidth pops PushItemWidth.
func (c *Context) PopItemWidth() {
	w := c.cur
	if len(w.itemWidths) == 0 {
		panic("goak: PopItemWidth without PushItemWidth")
	}
	w.itemWidths = w.itemWidths[:len(w.itemWidths)-1]
}

// CalcItemWidth resolves the width for the next item: SetNextItemWidth,
// else the PushItemWidth top, else 65 percent of the content width.
func (c *Context) CalcItemWidth() float32 {
	w := c.cur
	var v float32
	switch {
	case w.nextWidthSet:
		v = w.nextWidth
	case len(w.itemWidths) > 0:
		v = w.itemWidths[len(w.itemWidths)-1]
	default:
		v = 0.65 * (w.contentRight - w.contentMin.X)
	}
	if v < 0 {
		v = max(w.contentRight-w.cursor.X+v, 1)
	}
	return v
}
