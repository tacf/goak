package goak

// BeginMenuBar starts the menu bar row of a window created with the MenuBar
// flag. Pair a true result with EndMenuBar.
func (c *Context) BeginMenuBar() bool {
	w := c.cur
	if w == nil || w.Child || w.skip || w.menuH == 0 || w.inMenu {
		return false
	}
	top := w.Pos.Y + w.titleH
	w.menuRect = Rect{Min: Vec2{X: w.Pos.X, Y: top}, Max: Vec2{X: w.Pos.X + w.Size.X, Y: top + w.menuH}}
	w.menuX = w.menuRect.Min.X + 2
	w.inMenu = true
	w.DL.AddRectFilled(w.menuRect.Min, w.menuRect.Max, c.col(ColMenuBarBg), 0)
	return true
}

// EndMenuBar ends the menu bar row.
func (c *Context) EndMenuBar() {
	if c.cur != nil {
		c.cur.inMenu = false
	}
}

// BeginMainMenuBar starts a full width menu bar at the top of the display.
// It must be called outside any window, and always pair it with
// EndMainMenuBar. The bar draws above ordinary windows.
func (c *Context) BeginMainMenuBar() bool {
	h := c.frameHeight()
	c.SetNextWindowPos(Vec2{}, CondAlways)
	c.SetNextWindowSize(Vec2{X: c.DisplaySize.X, Y: h}, CondAlways)
	c.noMinSize = true
	fl := NoTitleBar | NoResize | NoMove | NoBackground | MenuBar
	ok := c.Begin("##MainMenuBar", nil, fl)
	w := c.cur
	w.top, w.topSeq = true, mainBarSeq
	if ok {
		c.BeginMenuBar()
	}
	return ok
}

// EndMainMenuBar ends the main menu bar.
func (c *Context) EndMainMenuBar() {
	c.EndMenuBar()
	c.End()
}

// barItemRect reserves width in the menu bar row.
func (c *Context) barItemRect(width float32) Rect {
	w := c.cur
	r := Rect{Min: Vec2{X: w.menuX, Y: w.menuRect.Min.Y}, Max: Vec2{X: w.menuX + width, Y: w.menuRect.Max.Y}}
	w.menuX += width
	c.lastItem = r
	return r
}

// menuRow reserves a popup menu row wide enough for need.
func (c *Context) menuRow(textH, need float32) Rect {
	w := max(c.ContentRegionAvail().X, need)
	return c.ItemAdd(Vec2{X: w, Y: textH + c.Style.FramePadding.Y})
}

func (c *Context) menuHighlight(r Rect, s ButtonState, id ID, open bool) {
	dl := c.cur.DL
	switch {
	case s.Held && c.activeID == id:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderActive), c.Style.FrameRounding)
	case s.Hovered:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderHovered), c.Style.FrameRounding)
	case open:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeader), c.Style.FrameRounding)
	}
}

// closeMenuChildren closes a submenu opened from the current popup level
// when the mouse moves to a plain item.
func (c *Context) closeMenuChildren() {
	lv := c.popupLevel()
	if lv < len(c.popups) && c.popups[lv].menu {
		c.closePopupsFrom(lv)
	}
}

// menuCheckCol returns the width of the check column of the current popup
// menu. It is lineHeight when last frame had a checkable item, else 0.
func (c *Context) menuCheckCol(w *Window) float32 {
	if w.menuFrame != c.Frame {
		w.menuCheck, w.menuCheckNext, w.menuFrame = w.menuCheckNext, false, c.Frame
	}
	if w.menuCheck {
		return c.lineHeight()
	}
	return 0
}

// MenuItem draws a menu entry with an optional shortcut hint and check
// mark. It returns true when clicked and closes all popups. The check
// column is reserved only when some item of the menu is selected or made
// with MenuItemCheck (one frame lag).
func (c *Context) MenuItem(label, shortcut string, selected bool) bool {
	return c.menuItem(label, shortcut, selected, selected)
}

// MenuItemCheck draws a checkable menu entry that toggles *v on click. It
// always reserves the check column so text does not shift when toggled.
func (c *Context) MenuItemCheck(label, shortcut string, v *bool) bool {
	sel := v != nil && *v
	if c.menuItem(label, shortcut, sel, true) {
		if v != nil {
			*v = !*v
		}
		return true
	}
	return false
}

