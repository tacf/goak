package components

import (
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

type MenuBarWidthMode int

const (
	MenuBarWidthAuto MenuBarWidthMode = iota
	MenuBarWidthFull
)

type MenuEntryKind int

const (
	MenuEntryItem MenuEntryKind = iota
	MenuEntrySeparator
)

type MenuEntry struct {
	Kind   MenuEntryKind
	Label  string
	action Action
}

type MenuItem struct {
	label    string
	action   Action
	subItems []MenuEntry
	owner    *MenuBar
}

// NewMenuItem creates a detached menu item for use with MenuBar.SetItems.
func NewMenuItem(label string, action Action) MenuItem {
	return MenuItem{label: label, action: action}
}

func (m *MenuItem) Label() string {
	if m == nil {
		return ""
	}
	return m.label
}

func (m *MenuItem) SetLabel(label string) {
	if m == nil || m.label == label {
		return
	}
	m.label = label
	m.touch()
}

func (m *MenuItem) SetAction(action Action) {
	if m != nil {
		m.action = action
	}
}

func (m *MenuItem) SubItems() []MenuEntry {
	if m == nil {
		return nil
	}
	return append([]MenuEntry(nil), m.subItems...)
}

func (m *MenuItem) AddSubItem(label string, action Action) *MenuItem {
	if m == nil {
		return nil
	}
	m.subItems = append(m.subItems, MenuEntry{Kind: MenuEntryItem, Label: label, action: action})
	m.touch()
	return m
}

func (m *MenuItem) AddSeparator() *MenuItem {
	if m == nil {
		return nil
	}
	m.subItems = append(m.subItems, MenuEntry{Kind: MenuEntrySeparator})
	m.touch()
	return m
}

func (m *MenuItem) touch() {
	if m.owner != nil {
		m.owner.invalidateGeometry()
	}
}

// MenuBar is a horizontal menu strip with a viewport-aware popup submenu.
type MenuBar struct {
	Element
	c         *layout.Container
	widthMode MenuBarWidthMode
	items     []*MenuItem

	openIndex int
	hoverTop  int
	hoverSub  int
	onAction  func(MenuActionEvent)
	viewport  layout.Rect

	metricsDirty bool
	metricsFont  *rendering.Font
	measureText  func(string) float64
	topWidths    []float64
	topRects     []layout.Rect
	subRects     []layout.Rect
	subBounds    layout.Rect
	scrollY      float64
}

func NewMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	menu := &MenuBar{
		c:            layout.NewContainer(layout.AutoSize(), height),
		widthMode:    normalizedMenuBarWidthMode(widthMode),
		openIndex:    -1,
		hoverTop:     -1,
		hoverSub:     -1,
		metricsDirty: true,
	}
	menu.Element.init(menu)
	menu.applyWidthMode()
	return menu
}

func (m *MenuBar) Container() *layout.Container { return m.c }
func (m *MenuBar) Bounds() layout.Rect          { return m.c.Bounds }

func (m *MenuBar) WidthMode() MenuBarWidthMode { return m.widthMode }

func (m *MenuBar) SetWidthMode(mode MenuBarWidthMode) {
	mode = normalizedMenuBarWidthMode(mode)
	if m.widthMode == mode {
		return
	}
	m.widthMode = mode
	m.applyWidthMode()
	m.invalidateGeometry()
}

func normalizedMenuBarWidthMode(mode MenuBarWidthMode) MenuBarWidthMode {
	if mode == MenuBarWidthFull {
		return MenuBarWidthFull
	}
	return MenuBarWidthAuto
}

func (m *MenuBar) applyWidthMode() {
	if m.widthMode == MenuBarWidthFull {
		m.c.Width = layout.PercentOf(100)
	} else if m.c.Width.Kind != layout.Static {
		m.c.Width = layout.AutoSize()
	}
}

func (m *MenuBar) Items() []MenuItem {
	if m == nil {
		return nil
	}
	items := make([]MenuItem, len(m.items))
	for index := range m.items {
		items[index] = cloneMenuItem(*m.items[index], nil)
	}
	return items
}

