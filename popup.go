package goak

import (
	"sort"

	"github.com/tacf/goak/internal/draw"
)

// Clipboard is the text clipboard used by InputText. The sdl3 package
// provides one backed by the system clipboard.
type Clipboard interface {
	Get() string
	Set(string)
}

// MemClipboard is the default in-memory clipboard.
type MemClipboard struct{ Text string }

// Get returns the stored text.
func (m *MemClipboard) Get() string { return m.Text }

// Set stores s.
func (m *MemClipboard) Set(s string) { m.Text = s }

// popupEntry is one open popup in the stack, outermost first.
type popupEntry struct {
	id        ID
	pos       Vec2
	seq       int
	openFrame int
	menu      bool
}

const (
	tooltipSeq = 1 << 30
	mainBarSeq = 1
	popupSeq0  = 16
)

// layers returns live windows back to front with overlays last. The slice
// is reused.
func (c *Context) layers() []*Window {
	c.layerBuf = c.layerBuf[:0]
	for layer := 0; layer < 3; layer++ {
		for _, w := range c.order {
			if group := c.floatingDockRoot(w); group != nil {
				if layer == 2 && w == group.host {
					c.layerBuf = append(c.layerBuf, w)
					for _, member := range c.order {
						if member != w && c.floatingDockRoot(member) == group && !member.top {
							c.layerBuf = append(c.layerBuf, member)
						}
					}
				}
				continue
			}
			wl := 2
			if w.dockHost {
				wl = 0
			} else if w.docked {
				wl = 1
			}
			if !w.top && wl == layer {
				c.layerBuf = append(c.layerBuf, w)
			}
		}
	}
	n := len(c.layerBuf)
	for _, w := range c.order {
		if w.top {
			c.layerBuf = append(c.layerBuf, w)
		}
	}
	tops := c.layerBuf[n:]
	sort.SliceStable(tops, func(i, j int) bool { return tops[i].topSeq < tops[j].topSeq })
	return c.layerBuf
}

func popupWinID(id ID) ID { return hashStr("#POPUP", id) }

func (c *Context) popupIndex(id ID) int {
	for i, p := range c.popups {
		if p.id == id {
			return i
		}
	}
	return -1
}

// popupLevel is the stack depth a popup opened from the current window gets.
func (c *Context) popupLevel() int {
	if c.cur == nil {
		return 0
	}
	if pid := c.cur.Root.popupID; pid != 0 {
		if i := c.popupIndex(pid); i >= 0 {
			return i + 1
		}
	}
	return 0
}

// closePopupsFrom closes popups at stack index i and above.
func (c *Context) closePopupsFrom(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(c.popups) {
		return
	}
	for _, p := range c.popups[i:] {
		c.closedNow = append(c.closedNow, p.id)
	}
	c.popups = c.popups[:i]
}

func (c *Context) closedThisFrame(id ID) bool {
	for _, v := range c.closedNow {
		if v == id {
			return true
		}
	}
	return false
}

// popupMaintain closes popups on Escape and on presses outside the chain.
func (c *Context) popupMaintain() {
	c.closedNow = c.closedNow[:0]
	if len(c.popups) == 0 {
		return
	}
	in := c.In
	if in.KeyPressed[KeyEscape] {
		c.closePopupsFrom(len(c.popups) - 1)
		return
	}
	if !(in.MousePress[MouseLeft] || in.MousePress[MouseRight] || in.MousePress[MouseMiddle]) {
		return
	}
	keep := -1
	if hw := c.hoveredWindow; hw != nil && hw.popupID != 0 {
		keep = c.popupIndex(hw.popupID)
	}
	c.closePopupsFrom(keep + 1)
}

// OpenPopup opens the popup named id at the mouse position. Popups opened
// from inside another popup stack on top of it.
func (c *Context) OpenPopup(id string) { c.openPopupID(c.GetID(id), c.mouse(), false, false) }

func (c *Context) openPopupID(id ID, pos Vec2, menu, keep bool) {
	level := c.popupLevel()
	if level < len(c.popups) {
		if p := &c.popups[level]; p.id == id {
			if keep || p.openFrame == c.Frame {
				return
			}
			c.closePopupsFrom(level)
		} else {
			c.closePopupsFrom(level)
		}
	}
	c.popupSeq++
	c.popups = append(c.popups, popupEntry{id: id, pos: pos, seq: popupSeq0 + c.popupSeq, openFrame: c.Frame, menu: menu})
	if w := c.windows[popupWinID(id)]; w != nil {
		w.hasSize = false
	}
}

// IsPopupOpen reports whether the popup named id is open.
func (c *Context) IsPopupOpen(id string) bool { return c.popupIndex(c.GetID(id)) >= 0 }

// AnyPopupOpen reports whether any popup is open.
func (c *Context) AnyPopupOpen() bool { return len(c.popups) > 0 }

// ClosePopup closes the popup named id and its children.
func (c *Context) ClosePopup(id string) {
	if i := c.popupIndex(c.GetID(id)); i >= 0 {
		c.closePopupsFrom(i)
	}
}

// CloseCurrentPopup closes the popup the current window belongs to.
func (c *Context) CloseCurrentPopup() {
	if c.cur == nil {
		return
	}
	if i := c.popupIndex(c.cur.Root.popupID); c.cur.Root.popupID != 0 && i >= 0 {
		c.closePopupsFrom(i)
	}
}

