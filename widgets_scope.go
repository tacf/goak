package goak

import "github.com/tacf/goak/colors"

type groupState struct {
	start   Vec2
	maxPos  Vec2
	sameRow bool
}

// BeginGroup starts a group. After EndGroup the group acts as one item, so
// SameLine and layout treat it as a single block.
func (c *Context) BeginGroup() {
	w := c.cur
	c.groups = append(c.groups, groupState{start: w.cursor, maxPos: w.maxPos, sameRow: w.onSameLine})
	w.maxPos = w.cursor
}

// EndGroup ends BeginGroup.
func (c *Context) EndGroup() {
	if len(c.groups) == 0 {
		panic("goak: EndGroup without BeginGroup")
	}
	w := c.cur
	g := c.groups[len(c.groups)-1]
	c.groups = c.groups[:len(c.groups)-1]
	ext := w.maxPos.Sub(g.start).Max(Vec2{})
	w.maxPos = w.maxPos.Max(g.maxPos)
	w.cursor = g.start
	if g.sameRow {
		w.onSameLine = true
	}
	c.ItemAdd(ext)
}

// BeginDisabled draws and treats the following widgets as disabled until
// the matching EndDisabled. Calls nest.
func (c *Context) BeginDisabled() { c.disabled++ }

// EndDisabled ends BeginDisabled.
func (c *Context) EndDisabled() {
	if c.disabled == 0 {
		panic("goak: EndDisabled without BeginDisabled")
	}
	c.disabled--
}

type styleColor struct {
	idx Col
	old colors.Color
}

// PushStyleColor overrides a style color until PopStyleColor.
func (c *Context) PushStyleColor(idx Col, col colors.Color) {
	c.styleCols = append(c.styleCols, styleColor{idx, c.Style.Colors[idx]})
	c.Style.Colors[idx] = col
}

// PopStyleColor restores the n most recently pushed style colors.
func (c *Context) PopStyleColor(n int) {
	for ; n > 0 && len(c.styleCols) > 0; n-- {
		s := c.styleCols[len(c.styleCols)-1]
		c.styleCols = c.styleCols[:len(c.styleCols)-1]
		c.Style.Colors[s.idx] = s.old
	}
}