func (m *MenuBar) SetItems(items []MenuItem) {
	if m == nil {
		return
	}
	m.items = make([]*MenuItem, len(items))
	for index := range items {
		item := cloneMenuItem(items[index], m)
		m.items[index] = &item
	}
	m.Close()
	m.invalidateGeometry()
}

func cloneMenuItem(item MenuItem, owner *MenuBar) MenuItem {
	return MenuItem{
		label:    item.label,
		action:   item.action,
		subItems: append([]MenuEntry(nil), item.subItems...),
		owner:    owner,
	}
}

func (m *MenuBar) AddItem(label string, action Action) *MenuItem {
	item := &MenuItem{label: label, action: action, owner: m}
	m.items = append(m.items, item)
	m.invalidateGeometry()
	return item
}

func (m *MenuBar) SetOnAction(onAction func(MenuActionEvent)) { m.onAction = onAction }

func (m *MenuBar) IsOpen() bool       { return m.openIndex >= 0 }
func (m *MenuBar) OpenIndex() int     { return m.openIndex }
func (m *MenuBar) HoverTopIndex() int { return m.hoverTop }
func (m *MenuBar) HoverSubIndex() int { return m.hoverSub }

func (m *MenuBar) Close() {
	m.openIndex = -1
	m.hoverSub = -1
	m.scrollY = 0
	m.subRects = m.subRects[:0]
	m.subBounds = layout.Rect{}
}

// SyncWidth measures labels with the active font and updates intrinsic width.
func (m *MenuBar) SyncWidth(font *rendering.Font) {
	m.syncWidth(font, func(text string) float64 {
		return measuredTextWidth(font, text)
	})
}

func (m *MenuBar) syncWidth(font *rendering.Font, measure func(string) float64) {
	if m == nil {
		return
	}
	if measure == nil {
		measure = func(text string) float64 { return measuredTextWidth(font, text) }
	}
	if m.metricsDirty || m.metricsFont != font || len(m.topWidths) != len(m.items) {
		m.topWidths = resizeFloatSlice(m.topWidths, len(m.items))
		for index, item := range m.items {
			m.topWidths[index] = measure(item.label) + menuTopPaddingX*2
		}
		m.metricsDirty = false
		m.metricsFont = font
		m.measureText = measure
	}
	if m.widthMode == MenuBarWidthFull {
		m.c.Width = layout.PercentOf(100)
		return
	}
	width := menuBarPaddingX * 2
	for _, itemWidth := range m.topWidths {
		width += itemWidth
	}
	width = max(width, 40)
	if m.c.Width.Kind != layout.Static || m.c.Width.Value != width {
		m.c.Width = layout.StaticPx(width)
		if m.ui != nil {
			m.ui.invalidateLayout()
		}
	}
}

func resizeFloatSlice(values []float64, length int) []float64 {
	if cap(values) < length {
		return make([]float64, length)
	}
	return values[:length]
}

func measuredTextWidth(font *rendering.Font, text string) float64 {
	if font == nil {
		return float64(len([]rune(text))) * 8
	}
	width, _ := font.Measure(text)
	return width
}

// Place updates cached bar and submenu geometry for a logical viewport.
func (m *MenuBar) Place(viewport layout.Rect) {
	if m == nil {
		return
	}
	m.viewport = viewport
	m.topRects = m.topRects[:0]
	x := m.c.Bounds.X + menuBarPaddingX
	for index := range m.items {
		width := measuredTextWidth(m.metricsFont, m.items[index].label) + menuTopPaddingX*2
		if index < len(m.topWidths) {
			width = m.topWidths[index]
		}
		m.topRects = append(m.topRects, layout.Rect{X: x, Y: m.c.Bounds.Y, W: width, H: m.c.Bounds.H})
		x += width
	}
	m.placeSubmenu()
}