// BeginPopup shows the popup named id when open. Pair a true result with
// EndPopup. A popup is invisible on its first frame while it measures itself.
func (c *Context) BeginPopup(id string) bool { return c.beginPopupID(c.GetID(id), id, 0) }

func (c *Context) beginPopupID(id ID, name string, minW float32) bool {
	i := c.popupIndex(id)
	if i < 0 {
		return false
	}
	p := c.popups[i]
	c.beginOverlay(popupWinID(id), name, p.pos, minW, p.seq, false)
	c.cur.popupID = id
	return true
}

// EndPopup ends a popup begun with BeginPopup.
func (c *Context) EndPopup() { c.endOverlay() }

// BeginPopupContextItem opens the popup on a right click of the last item
// and begins it. An empty id uses the last item ID.
func (c *Context) BeginPopupContextItem(id string) bool {
	pid := c.lastID
	if id != "" {
		pid = c.GetID(id)
	}
	if c.IsItemHovered() && c.In.MousePress[MouseRight] {
		c.openPopupID(pid, c.mouse(), false, false)
	}
	return c.beginPopupID(pid, id, 0)
}

// IsItemHovered reports whether the mouse is over the last item.
func (c *Context) IsItemHovered() bool {
	return c.cur != nil && (c.activeID == 0) && c.mouseOver(c.lastItem)
}

// beginOverlay starts a top layer window with its own ID scope. The
// window is hidden until it knows its size.
func (c *Context) beginOverlay(wid ID, name string, pos Vec2, minW float32, seq int, noInputs bool) *Window {
	w := c.windows[wid]
	if w == nil {
		w = &Window{ID: wid, Name: name, DL: draw.NewDrawList(), top: true}
		w.Root = w
		c.windows[wid] = w
		c.order = append(c.order, w)
	}
	w.top, w.noInputs, w.topSeq, w.minW = true, noInputs, seq, minW
	w.Flags = NoTitleBar | NoResize | AutoSize
	w.savedLast = c.lastItem
	w.savedID = c.lastID
	w.savedIDs = c.idStack
	w.idBuf = append(w.idBuf[:0], wid)
	c.idStack = w.idBuf
	w.idDepth = 1
	c.stack = append(c.stack, w)
	c.cur = w
	w.seenFrame = c.Frame
	w.DL.Reset()
	w.DL.Scale = c.pixelScale()
	w.titleH, w.menuH = 0, 0
	w.inMenu = false

	st := &c.Style
	pad := st.WindowPadding
	if w.hasSize {
		w.Size = w.ContentSize.Add(Vec2{X: 2 * pad.X, Y: 2 * pad.Y})
	} else {
		w.Size = st.WindowMinSize
	}
	w.Size.X = max(w.Size.X, minW)
	w.Pos = c.clampToDisplay(pos, w.Size)
	w.skip = !w.hasSize
	w.hasSize = true
	if w.skip {
		w.clip = Rect{}
		w.resetLayout(w.Pos.Add(pad), w.Pos.X+w.Size.X-pad.X)
		return w
	}
	w.lastFrame = c.Frame
	r := w.rect()
	w.DL.AddRectFilled(r.Min, r.Max, c.col(ColPopupBg), st.PopupRounding)
	w.DL.PushClipRect(r)
	w.clip = r
	w.resetLayout(w.Pos.Add(pad), r.Max.X-pad.X)
	return w
}

func (c *Context) clampToDisplay(p, size Vec2) Vec2 {
	if c.DisplaySize.X > 0 {
		p.X = min(p.X, c.DisplaySize.X-size.X)
	}
	if c.DisplaySize.Y > 0 {
		p.Y = min(p.Y, c.DisplaySize.Y-size.Y)
	}
	p.X, p.Y = max(p.X, 0), max(p.Y, 0)
	return p
}

func (c *Context) endOverlay() {
	w := c.cur
	if w == nil || !w.top || w.Child {
		panic("goak: EndPopup or EndTooltip without matching Begin")
	}
	if len(c.idStack) != w.idDepth {
		panic("goak: unbalanced PushID or PopID before EndPopup")
	}
	w.ContentSize = w.maxPos.Sub(w.contentMin)
	if !w.skip {
		w.DL.PopClipRect()
		if c.Style.BorderSize > 0 {
			r := w.rect()
			w.DL.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.PopupRounding, c.Style.BorderSize)
		}
	}
	c.stack = c.stack[:len(c.stack)-1]
	c.idStack = w.savedIDs
	w.savedIDs = nil
	c.lastItem = w.savedLast
	c.lastID = w.savedID
	if n := len(c.stack); n > 0 {
		c.cur = c.stack[n-1]
	} else {
		c.cur = nil
	}
}

// BeginTooltip starts a tooltip window that follows the mouse. Pair with
// EndTooltip.
func (c *Context) BeginTooltip() bool {
	m := c.mouse()
	c.beginOverlay(hashStr("#TOOLTIP", 0), "##Tooltip", Vec2{X: m.X + 14, Y: m.Y + 18}, 0, tooltipSeq, true)
	return true
}

// EndTooltip ends a tooltip.
func (c *Context) EndTooltip() { c.endOverlay() }

// SetTooltip shows a text tooltip for this frame.
func (c *Context) SetTooltip(text string) {
	c.BeginTooltip()
	c.Text(text)
	c.EndTooltip()
}
