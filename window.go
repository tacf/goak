package goak

import (
	"github.com/tacf/goak/internal/draw"
)

// WindowFlags tune Begin and BeginChild.
type WindowFlags uint32

// Window flags.
const (
	NoTitleBar WindowFlags = 1 << iota
	NoResize
	NoMove
	NoBackground
	AutoSize
	NoScrollbar // children only
	// MenuBar reserves a menu bar row under the title for BeginMenuBar.
	MenuBar
)

// Window is a top level window or a child region.
type Window struct {
	ID    ID
	Name  string
	Flags WindowFlags
	Pos   Vec2
	Size  Vec2
	// ContentSize is the extent of the items of the previous frame.
	ContentSize Vec2
	DL          *draw.DrawList
	Parent      *Window
	Root        *Window
	Child       bool

	lastFrame int
	seenFrame int // last frame Begin or BeginChild ran, used for pruning
	idDepth   int // ID stack depth right after the window ID was pushed
	skip      bool
	hasSize   bool
	titleH    float32
	menuH     float32

	// Popup menu check column: lagged flag, current frame flag, stamp.
	menuCheck, menuCheckNext bool
	menuFrame                int

	// collapsed is toggled by a title bar double click.
	collapsed bool
	fullH     float32

	// Overlay windows (popups and tooltips) draw above all others.
	top       bool
	noInputs  bool
	topSeq    int
	popupID   ID
	savedIDs  []ID
	idBuf     []ID
	savedLast Rect
	savedID   ID
	minW      float32
	inMenu    bool
	labelText string
	menuX     float32
	menuRect  Rect

	clip         Rect
	cursor       Vec2
	prevLineEnd  Vec2
	prevLineH    float32
	currLineH    float32
	indent       float32
	contentMin   Vec2
	contentRight float32
	maxPos       Vec2
	onSameLine   bool
	itemWidths   []float32
	nextWidth    float32
	nextWidthSet bool

	scrollY  float32
	needBar  bool
	childPad Vec2
	outer    Rect
}

func (w *Window) rect() Rect { return Rect{Min: w.Pos, Max: w.Pos.Add(w.Size)} }

// SetNextWindowPos sets the position for the next Begin.
func (c *Context) SetNextWindowPos(p Vec2, cond Cond) {
	c.nextPos, c.nextPosSet, c.nextPosCond = p, true, cond
}

// SetNextWindowSize sets the size for the next Begin.
func (c *Context) SetNextWindowSize(s Vec2, cond Cond) {
	c.nextSize, c.nextSizeSet, c.nextSizeCnd = s, true, cond
}

// WindowByName returns a top level window by title or nil.
func (c *Context) WindowByName(title string) *Window { return c.windows[hashLabel(title, 0)] }

