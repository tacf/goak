package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
	"github.com/tacf/goak/internal/font"
)

// TextRenderer measures and draws text. *font.Font implements it.
type TextRenderer interface {
	Measure(s string) (w, h float64)
	LineHeight() float64
	DrawText(dl *draw.DrawList, b draw.Renderer, x, y float32, s string, c colors.Color) error
	Finish(dl *draw.DrawList)
}

var _ TextRenderer = (*font.Font)(nil)

// Context owns all UI state across frames.
//
// Latency rules:
//   - Window rects used for hover and occlusion are those of the previous
//     frame, so a window that moved this frame still occludes where it was
//     drawn last frame.
//   - Window and child content size, and so AutoSize and scrollbar presence,
//     lag one frame behind the widgets that produced them.
//   - Items are drawn with the state computed in the same call, but window
//     z order changes from a click apply from that frame on.
type Context struct {
	Style   Style
	Font    TextRenderer
	Backend Renderer // passed to Font.DrawText, may be nil for fakes
	States  *StateStore
	// TextErr is the last error returned by the font.
	TextErr error

	// Clipboard backs InputText copy and paste. New sets an in-memory one.
	Clipboard Clipboard

	In          *Input
	DisplaySize Vec2
	Frame       int

	disabled  int // BeginDisabled depth
	groups    []groupState
	tables    []*tableState
	plots     []*plotBuild
	styleCols []styleColor

	platform Backend // set by NewContext
	fontData []byte
	fontPx   float64

	idStack    []ID
	seen       map[ID]int
	collisions int

	windows  map[ID]*Window
	children map[ID]*Window // reused child structs
	created  int
	order    []*Window // back to front
	stack    []*Window
	cur      *Window

	hoveredWindow *Window
	focusedWindow *Window
	activeWindow  *Window
	hotID         ID
	activeID      ID
	focusID       ID
	activeSeen    bool
	pressClaimed  bool
	blankPress    bool
	dragOffset    Vec2
	wheelUsed     bool

	focusList, prevFocusList []ID

	nextPos, nextSize        Vec2
	nextPosSet, nextSizeSet  bool
	nextPosCond, nextSizeCnd Cond

	lastItem Rect
	lastID   ID
	lists    []*draw.DrawList
	dd       draw.DrawData

	popups     []popupEntry
	popupSeq   int
	closedNow  []ID
	layerBuf   []*Window
	nextOpen   bool
	nextOpenOn bool
	nextOpenCd Cond
	tabBar     *tabBarState
	noMinSize  bool

	dockNodes     map[ID]*dockNode
	dockWindows   map[ID]*dockNode
	dockSerial    uint32
	dockDragging  *Window
	moveStart     Vec2
	dockPreview   *draw.DrawList
	floatingDocks []*dockNode
}

// Cond says when a SetNextWindow value applies.
type Cond uint8

// Conditions.
const (
	CondAlways Cond = iota
	CondOnce        // only when the window is first created
)

// New returns a context drawing text with f through b.
func New(f TextRenderer, b Renderer) *Context {
	return &Context{
		Style:     DefaultStyle(),
		Clipboard: &MemClipboard{},
		Font:      f,
		Backend:   b,
		States:    NewStateStore(),
		seen:      map[ID]int{},
		windows:   map[ID]*Window{},
		children:  map[ID]*Window{},
	}
}

func (c *Context) col(i Col) colors.Color {
	col := c.Style.Colors[i]
	if c.disabled > 0 {
		col.A /= 2
	}
	return col
}

func (c *Context) lineHeight() float32 {
	if c.Font == nil {
		return 0
	}
	return float32(c.Font.LineHeight())
}

// TextSize measures s, which may span several lines.
func (c *Context) TextSize(s string) Vec2 {
	if c.Font == nil {
		return Vec2{}
	}
	var w float32
	lines := 1
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			lw, _ := c.Font.Measure(s[start:i])
			w = max(w, float32(lw))
			if i < len(s) {
				lines++
			}
			start = i + 1
		}
	}
	return Vec2{X: w, Y: float32(lines) * c.lineHeight()}
}

// DrawText draws s at pos in the current window, splitting on newlines.
func (c *Context) DrawText(pos Vec2, s string, col colors.Color) {
	if c.Font == nil || c.cur == nil || s == "" {
		return
	}
	lh := c.lineHeight()
	clip := c.cur.DL.CurrentClip()
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			// Skip lines far above, below or right of the clip.
			if i > start && pos.Y+2*lh >= clip.Min.Y && pos.Y-lh <= clip.Max.Y && pos.X-lh <= clip.Max.X {
				if err := c.Font.DrawText(c.cur.DL, c.Backend, pos.X, pos.Y, s[start:i], col); err != nil {
					c.TextErr = err
				}
			}
			pos.Y += lh
			start = i + 1
		}
	}
}