func (m *MenuBar) placeSubmenu() {
	m.subRects = m.subRects[:0]
	m.subBounds = layout.Rect{}
	if m.openIndex < 0 || m.openIndex >= len(m.items) || m.openIndex >= len(m.topRects) {
		return
	}
	item := m.items[m.openIndex]
	if len(item.subItems) == 0 {
		return
	}
	width := m.openDropdownWidth(item)
	contentHeight := submenuContentHeight(item.subItems)
	x := m.topRects[m.openIndex].X
	y := m.c.Bounds.Y + m.c.Bounds.H
	height := contentHeight
	if finiteValue(m.viewport.X) && finiteValue(m.viewport.Y) &&
		validPositiveLength(m.viewport.W) && validPositiveLength(m.viewport.H) {
		width = min(width, m.viewport.W)
		x = min(max(x, m.viewport.X), m.viewport.X+m.viewport.W-width)
		below := max(0, m.viewport.Y+m.viewport.H-y)
		above := max(0, m.c.Bounds.Y-m.viewport.Y)
		if contentHeight > below && above > below {
			height = min(contentHeight, above)
			y = m.c.Bounds.Y - height
		} else {
			height = min(contentHeight, below)
		}
	}
	m.subBounds = layout.Rect{X: x, Y: y, W: width, H: height}
	m.scrollY = min(max(m.scrollY, 0), m.maxScroll())
	rowY := y - m.scrollY
	for _, entry := range item.subItems {
		rowHeight := menuSubItemHeight
		if entry.Kind == MenuEntrySeparator {
			rowHeight = menuSubSeparatorHeight
		}
		m.subRects = append(m.subRects, layout.Rect{X: x, Y: rowY, W: width, H: rowHeight})
		rowY += rowHeight
	}
}

func submenuContentHeight(entries []MenuEntry) float64 {
	height := 0.0
	for _, entry := range entries {
		if entry.Kind == MenuEntrySeparator {
			height += menuSubSeparatorHeight
		} else {
			height += menuSubItemHeight
		}
	}
	return height
}

func (m *MenuBar) TopItemRects() []layout.Rect {
	m.SyncWidth(m.metricsFont)
	m.Place(m.viewport)
	return append([]layout.Rect(nil), m.topRects...)
}

func (m *MenuBar) OpenSubItemRects() []layout.Rect {
	m.Place(m.viewport)
	return append([]layout.Rect(nil), m.subRects...)
}

func (m *MenuBar) OpenSubMenuBounds() layout.Rect { return m.subBounds }

func (m *MenuBar) OnMouseMove(x, y float64) {
	m.SyncWidth(m.metricsFont)
	m.Place(m.viewport)
	m.hoverTop = m.hitTopItem(x, y)
	if m.openIndex >= 0 && m.hoverTop >= 0 && m.hoverTop != m.openIndex &&
		len(m.items[m.hoverTop].subItems) > 0 {
		m.openIndex = m.hoverTop
		m.hoverSub = -1
		m.scrollY = 0
		m.placeSubmenu()
	}
	if m.openIndex >= 0 {
		m.hoverSub = m.hitSubItem(x, y, false)
	} else {
		m.hoverSub = -1
	}
}

func (m *MenuBar) OnMouseDown(x, y float64) bool {
	m.SyncWidth(m.metricsFont)
	m.Place(m.viewport)
	top := m.hitTopItem(x, y)
	if top >= 0 {
		item := m.items[top]
		if len(item.subItems) == 0 {
			m.Close()
			item.action.Invoke()
			m.emitAction(top, -1, item.label)
			return true
		}
		if m.openIndex == top {
			m.Close()
		} else {
			m.openIndex = top
			m.hoverSub = m.firstSubItem()
			m.scrollY = 0
			m.placeSubmenu()
			m.revealHovered()
		}
		return true
	}

	if m.openIndex >= 0 {
		sub := m.hitSubItem(x, y, true)
		if sub >= 0 {
			m.activateSubItem(sub)
			return true
		}
		m.Close()
		return true
	}
	return false
}

