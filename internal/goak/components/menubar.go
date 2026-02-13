package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

// MenuBarWidthMode controls how the menu bar width is computed.
type MenuBarWidthMode int

const (
	// MenuBarWidthAuto sizes the bar to its items.
	MenuBarWidthAuto MenuBarWidthMode = iota
	// MenuBarWidthFull stretches the bar to full parent width.
	MenuBarWidthFull
)

// MenuEntryKind describes a submenu entry kind.
type MenuEntryKind int

const (
	MenuEntryItem MenuEntryKind = iota
	MenuEntrySeparator
)

// MenuEntry is a submenu row: either a clickable item or a separator.
type MenuEntry struct {
	Kind    MenuEntryKind
	Label   string
	OnClick func()
}

// MenuItem is a top-level menu label and optional submenu.
type MenuItem struct {
	Label    string
	OnClick  func()
	SubItems []MenuEntry
}

// AddSubItem appends a clickable submenu item.
func (m *MenuItem) AddSubItem(label string, onClick func()) *MenuItem {
	m.SubItems = append(m.SubItems, MenuEntry{
		Kind:    MenuEntryItem,
		Label:   label,
		OnClick: onClick,
	})
	return m
}

// AddSeparator appends a submenu separator.
func (m *MenuItem) AddSeparator() *MenuItem {
	m.SubItems = append(m.SubItems, MenuEntry{Kind: MenuEntrySeparator})
	return m
}

// MenuBar is a horizontal menu strip with optional dropdown submenus.
type MenuBar struct {
	ui        *UI
	c         *layout.Container
	WidthMode MenuBarWidthMode
	Items     []MenuItem

	openIndex int
	hoverTop  int
	hoverSub  int
}

// NewMenuBar creates a standalone menu bar (not in the tree).
func NewMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	width := layout.AutoSize()
	if widthMode == MenuBarWidthFull {
		width = layout.PercentOf(100)
	}
	return &MenuBar{
		c:         layout.NewContainer(width, height),
		WidthMode: widthMode,
		openIndex: -1,
		hoverTop:  -1,
		hoverSub:  -1,
	}
}

// Container returns the layout node for this menu bar (internal use).
func (m *MenuBar) Container() *layout.Container { return m.c }

// Bounds returns the computed layout rect after Layout.
func (m *MenuBar) Bounds() layout.Rect { return m.c.Bounds }

// AddItem appends a top-level menu item.
func (m *MenuBar) AddItem(label string, onClick func()) *MenuItem {
	m.Items = append(m.Items, MenuItem{Label: label, OnClick: onClick})
	return &m.Items[len(m.Items)-1]
}

// IsOpen reports whether any submenu is currently open.
func (m *MenuBar) IsOpen() bool { return m.openIndex >= 0 }

// OpenIndex returns the currently open top-level item index, or -1.
func (m *MenuBar) OpenIndex() int { return m.openIndex }

// HoverTopIndex returns the top-level hovered index, or -1.
func (m *MenuBar) HoverTopIndex() int { return m.hoverTop }

// HoverSubIndex returns the hovered submenu index, or -1.
func (m *MenuBar) HoverSubIndex() int { return m.hoverSub }

// Close closes any open submenu.
func (m *MenuBar) Close() {
	m.openIndex = -1
	m.hoverSub = -1
}

// SyncWidth updates layout width based on width mode.
func (m *MenuBar) SyncWidth() {
	if m.WidthMode == MenuBarWidthFull {
		m.c.Width = layout.PercentOf(100)
		return
	}
	w := menuBarPaddingX * 2
	for _, it := range m.Items {
		w += menuTopItemWidth(it.Label)
	}
	if w < 40 {
		w = 40
	}
	m.c.Width = layout.StaticPx(w)
}

// TopItemRects returns top-level item bounds for drawing and hit-testing.
func (m *MenuBar) TopItemRects() []layout.Rect {
	out := make([]layout.Rect, 0, len(m.Items))
	x := m.c.Bounds.X + menuBarPaddingX
	y := m.c.Bounds.Y
	h := m.c.Bounds.H
	for _, it := range m.Items {
		w := menuTopItemWidth(it.Label)
		out = append(out, layout.Rect{X: x, Y: y, W: w, H: h})
		x += w
	}
	return out
}

