package goak

import (
	"github.com/tacf/goak/internal/draw"
)

// comboMaxItems is how many items a Combo shows before it scrolls.
const comboMaxItems = 8

// Selectable draws a full width row and returns true when clicked. Inside a
// popup window it closes that popup.
func (c *Context) Selectable(label string, selected bool) bool {
	return c.selectable(label, selected, 0)
}

func (c *Context) selectable(label string, selected bool, extraW float32) bool {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	st := &c.Style
	ts := c.TextSize(text)
	h := ts.Y + st.FramePadding.Y
	need := ts.X + 2*st.ItemInnerSpacing.X + extraW
	w := max(c.ContentRegionAvail().X, need)
	r := c.ItemAdd(Vec2{X: w, Y: h})
	if c.cur.skip {
		return false
	}
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	dl := c.cur.DL
	switch {
	case s.Held && c.activeID == id:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderActive), st.FrameRounding)
	case s.Hovered:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeaderHovered), st.FrameRounding)
	case selected:
		dl.AddRectFilled(r.Min, r.Max, c.col(ColHeader), st.FrameRounding)
	}
	if selected && s.Hovered {
		dl.AddRect(r.Min, r.Max, c.col(ColHeader), st.FrameRounding, 1)
	}
	c.DrawText(Vec2{X: r.Min.X + st.ItemInnerSpacing.X, Y: r.Min.Y + st.FramePadding.Y/2}, text, c.col(ColText))
	if s.Pressed && c.cur.top && c.cur.popupID != 0 {
		c.CloseCurrentPopup()
	}
	return s.Pressed
}

// BeginCombo draws the combo frame with the preview text and, when open,
// begins its popup. Pair a true result with EndCombo.
func (c *Context) BeginCombo(label, preview string) bool {
	id := c.ItemID(label)
	text := DisplayLabel(label)
	st := &c.Style
	w := c.CalcItemWidth()
	h := c.frameHeight()
	total := w
	if text != "" {
		total += st.ItemInnerSpacing.X + c.TextSize(text).X
	}
	r := c.ItemAdd(Vec2{X: total, Y: h})
	if c.cur.skip {
		return false
	}
	frame := Rect{Min: r.Min, Max: Vec2{X: r.Min.X + w, Y: r.Max.Y}}
	s := c.ButtonBehaviorEx(frame, id, 0)
	key := c.focusID == id && (c.In.KeyPressed[KeyEnter] || c.In.KeyPressed[KeySpace])
	pid := hashStr("#COMBO", id)
	open := c.popupIndex(pid) >= 0
	if (s.Started || key) && !c.closedThisFrame(pid) {
		if open {
			c.closePopupsFrom(c.popupIndex(pid))
		} else {
			c.openPopupID(pid, Vec2{X: frame.Min.X, Y: frame.Max.Y}, false, false)
		}
		open = c.popupIndex(pid) >= 0
	}
	if s.Started && c.closedThisFrame(pid) {
		open = false
	}
	dl := c.cur.DL
	col := c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive)
	c.drawFrame(frame, col, c.focusID == id)
	ar := Rect{Min: Vec2{X: frame.Max.X - h, Y: frame.Min.Y}, Max: frame.Max}
	dl.AddRectFilledEx(ar.Min, ar.Max, c.col(ColButton), st.FrameRounding, draw.CornerTopRight|draw.CornerBottomRight)
	m := h * 0.3
	cx, cy := (ar.Min.X+ar.Max.X)/2, (ar.Min.Y+ar.Max.Y)/2
	dl.AddTriangleFilled(Vec2{X: cx - m, Y: cy - m/2}, Vec2{X: cx + m, Y: cy - m/2}, Vec2{X: cx, Y: cy + m/2}, c.col(ColText))
	dl.PushClipRect(Rect{Min: frame.Min, Max: Vec2{X: ar.Min.X, Y: frame.Max.Y}})
	c.DrawText(Vec2{X: frame.Min.X + st.FramePadding.X, Y: frame.Min.Y + st.FramePadding.Y}, preview, c.col(ColText))
	dl.PopClipRect()
	c.frameLabel(frame, text)
	if !open {
		return false
	}
	i := c.popupIndex(pid)
	c.popups[i].pos = Vec2{X: frame.Min.X, Y: frame.Max.Y}
	return c.beginPopupID(pid, label, w)
}

// EndCombo ends a combo popup.
func (c *Context) EndCombo() { c.EndPopup() }

// Combo shows items and edits the selected index. It returns true when the
// selection changed.
func (c *Context) Combo(label string, cur *int, items []string) bool {
	if cur == nil {
		return false
	}
	preview := ""
	if *cur >= 0 && *cur < len(items) {
		preview = items[*cur]
	}
	if !c.BeginCombo(label, preview) {
		return false
	}
	changed := false
	scroll := len(items) > comboMaxItems
	if scroll {
		row := c.lineHeight() + c.Style.FramePadding.Y + c.Style.ItemSpacing.Y
		c.BeginChild("##items", Vec2{X: c.ContentRegionAvail().X, Y: float32(comboMaxItems)*row + 2*c.Style.ChildPadding.Y}, NoBackground)
	}
	for i, it := range items {
		c.PushID(i)
		if c.Selectable(it, i == *cur) {
			*cur, changed = i, true
			c.CloseCurrentPopup()
		}
		c.PopID()
	}
	if scroll {
		c.EndChild()
	}
	c.EndCombo()
	return changed
}

// BeginListBox starts a scrolling list frame of the given size. A size
// component <= 0 means the item width and seven rows. Always pair it with
// EndListBox, even when it returns false, like BeginChild.
func (c *Context) BeginListBox(label string, size Vec2) bool {
	text := DisplayLabel(label)
	if size.X <= 0 {
		size.X = c.CalcItemWidth()
	}
	if size.Y <= 0 {
		size.Y = c.listRows(7)
	}
	c.PushID(label)
	ok := c.BeginChild("##lb", size, 0)
	c.cur.labelText = text
	return ok
}

func (c *Context) listRows(n int) float32 {
	row := c.lineHeight() + c.Style.FramePadding.Y + c.Style.ItemSpacing.Y
	return float32(n)*row + 2*c.Style.ChildPadding.Y
}

// EndListBox ends a list box and draws its label to the right.
func (c *Context) EndListBox() {
	text := c.cur.labelText
	c.EndChild()
	c.PopID()
	if text != "" && !c.cur.skip {
		c.SameLine(0, -1)
		c.Text(text)
	}
}

// ListBox shows items and edits the selected index. heightInItems <= 0 means
// seven rows. It returns true when the selection changed.
func (c *Context) ListBox(label string, cur *int, items []string, heightInItems int) bool {
	if cur == nil {
		return false
	}
	if heightInItems <= 0 {
		heightInItems = 7
	}
	if !c.BeginListBox(label, Vec2{Y: c.listRows(heightInItems)}) {
		c.EndListBox()
		return false
	}
	changed := false
	for i, it := range items {
		c.PushID(i)
		if c.Selectable(it, i == *cur) {
			*cur, changed = i, true
		}
		c.PopID()
	}
	c.EndListBox()
	return changed
}