// Begin starts a top level window. Always pair with End, even when it
// returns false. It returns false when the window is closed through open.
// Calling Begin twice with the same title in one frame is not supported.
func (c *Context) Begin(title string, open *bool, flags WindowFlags) bool {
	if c.cur != nil {
		panic("goak: nested Begin, use BeginChild")
	}
	id := hashLabel(title, 0)
	w := c.windows[id]
	created := w == nil
	if created {
		off := float32(24 * (c.created % 10))
		c.created++
		w = &Window{ID: id, Name: title, DL: draw.NewDrawList(), Pos: Vec2{X: 60 + off, Y: 60 + off}, Size: Vec2{X: 240, Y: 160}}
		w.Root = w
		c.windows[id] = w
		c.order = append(c.order, w)
	}
	if c.nextPosSet && (c.nextPosCond == CondAlways || created) {
		w.Pos = c.nextPos
	}
	if c.nextSizeSet && (c.nextSizeCnd == CondAlways || created) {
		w.Size = c.nextSize
	}
	c.nextPosSet, c.nextSizeSet = false, false
	w.Flags = flags
	c.stack = append(c.stack, w)
	c.cur = w
	c.idStack = append(c.idStack, id)
	w.idDepth = len(c.idStack)
	w.seenFrame = c.Frame
	w.DL.Reset()
	w.DL.Scale = c.pixelScale()

	st := &c.Style
	if open != nil && !*open {
		w.skip = true
		w.clip = Rect{}
		return false
	}
	w.skip = false
	w.lastFrame = c.Frame
	w.titleH, w.menuH = 0, 0
	if flags&NoTitleBar == 0 {
		w.titleH = c.lineHeight() + 2*st.FramePadding.Y
	} else {
		w.collapsed = false
	}
	if flags&MenuBar != 0 {
		w.menuH = c.lineHeight() + 2*st.FramePadding.Y
	}
	w.inMenu = false
	pad := st.WindowPadding
	if flags&AutoSize != 0 {
		if w.hasSize {
			w.Size = w.ContentSize.Add(Vec2{X: 2 * pad.X, Y: 2*pad.Y + w.titleH + w.menuH})
		} else {
			w.Size = st.WindowMinSize
		}
	}
	if c.noMinSize {
		w.Size = w.Size.Max(Vec2{X: st.WindowMinSize.X})
		c.noMinSize = false
	} else {
		w.Size = w.Size.Max(st.WindowMinSize)
	}
	if w.collapsed {
		w.Size.Y = w.titleH
	}

	mouse := c.mouse()
	// Resize grip.
	if flags&(NoResize|AutoSize) == 0 && !w.collapsed {
		g := st.ResizeGripSize
		r := w.rect()
		grip := Rect{Min: Vec2{X: r.Max.X - g, Y: r.Max.Y - g}, Max: r.Max}
		w.clip = r
		s := c.ButtonBehaviorEx(grip, c.GetID("#RESIZE"), BtnNoFocus)
		if s.Started {
			c.dragOffset = mouse.Sub(r.Max)
		}
		if s.Held {
			w.Size = mouse.Sub(c.dragOffset).Sub(w.Pos).Max(st.WindowMinSize)
		}
	}
	// Title bar: close button then drag area.
	var closeR Rect
	var closeHot, hasClose bool
	moveID := c.GetID("#MOVE")
	if w.titleH > 0 {
		r := w.rect()
		w.clip = r
		title := Rect{Min: r.Min, Max: Vec2{X: r.Max.X, Y: r.Min.Y + w.titleH}}
		dragArea := title
		if open != nil {
			cs := w.titleH - 4
			cr := Rect{Min: Vec2{X: r.Max.X - cs - 2, Y: r.Min.Y + 2}, Max: Vec2{X: r.Max.X - 2, Y: r.Min.Y + 2 + cs}}
			dragArea.Max.X = cr.Min.X
			s := c.ButtonBehaviorEx(cr, c.GetID("#CLOSE"), BtnNoFocus)
			if s.Pressed {
				*open = false
			}
			closeR, closeHot, hasClose = cr, s.Hovered, true
		}
		if flags&NoMove == 0 {
			s := c.ButtonBehaviorEx(dragArea, moveID, BtnNoFocus)
			if s.Started {
				c.dragOffset = mouse.Sub(w.Pos)
				if c.In.MouseClicks[MouseLeft] >= 2 {
					c.toggleCollapse(w)
				}
			}
			if s.Held {
				w.Pos = mouse.Sub(c.dragOffset)
			}
		}
	}
	// Body drag started by End on a previous frame, only for title-less windows.
	if c.activeID == moveID && w.titleH == 0 {
		c.activeSeen = true
		if c.In.MouseDown[MouseLeft] {
			w.Pos = mouse.Sub(c.dragOffset)
		}
	}

	c.drawChrome(w, closeR, closeHot, hasClose)
	w.Root = w
	w.Parent = nil
	w.hasSize = true
	inner := Rect{Min: Vec2{X: w.Pos.X, Y: w.Pos.Y + w.titleH}, Max: w.Pos.Add(w.Size)}
	w.DL.PushClipRect(inner)
	w.clip = inner
	w.resetLayout(Vec2{X: w.Pos.X + pad.X, Y: w.Pos.Y + w.titleH + w.menuH + pad.Y}, inner.Max.X-pad.X)
	if w.collapsed {
		w.clip = Rect{}
		if st.BorderSize > 0 && flags&NoBackground == 0 {
			w.DL.AddRect(w.Pos, w.Pos.Add(w.Size), c.col(ColBorder), st.WindowRounding, st.BorderSize)
		}
		return false
	}
	return true
}