// DL returns the draw list of the current window for custom drawing.
func (c *Context) DL() *draw.DrawList {
	if c.cur == nil {
		return nil
	}
	return c.cur.DL
}

// pixelScale is the font pixel scale, or 0 when the font has none.
func (c *Context) pixelScale() float32 {
	if sc, ok := c.Font.(interface{ Scale() float32 }); ok {
		return sc.Scale()
	}
	return 0
}

// NewFrame starts a frame. in must stay valid until Render. A context made
// by NewContext also rebuilds its font here when the backend scale changed.
func (c *Context) NewFrame(in *Input, display Vec2) {
	if c.cur != nil || len(c.stack) != 0 {
		panic("goak: NewFrame called inside a window")
	}
	if c.platform != nil {
		if s, ok := c.platform.ScaleChanged(); ok {
			if err := c.RebuildFont(c.fontData, c.fontPx, s); err != nil {
				c.TextErr = err
			}
		}
	}
	if in == nil {
		in = &Input{}
	}
	c.In = in
	c.DisplaySize = display
	c.Frame++
	c.States.Advance(c.Frame)
	c.pruneWindows()
	clear(c.seen)
	c.collisions = 0
	c.idStack = c.idStack[:0]
	c.hotID = 0
	c.wheelUsed = false
	c.activeSeen = false
	c.pressClaimed = false
	c.prevFocusList, c.focusList = c.focusList, c.prevFocusList[:0]

	left := MouseLeft
	if c.activeID != 0 && !in.MouseDown[left] && !in.MouseRel[left] {
		c.activeID, c.activeWindow = 0, nil
	}

	// Window under the mouse, topmost first. While an item is active the
	// mouse is captured by its window.
	c.hoveredWindow = nil
	if c.activeID != 0 && c.activeWindow != nil {
		c.hoveredWindow = c.activeWindow
	} else {
		m := c.mouse()
		layers := c.layers()
		for i := len(layers) - 1; i >= 0; i-- {
			w := layers[i]
			if w.lastFrame != c.Frame-1 || w.skip || w.noInputs {
				continue
			}
			if w.rect().Contains(m) && !c.dockHostPassesInput(w, m) {
				if group := c.floatingDockRoot(w); group != nil && c.floatingDockGrip(group).Contains(m) {
					c.hoveredWindow = group.host
					break
				}
				c.hoveredWindow = w
				break
			}
		}
	}
	if in.MousePress[left] && c.hoveredWindow != nil && c.activeID == 0 && !c.hoveredWindow.top {
		c.bringToFront(c.hoveredWindow)
		c.focusedWindow = c.hoveredWindow
	}
	c.popupMaintain()
	c.tabNavigate()
	c.lastItem = Rect{}
	c.tabBar = nil
	c.nextOpenOn = false
	c.submitFloatingDocks()
}

// Render finishes the frame and returns the draw data, lists back to front,
// ready for a backend. The data and its list slice are reused by the next
// Render.
func (c *Context) Render() *draw.DrawData {
	if len(c.stack) != 0 {
		panic("goak: Render with unbalanced Begin or BeginChild")
	}
	if len(c.idStack) != 0 {
		panic("goak: Render with unbalanced PushID")
	}
	in := c.In
	left := MouseLeft
	c.finishDockDrag()
	if c.activeID != 0 && !c.activeSeen {
		c.activeID, c.activeWindow = 0, nil
	}
	if in.MousePress[left] && !c.pressClaimed && c.activeID == 0 {
		c.blankPress = true
		c.focusID = 0
	}
	if !in.MouseDown[left] {
		c.blankPress = false
	}
	c.lists = c.lists[:0]
	layers := c.layers()
	for i, w := range layers {
		if w.lastFrame != c.Frame || w.skip {
			continue
		}
		if c.Font != nil {
			c.Font.Finish(w.DL)
		}
		c.lists = append(c.lists, w.DL)
		if group := c.floatingDockRoot(w); group != nil && group.grip != nil &&
			(i+1 == len(layers) || c.floatingDockRoot(layers[i+1]) != group) {
			c.lists = append(c.lists, group.grip)
		}
	}
	if c.dockPreview != nil && len(c.dockPreview.Cmds) != 0 {
		c.lists = append(c.lists, c.dockPreview)
	}
	c.dd.Lists = c.lists
	c.dd.DisplaySize = draw.Vec2(c.DisplaySize)
	c.dd.Scale = draw.Vec2{X: 1, Y: 1}
	if sc, ok := c.Font.(interface{ Scale() float32 }); ok {
		c.dd.Scale = draw.Vec2{X: sc.Scale(), Y: sc.Scale()}
	}
	return &c.dd
}