// OpenSubItemRects returns rects for the currently open submenu entries.
func (m *MenuBar) OpenSubItemRects() []layout.Rect {
	if m.openIndex < 0 || m.openIndex >= len(m.Items) {
		return nil
	}
	top := m.TopItemRects()
	if m.openIndex >= len(top) {
		return nil
	}
	item := m.Items[m.openIndex]
	if len(item.SubItems) == 0 {
		return nil
	}
	dropW := m.openDropdownWidth(item)
	x := top[m.openIndex].X
	y := m.c.Bounds.Y + m.c.Bounds.H

	rects := make([]layout.Rect, 0, len(item.SubItems))
	for _, ent := range item.SubItems {
		h := menuSubItemHeight
		if ent.Kind == MenuEntrySeparator {
			h = menuSubSeparatorHeight
		}
		rects = append(rects, layout.Rect{X: x, Y: y, W: dropW, H: h})
		y += h
	}
	return rects
}

// OpenSubMenuBounds returns the full dropdown bounds for the currently open submenu.
func (m *MenuBar) OpenSubMenuBounds() layout.Rect {
	rects := m.OpenSubItemRects()
	if len(rects) == 0 {
		return layout.Rect{}
	}
	var h float64
	for _, r := range rects {
		h += r.H
	}
	return layout.Rect{X: rects[0].X, Y: rects[0].Y, W: rects[0].W, H: h}
}

// OnMouseMove updates hover state. If a submenu is open, moving across top
// items switches the open submenu.
func (m *MenuBar) OnMouseMove(x, y float64) {
	m.hoverTop = m.hitTopItem(x, y)
	if m.openIndex >= 0 && m.hoverTop >= 0 && m.hoverTop != m.openIndex {
		if len(m.Items[m.hoverTop].SubItems) > 0 {
			m.openIndex = m.hoverTop
		}
	}
	if m.openIndex >= 0 {
		m.hoverSub = m.hitSubItem(x, y, false)
	} else {
		m.hoverSub = -1
	}
}

// OnMouseDown handles menu clicks. Returns true when the click was consumed.
func (m *MenuBar) OnMouseDown(x, y float64) bool {
	top := m.hitTopItem(x, y)
	if top >= 0 {
		item := m.Items[top]
		if len(item.SubItems) == 0 {
			if item.OnClick != nil {
				item.OnClick()
			}
			m.Close()
			return true
		}
		if m.openIndex == top {
			m.Close()
		} else {
			m.openIndex = top
			m.hoverSub = -1
		}
		return true
	}

	if m.openIndex >= 0 {
		sub := m.hitSubItem(x, y, true)
		if sub >= 0 {
			ent := m.Items[m.openIndex].SubItems[sub]
			if ent.Kind == MenuEntryItem {
				if ent.OnClick != nil {
					ent.OnClick()
				}
				m.Close()
				return true
			}
			return true
		}
		m.Close()
	}
	return false
}

func (m *MenuBar) openDropdownWidth(item MenuItem) float64 {
	w := menuSubContentPaddingX * 2
	for _, ent := range item.SubItems {
		if ent.Kind == MenuEntryItem {
			tw := menuTextWidth(ent.Label)
			if tw+menuSubContentPaddingX*2 > w {
				w = tw + menuSubContentPaddingX*2
			}
		}
	}
	if w < 120 {
		w = 120
	}
	return w
}

func (m *MenuBar) hitTopItem(x, y float64) int {
	for i, r := range m.TopItemRects() {
		if pointInRect(x, y, r) {
			return i
		}
	}
	return -1
}

func (m *MenuBar) hitSubItem(x, y float64, includeSeparator bool) int {
	rects := m.OpenSubItemRects()
	if len(rects) == 0 || m.openIndex < 0 || m.openIndex >= len(m.Items) {
		return -1
	}
	for i, r := range rects {
		if pointInRect(x, y, r) {
			if !includeSeparator && m.Items[m.openIndex].SubItems[i].Kind == MenuEntrySeparator {
				return -1
			}
			return i
		}
	}
	return -1
}