// toggleCollapse folds or unfolds w to its title bar.
func (c *Context) toggleCollapse(w *Window) {
	w.collapsed = !w.collapsed
	if w.collapsed {
		w.fullH = w.Size.Y
		w.Size.Y = w.titleH
	} else {
		w.Size.Y = w.fullH
	}
}

// Collapsed reports whether the window is folded to its title bar.
func (w *Window) Collapsed() bool { return w.collapsed }

// SetCollapsed folds or unfolds the window, taking effect on the next Begin.
func (w *Window) SetCollapsed(v bool) {
	if v != w.collapsed {
		w.collapsed = v
		if v {
			w.fullH = w.Size.Y
		} else {
			w.Size.Y = w.fullH
		}
	}
}

func (c *Context) drawChrome(w *Window, closeR Rect, closeHot, hasClose bool) {
	st := &c.Style
	dl := w.DL
	r := w.rect()
	if w.Flags&NoBackground == 0 {
		dl.AddRectFilled(r.Min, r.Max, c.col(ColWindowBg), st.WindowRounding)
	}
	if w.titleH > 0 {
		col := c.col(ColTitleBg)
		if c.focusedWindow == w {
			col = c.col(ColTitleBgActive)
		}
		corners := draw.CornerTopLeft | draw.CornerTopRight
		dl.AddRectFilledEx(r.Min, Vec2{X: r.Max.X, Y: r.Min.Y + w.titleH}, col, st.WindowRounding, corners)
		c.DrawText(Vec2{X: r.Min.X + st.FramePadding.X, Y: r.Min.Y + st.FramePadding.Y}, DisplayLabel(w.Name), c.col(ColText))
	}
	if hasClose {
		if closeHot {
			dl.AddRectFilled(closeR.Min, closeR.Max, c.col(ColCloseHovered), st.FrameRounding)
		}
		m := float32(4)
		tc := c.col(ColText)
		dl.AddLine(Vec2{X: closeR.Min.X + m, Y: closeR.Min.Y + m}, Vec2{X: closeR.Max.X - m, Y: closeR.Max.Y - m}, tc, 1.5)
		dl.AddLine(Vec2{X: closeR.Max.X - m, Y: closeR.Min.Y + m}, Vec2{X: closeR.Min.X + m, Y: closeR.Max.Y - m}, tc, 1.5)
	}
}

// End finishes the window started by Begin.
func (c *Context) End() {
	w := c.cur
	if w == nil || w.Child {
		panic("goak: End without matching Begin")
	}
	if len(c.idStack) != w.idDepth {
		panic("goak: unbalanced PushID or PopID before End")
	}
	if !w.skip {
		if !w.collapsed {
			w.ContentSize = w.maxPos.Sub(w.contentMin)
		}
		w.DL.PopClipRect()
		st := &c.Style
		r := w.rect()
		if st.BorderSize > 0 && w.Flags&NoBackground == 0 && !w.collapsed {
			w.DL.AddRect(r.Min, r.Max, c.col(ColBorder), st.WindowRounding, st.BorderSize)
		}
		if w.Flags&(NoResize|AutoSize) == 0 && !w.collapsed {
			c.drawGrip(w)
		}
		// Press on empty window body moves title-less windows.
		moveID := hashStr("#MOVE", w.ID)
		if w.titleH == 0 && w.Flags&NoMove == 0 && c.activeID == 0 && !c.pressClaimed &&
			c.In.MousePress[MouseLeft] && c.hoveredWindow == w && r.Contains(c.mouse()) {
			c.activeID, c.activeWindow, c.pressClaimed = moveID, w, true
			c.activeSeen = true
			c.dragOffset = c.mouse().Sub(w.Pos)
		}
	}
	c.popWindow()
}

