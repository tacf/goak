package goak

// TreeFlags tune TreeNodeEx.
type TreeFlags uint8

// Tree node flags.
const (
	// TreeLeaf has no arrow and cannot be toggled. It still needs TreePop.
	TreeLeaf TreeFlags = 1 << iota
	// TreeDefaultOpen opens the node the first time it is seen.
	TreeDefaultOpen
	// TreeSelected highlights the node.
	TreeSelected
)

// SetNextItemOpen sets the open state of the next TreeNode or
// CollapsingHeader. CondOnce applies it only when the node has no state yet.
func (c *Context) SetNextItemOpen(open bool, cond Cond) {
	c.nextOpen, c.nextOpenOn, c.nextOpenCd = open, true, cond
}

// TreeNode draws a node with an arrow. When it returns true, submit the
// children and call TreePop. The open state persists per ID.
func (c *Context) TreeNode(label string) bool { return c.TreeNodeEx(label, 0) }

// TreeNodeEx is TreeNode with flags.
func (c *Context) TreeNodeEx(label string, fl TreeFlags) bool {
	open, leaf := c.treeRow(label, fl, false)
	if !open && !leaf {
		return false
	}
	c.Indent(0)
	c.PushID(label)
	return true
}

// TreePop ends the children of an open TreeNode.
func (c *Context) TreePop() {
	c.PopID()
	c.Unindent(0)
}

// CollapsingHeader draws a framed header and returns whether it is open.
// There is no TreePop.
func (c *Context) CollapsingHeader(label string) bool {
	open, _ := c.treeRow(label, 0, true)
	return open
}

// treeRow handles state, interaction and drawing of tree rows.
func (c *Context) treeRow(label string, fl TreeFlags, framed bool) (open, leaf bool) {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	st := &c.Style
	ts := c.TextSize(text)
	lh := c.lineHeight()
	leaf = fl&TreeLeaf != 0
	skey := hashStr("#TREE", id)
	v := StateGet[int8](c.States, skey, 0)
	open = v == 2
	if v == 0 {
		open = fl&TreeDefaultOpen != 0
	}
	if c.nextOpenOn {
		if c.nextOpenCd == CondAlways || v == 0 {
			open = c.nextOpen
		}
		c.nextOpenOn = false
	}
	w := max(c.ContentRegionAvail().X, lh+ts.X)
	r := c.ItemAdd(Vec2{X: w, Y: ts.Y + st.FramePadding.Y})
	if c.cur.skip {
		return false, false
	}
	s := c.ButtonBehaviorEx(r, id, 0)
	if s.Pressed && !leaf {
		open = !open
	}
	state := int8(1)
	if open {
		state = 2
	}
	StateSetKeep(c.States, skey, state, LongAge)

	dl := c.cur.DL
	switch {
	case s.Held && c.activeID == id:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderActive), st.FrameRounding)
	case s.Hovered:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderHovered), st.FrameRounding)
	case framed || fl&TreeSelected != 0:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeader), st.FrameRounding)
	}
	if c.focusID == id {
		dl.AddRect(r.Min, r.Max, c.col(ColBorder), st.FrameRounding, 1)
	}
	cx, cy := r.Min.X+lh*0.5+2, (r.Min.Y+r.Max.Y)/2
	if !leaf {
		m := lh * 0.22
		if open {
			dl.AddTriangleFilled(Vec2{X: cx - m*1.3, Y: cy - m}, Vec2{X: cx + m*1.3, Y: cy - m}, Vec2{X: cx, Y: cy + m*1.2}, c.col(ColText))
		} else {
			dl.AddTriangleFilled(Vec2{X: cx - m, Y: cy - m*1.3}, Vec2{X: cx - m, Y: cy + m*1.3}, Vec2{X: cx + m*1.2, Y: cy}, c.col(ColText))
		}
	}
	c.DrawText(Vec2{X: r.Min.X + lh + 2, Y: r.Min.Y + st.FramePadding.Y/2}, text, c.col(ColText))
	return open, leaf
}