func (c *Context) menuItem(label, shortcut string, selected, checkable bool) bool {
	w := c.cur
	id := c.ItemID(label)
	text := DisplayLabel(label)
	st := &c.Style
	ts := c.TextSize(text)
	if w.inMenu {
		r := c.barItemRect(ts.X + 2*st.FramePadding.X)
		s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
		c.menuHighlight(r, s, id, false)
		c.DrawText(Vec2{X: r.Min.X + st.FramePadding.X, Y: r.Min.Y + st.FramePadding.Y}, text, c.col(ColText))
		return s.Pressed
	}
	lh := c.lineHeight()
	col := c.menuCheckCol(w)
	if checkable {
		w.menuCheckNext = true
	}
	ss := c.TextSize(shortcut)
	need := col + ts.X + 2*st.ItemInnerSpacing.X
	if shortcut != "" {
		need += 20 + ss.X
	}
	r := c.menuRow(ts.Y, need)
	if w.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	if s.Hovered {
		c.closeMenuChildren()
	}
	c.menuHighlight(r, s, id, false)
	dl := w.DL
	ty := r.Min.Y + st.FramePadding.Y/2
	if selected {
		m := lh * 0.28
		cx, cy := r.Min.X+st.ItemInnerSpacing.X+lh/2, r.Min.Y+(r.Max.Y-r.Min.Y)/2
		dl.AddLine(Vec2{X: cx - m, Y: cy}, Vec2{X: cx - m/3, Y: cy + m*0.8}, c.col(ColCheckMark), 2)
		dl.AddLine(Vec2{X: cx - m/3, Y: cy + m*0.8}, Vec2{X: cx + m, Y: cy - m*0.8}, c.col(ColCheckMark), 2)
	}
	c.DrawText(Vec2{X: r.Min.X + st.ItemInnerSpacing.X + col, Y: ty}, text, c.col(ColText))
	if shortcut != "" {
		c.DrawText(Vec2{X: r.Max.X - st.ItemInnerSpacing.X - ss.X, Y: ty}, shortcut, c.col(ColTextDisabled))
	}
	if s.Pressed {
		c.closePopupsFrom(0)
	}
	return s.Pressed
}

// BeginMenu draws a menu header in a menu bar or a submenu entry in a popup
// and returns true while its popup is open. Pair a true result with
// EndMenu. Hovering another top menu while one is open switches to it.
func (c *Context) BeginMenu(label string) bool {
	w := c.cur
	id := c.ItemID(label)
	pid := hashStr("#MENU", id)
	text := DisplayLabel(label)
	st := &c.Style
	ts := c.TextSize(text)
	inBar := w.inMenu
	var r Rect
	var col float32
	if inBar {
		r = c.barItemRect(ts.X + 2*st.FramePadding.X)
	} else {
		col = c.menuCheckCol(w)
		r = c.menuRow(ts.Y, ts.X+col+c.lineHeight()+2*st.ItemInnerSpacing.X)
		if w.skip {
			return false
		}
	}
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	pos := Vec2{X: r.Max.X, Y: r.Min.Y}
	if inBar {
		pos = Vec2{X: r.Min.X, Y: r.Max.Y}
	}
	open := c.popupIndex(pid) >= 0
	switch {
	case inBar && s.Started && !c.closedThisFrame(pid):
		c.openPopupID(pid, pos, true, true)
	case inBar && s.Hovered && !open && len(c.popups) > 0 && c.popups[0].menu:
		c.openPopupID(pid, pos, true, false)
	case !inBar && s.Hovered && !open:
		c.openPopupID(pid, pos, true, false)
	}
	open = c.popupIndex(pid) >= 0
	c.menuHighlight(r, s, id, open)
	dl := w.DL
	if inBar {
		c.DrawText(Vec2{X: r.Min.X + st.FramePadding.X, Y: r.Min.Y + st.FramePadding.Y}, text, c.col(ColText))
	} else {
		lh := c.lineHeight()
		c.DrawText(Vec2{X: r.Min.X + st.ItemInnerSpacing.X + col, Y: r.Min.Y + st.FramePadding.Y/2}, text, c.col(ColText))
		cx, cy := r.Max.X-lh*0.6, (r.Min.Y+r.Max.Y)/2
		m := lh * 0.22
		dl.AddTriangleFilled(Vec2{X: cx - m, Y: cy - m*1.3}, Vec2{X: cx - m, Y: cy + m*1.3}, Vec2{X: cx + m, Y: cy}, c.col(ColText))
	}
	if !open {
		return false
	}
	c.popups[c.popupIndex(pid)].pos = pos
	return c.beginPopupID(pid, label, 0)
}

// EndMenu ends a menu popup.
func (c *Context) EndMenu() { c.EndPopup() }