func pointInRect(x, y float64, r layout.Rect) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

const (
	menuBarPaddingX        = 8.0
	menuTopPaddingX        = 8.0
	menuSubContentPaddingX = 10.0
	menuSubItemHeight      = 22.0
	menuSubSeparatorHeight = 8.0
	menuCharWidth          = 8.0
)

func menuTopItemWidth(label string) float64 {
	return menuTextWidth(label) + menuTopPaddingX*2
}

func menuTextWidth(label string) float64 {
	return float64(len([]rune(label))) * menuCharWidth
}

// MenuTheme controls menu bar and dropdown colors.
type MenuTheme struct {
	Fill      colors.Color
	Stroke    colors.Color
	Hover     colors.Color
	Active    colors.Color
	Text      colors.Color
	Separator colors.Color
}

// DefaultMenuTheme returns the default menu color theme.
func DefaultMenuTheme() MenuTheme {
	return MenuTheme{
		Fill:      colors.HexOr("#202020", colors.RGB(32, 32, 32)),
		Stroke:    colors.HexOr("#525252", colors.RGB(82, 82, 82)),
		Hover:     colors.HexOr("#2f2f2f", colors.RGB(47, 47, 47)),
		Active:    colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
		Text:      colors.HexOr("#f0f0f0", colors.RGB(240, 240, 240)),
		Separator: colors.HexOr("#606060", colors.RGB(96, 96, 96)),
	}
}

// DrawBar draws the menu strip and top-level items.
func (m *MenuBar) DrawBar(dst *ebiten.Image, face font.Face, theme MenuTheme) {
	mb := m.Bounds()
	vector.FillRect(dst, float32(mb.X), float32(mb.Y), float32(mb.W), float32(mb.H), theme.Fill, true)
	drawMenuStrokeRect(dst, mb.X, mb.Y, mb.W, mb.H, theme.Stroke)

	topRects := m.TopItemRects()
	for i, r := range topRects {
		if m.HoverTopIndex() == i {
			vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), theme.Hover, true)
		}
		if m.OpenIndex() == i {
			vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), theme.Active, true)
		}
		text.Draw(dst, m.Items[i].Label, face, int(r.X)+8, int(r.Y+r.H/2)+5, theme.Text)
	}
}

// DrawDropdown draws the currently open dropdown, if any.
func (m *MenuBar) DrawDropdown(dst *ebiten.Image, face font.Face, theme MenuTheme) {
	if !m.IsOpen() {
		return
	}
	drop := m.OpenSubMenuBounds()
	if drop.W > 0 && drop.H > 0 {
		vector.FillRect(dst, float32(drop.X), float32(drop.Y), float32(drop.W), float32(drop.H), theme.Fill, true)
		drawMenuStrokeRect(dst, drop.X, drop.Y, drop.W, drop.H, theme.Stroke)
	}

	subRects := m.OpenSubItemRects()
	open := m.OpenIndex()
	if open < 0 || open >= len(m.Items) {
		return
	}
	subItems := m.Items[open].SubItems
	for i, r := range subRects {
		if i >= len(subItems) {
			break
		}
		entry := subItems[i]
		if entry.Kind == MenuEntrySeparator {
			y := r.Y + r.H/2
			ebitenutil.DrawRect(dst, r.X+6, y, r.W-12, 1, theme.Separator)
			continue
		}
		if m.HoverSubIndex() == i {
			vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), theme.Hover, true)
		}
		text.Draw(dst, entry.Label, face, int(r.X)+10, int(r.Y+r.H/2)+5, theme.Text)
	}
}

func drawMenuStrokeRect(dst *ebiten.Image, x, y, w, h float64, c colors.Color) {
	const t = 1.0
	ebitenutil.DrawRect(dst, x, y, w, t, c)
	ebitenutil.DrawRect(dst, x, y+h-t, w, t, c)
	ebitenutil.DrawRect(dst, x, y, t, h, c)
	ebitenutil.DrawRect(dst, x+w-t, y, t, h, c)
}