func (c *Context) mouse() Vec2 { return Vec2{X: c.In.MousePos[0], Y: c.In.MousePos[1]} }

// MousePos is the mouse position of this frame.
func (c *Context) MousePos() Vec2 { return c.mouse() }

// WantCaptureMouse reports whether the UI uses the mouse this frame, so
// the host should not forward it elsewhere.
func (c *Context) WantCaptureMouse() bool { return c.hoveredWindow != nil || c.activeID != 0 }

// HotID, ActiveID and FocusID expose the interaction state.
func (c *Context) HotID() ID    { return c.hotID }
func (c *Context) ActiveID() ID { return c.activeID }
func (c *Context) FocusID() ID  { return c.focusID }

// SetFocusID sets keyboard focus.
func (c *Context) SetFocusID(id ID) { c.focusID = id }

// LastItemRect is the rect of the last item added.
func (c *Context) LastItemRect() Rect { return c.lastItem }

func (c *Context) bringToFront(w *Window) {
	if group := c.floatingDockRoot(w); group != nil {
		w = group.host
	}
	for i, o := range c.order {
		if o == w {
			copy(c.order[i:], c.order[i+1:])
			c.order[len(c.order)-1] = w
			return
		}
	}
}

// tabNavigate moves focus with Tab using last frame's focusable order.
func (c *Context) tabNavigate() {
	in := c.In
	if in.KeyPressed[KeyEscape] {
		c.focusID = 0
	}
	list := c.prevFocusList
	if !in.KeyPressed[KeyTab] || len(list) == 0 {
		return
	}
	back := in.Has(ModShift)
	idx := -1
	for i, id := range list {
		if id == c.focusID {
			idx = i
		}
	}
	switch {
	case idx < 0 && back:
		idx = len(list) - 1
	case idx < 0:
		idx = 0
	case back:
		idx = (idx - 1 + len(list)) % len(list)
	default:
		idx = (idx + 1) % len(list)
	}
	c.focusID = list[idx]
}

// WindowMaxAge is how many frames a window may go unsubmitted before its
// position, size and draw list are forgotten.
const WindowMaxAge = 600

func (c *Context) pruneWindows() {
	if c.Frame%60 != 0 {
		return
	}
	for id, w := range c.children {
		if c.Frame-w.seenFrame > WindowMaxAge {
			delete(c.children, id)
		}
	}
	dropped := false
	for id, w := range c.windows {
		if c.Frame-w.seenFrame > WindowMaxAge {
			c.UndockWindow(w.Name)
			delete(c.windows, id)
			dropped = true
		}
	}
	if !dropped {
		return
	}
	keep := c.order[:0]
	for _, w := range c.order {
		if c.windows[w.ID] == w {
			keep = append(keep, w)
		}
	}
	clear(c.order[len(keep):])
	c.order = keep
	if c.focusedWindow != nil && c.windows[c.focusedWindow.ID] != c.focusedWindow {
		c.focusedWindow = nil
	}
}

// SetFont replaces the font and releases the old one: its atlas texture is
// destroyed through Backend and, if it has a Close method, it is closed. Call
// it between frames. Passing the current font is a no-op.
func (c *Context) SetFont(f TextRenderer) {
	old := c.Font
	if old == f {
		return
	}
	c.Font = f
	if a, ok := old.(interface{ Atlas() *font.Atlas }); ok && c.Backend != nil {
		if t := a.Atlas().Texture(); t != 0 {
			c.Backend.DestroyTexture(t)
		}
	}
	if cl, ok := old.(interface{ Close() }); ok {
		cl.Close()
	}
}

// RebuildFont builds a font from data (nil for the embedded one) at sizePx
// logical pixels for the given scale and installs it with SetFont. On error
// the current font stays in place and the error is returned.
func (c *Context) RebuildFont(data []byte, sizePx float64, scale float32) error {
	f, err := font.New(data, sizePx, float64(scale))
	if err != nil {
		return err
	}
	c.SetFont(f)
	return nil
}