// HandleKey provides consistent keyboard navigation for open submenus.
func (m *MenuBar) HandleKey(key string) bool {
	if !m.IsOpen() {
		return false
	}
	switch key {
	case "left":
		m.moveTop(-1)
	case "right":
		m.moveTop(1)
	case "up":
		m.moveSub(-1)
	case "down":
		m.moveSub(1)
	case "home":
		m.hoverSub = m.firstSubItem()
		m.revealHovered()
	case "end":
		m.hoverSub = m.lastSubItem()
		m.revealHovered()
	case "return", "space":
		m.activateSubItem(m.hoverSub)
	case "escape":
		m.Close()
	default:
		return false
	}
	return true
}

func (m *MenuBar) ScrollWheel(y float64) {
	if !m.IsOpen() || y == 0 {
		return
	}
	m.scrollY = min(max(m.scrollY-y*menuSubItemHeight, 0), m.maxScroll())
	m.placeSubmenu()
}

func (m *MenuBar) moveTop(delta int) {
	if len(m.items) == 0 || delta == 0 {
		return
	}
	index := m.openIndex
	for range len(m.items) {
		index = (index + delta + len(m.items)) % len(m.items)
		if len(m.items[index].subItems) > 0 {
			m.openIndex = index
			m.hoverTop = index
			m.hoverSub = m.firstSubItem()
			m.scrollY = 0
			m.placeSubmenu()
			m.revealHovered()
			return
		}
	}
}

func (m *MenuBar) moveSub(delta int) {
	if m.openIndex < 0 || m.openIndex >= len(m.items) || delta == 0 {
		return
	}
	entries := m.items[m.openIndex].subItems
	if len(entries) == 0 {
		return
	}
	index := m.hoverSub
	if index < 0 {
		if delta > 0 {
			index = -1
		} else {
			index = 0
		}
	}
	for range len(entries) {
		index = (index + delta + len(entries)) % len(entries)
		if entries[index].Kind == MenuEntryItem {
			m.hoverSub = index
			m.revealHovered()
			return
		}
	}
}

func (m *MenuBar) firstSubItem() int {
	if m.openIndex < 0 || m.openIndex >= len(m.items) {
		return -1
	}
	for index, entry := range m.items[m.openIndex].subItems {
		if entry.Kind == MenuEntryItem {
			return index
		}
	}
	return -1
}

func (m *MenuBar) lastSubItem() int {
	if m.openIndex < 0 || m.openIndex >= len(m.items) {
		return -1
	}
	entries := m.items[m.openIndex].subItems
	for index := len(entries) - 1; index >= 0; index-- {
		if entries[index].Kind == MenuEntryItem {
			return index
		}
	}
	return -1
}

func (m *MenuBar) revealHovered() {
	if m.hoverSub < 0 || m.hoverSub >= len(m.subRects) {
		return
	}
	row := m.subRects[m.hoverSub]
	if row.Y < m.subBounds.Y {
		m.scrollY -= m.subBounds.Y - row.Y
	} else if bottom := row.Y + row.H; bottom > m.subBounds.Y+m.subBounds.H {
		m.scrollY += bottom - (m.subBounds.Y + m.subBounds.H)
	}
	m.scrollY = min(max(m.scrollY, 0), m.maxScroll())
	m.placeSubmenu()
}

func (m *MenuBar) maxScroll() float64 {
	if m.openIndex < 0 || m.openIndex >= len(m.items) {
		return 0
	}
	return max(0, submenuContentHeight(m.items[m.openIndex].subItems)-m.subBounds.H)
}

func (m *MenuBar) activateSubItem(index int) {
	if m.openIndex < 0 || m.openIndex >= len(m.items) {
		return
	}
	top := m.openIndex
	entries := m.items[top].subItems
	if index < 0 || index >= len(entries) || entries[index].Kind != MenuEntryItem {
		return
	}
	entry := entries[index]
	m.Close()
	entry.action.Invoke()
	m.emitAction(top, index, entry.Label)
}

func (m *MenuBar) emitAction(topIndex, subIndex int, label string) {
	if m.onAction != nil {
		m.onAction(MenuActionEvent{Menu: m, TopIndex: topIndex, SubIndex: subIndex, Label: label})
	}
}