func (c *Context) drawGrip(w *Window) {
	g := c.Style.ResizeGripSize
	r := w.rect()
	id := hashStr("#RESIZE", w.ID)
	col := c.col(ColResizeGrip)
	switch {
	case c.activeID == id:
		col = c.col(ColResizeGripActive)
	case c.hotID == id:
		col = c.col(ColResizeGripHovered)
	}
	// Inset by the border so the grip does not paint over it.
	b := c.Style.BorderSize
	br := Vec2{X: r.Max.X - b, Y: r.Max.Y - b}
	w.DL.AddTriangleFilled(Vec2{X: br.X, Y: br.Y - g}, br, Vec2{X: br.X - g, Y: br.Y}, col)
}

func (c *Context) popWindow() {
	if len(c.stack) == 0 || len(c.idStack) == 0 {
		panic("goak: window stack underflow")
	}
	c.stack = c.stack[:len(c.stack)-1]
	c.idStack = c.idStack[:len(c.idStack)-1]
	if n := len(c.stack); n > 0 {
		c.cur = c.stack[n-1]
	} else {
		c.cur = nil
	}
}

type childState struct {
	Y   float32
	Bar bool
}

// BeginChild starts a scrolling region of the given size inside the current
// window. A size component <= 0 means the available space plus that value.
// Always pair with EndChild. Scroll extent and scrollbar presence lag one
// frame behind the content. Only vertical scrolling is supported.
func (c *Context) BeginChild(name string, size Vec2, flags WindowFlags) bool {
	p := c.cur
	if p == nil {
		panic("goak: BeginChild outside Begin")
	}
	st := &c.Style
	avail := c.ContentRegionAvail()
	if size.X <= 0 {
		size.X = max(avail.X+size.X, 1)
	}
	if size.Y <= 0 {
		size.Y = max(avail.Y+size.Y, c.lineHeight())
	}
	id := c.GetID(name)
	cs := StateGet(c.States, id, childState{})
	r := c.ItemAdd(size)
	w := c.children[id]
	if w == nil {
		w = &Window{}
		c.children[id] = w
	}
	*w = Window{ID: id, Name: name, Flags: flags, Pos: r.Min, Size: size, DL: p.DL, Parent: p, Root: p.Root, Child: true,
		lastFrame: c.Frame, seenFrame: c.Frame, skip: p.skip, scrollY: cs.Y, needBar: cs.Bar && flags&NoScrollbar == 0,
		childPad: st.ChildPadding, itemWidths: w.itemWidths[:0]}
	w.outer = r.Intersect(p.clip)
	bar := float32(0)
	if w.needBar {
		bar = st.ScrollbarSize
	}
	StateSetKeep(c.States, id, cs, LongAge)
	c.stack = append(c.stack, w)
	c.cur = w
	c.idStack = append(c.idStack, id)
	w.idDepth = len(c.idStack)
	content := Rect{Min: r.Min, Max: Vec2{X: r.Max.X - bar, Y: r.Max.Y}}
	w.clip = content.Intersect(p.clip)
	if p.skip || w.clip.Empty() {
		w.skip = true
		w.clip = Rect{}
	}
	if !p.skip {
		if flags&NoBackground == 0 {
			p.DL.AddRectFilled(r.Min, r.Max, c.col(ColChildBg), st.ChildRounding)
		}
		p.DL.PushClipRect(content)
	}
	pad := w.childPad
	w.resetLayout(Vec2{X: r.Min.X + pad.X, Y: r.Min.Y + pad.Y - w.scrollY}, r.Max.X-bar-pad.X)
	return !w.skip
}

