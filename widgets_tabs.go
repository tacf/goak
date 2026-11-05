package goak

import "github.com/tacf/goak/internal/draw"

type tabBarState struct {
	id      ID
	rect    Rect
	x       float32
	persist ID // selection kept across frames
	sel     ID // selection whose content shows this frame
	first   ID
	seen    bool
	parent  *tabBarState
}

// BeginTabBar starts a row of tabs. Pair a true result with EndTabBar.
// The selected tab persists per bar. Switching shows the new content one
// frame after the click.
func (c *Context) BeginTabBar(id string) bool {
	w := c.cur
	if w == nil {
		return false
	}
	bid := c.GetID(id)
	c.PushID(id)
	skey := hashStr("#TABBAR", bid)
	bar := StateGet[*tabBarState](c.States, skey, nil)
	if bar == nil {
		bar = &tabBarState{id: bid}
	}
	StateSetKeep(c.States, skey, bar, LongAge)
	r := c.ItemAdd(Vec2{X: max(c.ContentRegionAvail().X, 1), Y: c.frameHeight()})
	bar.rect, bar.x = r, r.Min.X
	bar.sel, bar.first, bar.seen = bar.persist, 0, false
	bar.parent = c.tabBar
	c.tabBar = bar
	if !w.skip {
		w.DL.AddLine(Vec2{X: r.Min.X, Y: r.Max.Y - 0.5}, Vec2{X: r.Max.X, Y: r.Max.Y - 0.5}, c.col(ColSeparator), 1)
	}
	return true
}

// EndTabBar ends the tab row.
func (c *Context) EndTabBar() {
	bar := c.tabBar
	if bar == nil {
		panic("goak: EndTabBar without BeginTabBar")
	}
	if !bar.seen && !c.cur.skip {
		bar.persist = 0
	}
	c.tabBar = bar.parent
	c.PopID()
}

// BeginTabItem adds a tab. When open is not nil the tab shows a close
// button that clears *open, and a false *open hides the tab. It returns true
// for the selected tab, whose content follows. Pair a true result with
// EndTabItem.
func (c *Context) BeginTabItem(label string, open *bool) bool {
	bar := c.tabBar
	if bar == nil {
		panic("goak: BeginTabItem outside a tab bar")
	}
	id := c.ItemID(label)
	if open != nil && !*open {
		if bar.sel == id {
			bar.sel, bar.persist = 0, 0
		}
		return false
	}
	text := DisplayLabel(label)
	st := &c.Style
	fp := st.FramePadding
	ts := c.TextSize(text)
	lh := c.lineHeight()
	closeW := float32(0)
	if open != nil {
		closeW = lh*0.6 + st.ItemInnerSpacing.X
	}
	width := ts.X + 2*fp.X + closeW
	r := Rect{Min: Vec2{X: bar.x, Y: bar.rect.Min.Y}, Max: Vec2{X: bar.x + width, Y: bar.rect.Max.Y}}
	bar.x += width + 1
	c.lastItem = r
	if bar.first == 0 {
		bar.first = id
	}
	if bar.sel == 0 {
		bar.sel, bar.persist = id, id
	}
	if c.cur.skip {
		return false
	}
	var closed, closeHot bool
	var cr Rect
	if open != nil {
		cs := lh * 0.6
		cr = Rect{Min: Vec2{X: r.Max.X - fp.X/2 - cs, Y: (r.Min.Y+r.Max.Y)/2 - cs/2}, Max: Vec2{X: r.Max.X - fp.X/2, Y: (r.Min.Y+r.Max.Y)/2 + cs/2}}
		cs2 := c.ButtonBehaviorEx(cr, hashStr("#close", id), BtnNoFocus)
		closeHot = cs2.Hovered
		if cs2.Pressed {
			*open = false
			closed = true
		}
	}
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	if s.Started {
		bar.persist = id
	}
	shown := bar.sel == id
	if shown {
		bar.seen = true
	}
	col := ColTab
	switch {
	case bar.persist == id:
		col = ColTabActive
	case s.Hovered:
		col = ColTabHovered
	}
	dl := c.cur.DL
	dl.AddRectFilledEx(r.Min, r.Max, c.col(col), st.TabRounding, draw.CornerTopLeft|draw.CornerTopRight)
	c.DrawText(Vec2{X: r.Min.X + fp.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
	if open != nil {
		if closeHot {
			dl.AddRectFilled(cr.Min, cr.Max, c.col(ColCloseHovered), st.FrameRounding)
		}
		m := float32(3)
		tc := c.col(ColText)
		dl.AddLine(Vec2{X: cr.Min.X + m, Y: cr.Min.Y + m}, Vec2{X: cr.Max.X - m, Y: cr.Max.Y - m}, tc, 1.2)
		dl.AddLine(Vec2{X: cr.Max.X - m, Y: cr.Min.Y + m}, Vec2{X: cr.Min.X + m, Y: cr.Max.Y - m}, tc, 1.2)
	}
	if closed && shown {
		bar.sel, bar.persist = 0, 0
	}
	if closed || !shown {
		return false
	}
	c.PushID(uint32(id))
	return true
}

// EndTabItem ends the content of the selected tab.
func (c *Context) EndTabItem() { c.PopID() }