func (m *MenuBar) openDropdownWidth(item *MenuItem) float64 {
	width := 120.0
	for _, entry := range item.subItems {
		if entry.Kind == MenuEntryItem {
			textWidth := measuredTextWidth(m.metricsFont, entry.Label)
			if m.measureText != nil {
				textWidth = m.measureText(entry.Label)
			}
			width = max(width, textWidth+menuSubContentPaddingX*2)
		}
	}
	return width
}

func (m *MenuBar) hitTopItem(x, y float64) int {
	for index, rect := range m.topRects {
		if rendering.PointWithinBounds(x, y, rect) {
			return index
		}
	}
	return -1
}

func (m *MenuBar) hitSubItem(x, y float64, includeSeparator bool) int {
	if !rendering.PointWithinBounds(x, y, m.subBounds) || m.openIndex < 0 || m.openIndex >= len(m.items) {
		return -1
	}
	for index, rect := range m.subRects {
		if rendering.PointWithinBounds(x, y, rect) {
			if !includeSeparator && m.items[m.openIndex].subItems[index].Kind == MenuEntrySeparator {
				return -1
			}
			return index
		}
	}
	return -1
}

func (m *MenuBar) invalidateGeometry() {
	m.metricsDirty = true
	if m.ui != nil {
		m.ui.invalidateLayout()
	}
}

const (
	menuBarPaddingX        = 8.0
	menuTopPaddingX        = 8.0
	menuSubContentPaddingX = 10.0
	menuSubItemHeight      = 22.0
	menuSubSeparatorHeight = 8.0
)

func (m *MenuBar) DrawBar(renderer *sdl.Renderer, font *rendering.Font, theme MenuTheme) {
	bounds := m.Bounds()
	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	for index, rect := range m.topRects {
		if index >= len(m.items) {
			break
		}
		if m.hoverTop == index {
			rendering.FillRect(renderer, rect.X, rect.Y, rect.W, rect.H, theme.Hover)
		}
		if m.openIndex == index {
			rendering.FillRect(renderer, rect.X, rect.Y, rect.W, rect.H, theme.Active)
		}
		textY := textTopY(m.items[index].label, font, rect.Y, rect.H)
		rendering.DrawText(renderer, m.items[index].label, font, rect.X+menuTopPaddingX, textY, theme.Text)
	}
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1, theme.Stroke)
}

func (m *MenuBar) DrawDropdown(renderer *sdl.Renderer, font *rendering.Font, theme MenuTheme) {
	if !m.IsOpen() || m.subBounds.W <= 0 || m.subBounds.H <= 0 ||
		m.openIndex >= len(m.items) {
		return
	}
	rendering.FillRect(renderer, m.subBounds.X, m.subBounds.Y, m.subBounds.W, m.subBounds.H, theme.Fill)
	clip := sdl.Rect{X: int32(m.subBounds.X), Y: int32(m.subBounds.Y), W: int32(m.subBounds.W), H: int32(m.subBounds.H)}
	_ = renderer.SetClipRect(&clip)
	entries := m.items[m.openIndex].subItems
	for index, rect := range m.subRects {
		if index >= len(entries) {
			break
		}
		entry := entries[index]
		if entry.Kind == MenuEntrySeparator {
			y := rect.Y + rect.H/2
			rendering.FillRect(renderer, rect.X+6, y, max(0, rect.W-12), 1, theme.Separator)
			continue
		}
		if m.hoverSub == index {
			rendering.FillRect(renderer, rect.X, rect.Y, rect.W, rect.H, theme.Hover)
		}
		textY := textTopY(entry.Label, font, rect.Y, rect.H)
		rendering.DrawText(renderer, entry.Label, font, rect.X+menuSubContentPaddingX, textY, theme.Text)
	}
	_ = renderer.SetClipRect(nil)
	rendering.DrawStrokeRect(renderer, m.subBounds.X, m.subBounds.Y, m.subBounds.W, m.subBounds.H, 1, theme.Stroke)
}