// EndChild finishes a child region.
func (c *Context) EndChild() {
	w := c.cur
	if w == nil || !w.Child {
		panic("goak: EndChild without matching BeginChild")
	}
	if len(c.idStack) != w.idDepth {
		panic("goak: unbalanced PushID or PopID before EndChild")
	}
	p := w.Parent
	st := &c.Style
	if !p.skip {
		p.DL.PopClipRect()
	}
	total := (w.maxPos.Y - w.contentMin.Y) + 2*w.childPad.Y
	maxScroll := max(total-w.Size.Y, 0)
	hasBar := maxScroll > 0 && w.Flags&NoScrollbar == 0
	scroll := w.scrollY

	// Wheel goes to the innermost hovered child.
	w.clip = w.outer
	if hasBar && !w.skip && !c.wheelUsed && c.In.Wheel[1] != 0 && c.mouseOver(w.outer) {
		scroll -= c.In.Wheel[1] * st.ScrollStep * c.lineHeight()
		c.wheelUsed = true
	}
	scroll = min(max(scroll, 0), maxScroll)

	if hasBar && !w.skip {
		scroll = c.scrollbar(w, scroll, maxScroll, total)
	}
	StateSetKeep(c.States, w.ID, childState{Y: scroll, Bar: hasBar}, LongAge)
	w.scrollY = scroll
	c.lastItem = Rect{Min: w.Pos, Max: w.Pos.Add(w.Size)}
	c.popWindow()
}

func (c *Context) scrollbar(w *Window, scroll, maxScroll, total float32) float32 {
	st := &c.Style
	r := Rect{Min: w.Pos, Max: w.Pos.Add(w.Size)}
	track := Rect{Min: Vec2{X: r.Max.X - st.ScrollbarSize, Y: r.Min.Y}, Max: r.Max}
	th := track.Max.Y - track.Min.Y
	grabH := min(max(th*w.Size.Y/total, st.GrabMinSize), th)
	span := th - grabH
	grabY := track.Min.Y
	if maxScroll > 0 {
		grabY += span * scroll / maxScroll
	}
	grab := Rect{Min: Vec2{X: track.Min.X, Y: grabY}, Max: Vec2{X: track.Max.X, Y: grabY + grabH}}
	id := c.GetID("#SCROLLY")
	s := c.ButtonBehaviorEx(track, id, BtnNoFocus)
	if s.Started {
		if grab.Contains(c.mouse()) {
			c.dragOffset.Y = c.mouse().Y - grab.Min.Y
		} else {
			c.dragOffset.Y = grabH / 2
		}
	}
	if s.Held && span > 0 {
		t := (c.mouse().Y - c.dragOffset.Y - track.Min.Y) / span
		scroll = min(max(t, 0), 1) * maxScroll
		grab.Min.Y = track.Min.Y + span*scroll/maxScroll
		grab.Max.Y = grab.Min.Y + grabH
	}
	col := c.col(ColScrollbarGrab)
	switch {
	case c.activeID == id:
		col = c.col(ColScrollbarGrabActive)
	case s.Hovered:
		col = c.col(ColScrollbarGrabHovered)
	}
	dl := w.DL
	dl.PushClipRect(w.outer)
	dl.AddRectFilled(track.Min, track.Max, c.col(ColScrollbarBg), st.ScrollbarRounding)
	dl.AddRectFilled(grab.Min, grab.Max, col, st.ScrollbarRounding)
	dl.PopClipRect()
	return scroll
}

// ScrollY is the scroll offset of the current child.
func (c *Context) ScrollY() float32 {
	if c.cur == nil {
		return 0
	}
	return c.cur.scrollY
}

// ChildScrollY returns the stored scroll of a child by name in the
// current ID scope.
func (c *Context) ChildScrollY(name string) float32 {
	return StateGet(c.States, c.GetID(name), childState{}).Y
}
