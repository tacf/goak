package components

import (
	"math"

	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// ContextMenuItemKind describes a context menu entry kind.
type ContextMenuItemKind int

const (
	ContextMenuItemAction ContextMenuItemKind = iota
	ContextMenuItemSeparator
)

// ContextMenuItem is a context menu entry.
type ContextMenuItem struct {
	Kind     ContextMenuItemKind
	Label    string
	Hint     string
	Disabled bool
	action   Action
}

// NewContextMenuAction creates an enabled context menu action.
func NewContextMenuAction(label string, action Action) ContextMenuItem {
	return ContextMenuItem{
		Kind:   ContextMenuItemAction,
		Label:  label,
		action: action,
	}
}

// NewContextMenuActionWithHint creates an enabled context menu action with a
// right-aligned hint, typically a keyboard shortcut.
func NewContextMenuActionWithHint(label, hint string, action Action) ContextMenuItem {
	item := NewContextMenuAction(label, action)
	item.Hint = hint
	return item
}

// NewDisabledContextMenuAction creates a disabled context menu action.
func NewDisabledContextMenuAction(label string) ContextMenuItem {
	return ContextMenuItem{
		Kind:     ContextMenuItemAction,
		Label:    label,
		Disabled: true,
	}
}

// NewDisabledContextMenuActionWithHint creates a disabled context menu action
// with a right-aligned hint.
func NewDisabledContextMenuActionWithHint(label, hint string) ContextMenuItem {
	item := NewDisabledContextMenuAction(label)
	item.Hint = hint
	return item
}

// NewContextMenuSeparator creates a context menu separator.
func NewContextMenuSeparator() ContextMenuItem {
	return ContextMenuItem{Kind: ContextMenuItemSeparator}
}

// SetAction replaces this context menu item's action.
func (item *ContextMenuItem) SetAction(action Action) {
	item.action = action
}

// ContextMenu is a right-click popup menu.
type ContextMenu struct {
	items        []ContextMenuItem
	owner        *Element
	ui           *UI
	isOpen       bool
	anchorX      float64
	anchorY      float64
	x            float64
	y            float64
	hoveredIndex int
	itemHeight   float64
	separatorH   float64
	minWidth     float64
	measuredW    float64
	paddingX     float64
	paddingY     float64
	hintGap      float64
	viewport     layout.Rect
	hasViewport  bool
	scrollY      float64
	autoOpen     bool
	onAction     func(ContextMenuActionEvent)
}

// NewContextMenu creates a context menu with the given items.
func NewContextMenu(items []ContextMenuItem) *ContextMenu {
	menu := &ContextMenu{
		hoveredIndex: -1,
		itemHeight:   24.0,
		separatorH:   8.0,
		minWidth:     150.0,
		paddingX:     10.0,
		hintGap:      24.0,
		autoOpen:     true,
	}
	menu.SetItems(items)
	return menu
}

// IsOpen returns whether the context menu is currently visible.
func (cm *ContextMenu) IsOpen() bool { return cm.isOpen }

// Items returns a copy of the configured entries.
func (cm *ContextMenu) Items() []ContextMenuItem {
	if cm == nil {
		return nil
	}
	return append([]ContextMenuItem(nil), cm.items...)
}

// Remove detaches the menu overlay from its retained owner.
func (cm *ContextMenu) Remove() bool {
	if cm == nil || cm.owner == nil {
		return false
	}
	owner := cm.owner
	ui := cm.ui
	for index, menu := range owner.menus {
		if menu != cm {
			continue
		}
		copy(owner.menus[index:], owner.menus[index+1:])
		owner.menus = owner.menus[:len(owner.menus)-1]
		break
	}
	cm.owner = nil
	cm.ui = nil
	cm.Close()
	if ui != nil {
		ui.invalidateTree()
	}
	return true
}

// AutoOpen reports whether a retained-only window should open this menu on a
// right-click. It defaults to true. Scene overlays resolve application context
// themselves and open menus through SceneContext.
func (cm *ContextMenu) AutoOpen() bool { return cm.autoOpen }

// SetAutoOpen controls automatic right-click opening in retained-only windows.
// Scene overlays always resolve application context themselves.
func (cm *ContextMenu) SetAutoOpen(autoOpen bool) { cm.autoOpen = autoOpen }

// Open displays the context menu at the given position.
func (cm *ContextMenu) Open(x, y float64) {
	cm.open(x, y)
	cm.hasViewport = false
}

// OpenAt displays the context menu at the given position, clamped to the
// supplied viewport. Its final width is measured when the menu is drawn, so
// labels and hints remain inside the viewport whenever space permits.
func (cm *ContextMenu) OpenAt(x, y float64, viewport layout.Rect) {
	cm.open(x, y)
	cm.viewport = viewport
	cm.hasViewport = finiteValue(viewport.X) && finiteValue(viewport.Y) &&
		validPositiveLength(viewport.W) && validPositiveLength(viewport.H)
	cm.place()
}

func (cm *ContextMenu) open(x, y float64) {
	cm.isOpen = true
	cm.anchorX, cm.anchorY = x, y
	cm.x, cm.y = x, y
	cm.hoveredIndex = -1
	cm.scrollY = 0
}

// Close hides the context menu.
func (cm *ContextMenu) Close() {
	cm.isOpen = false
	cm.hoveredIndex = -1
}

// SetItemHeight sets the height of each menu item.
func (cm *ContextMenu) SetItemHeight(height float64) error {
	if !validPositiveLength(height) {
		return ErrInvalidLength
	}
	cm.itemHeight = height
	cm.place()
	return nil
}

// SetSeparatorHeight sets the vertical space occupied by each separator.
func (cm *ContextMenu) SetSeparatorHeight(height float64) error {
	if !validPositiveLength(height) {
		return ErrInvalidLength
	}
	cm.separatorH = height
	cm.place()
	return nil
}

// SetPadding sets the horizontal and vertical padding inside the menu.
func (cm *ContextMenu) SetPadding(horizontal, vertical float64) error {
	if !validLength(horizontal) || !validLength(vertical) {
		return ErrInvalidLength
	}
	cm.paddingX = horizontal
	cm.paddingY = vertical
	cm.measuredW = 0
	cm.place()
	return nil
}

// SetMinWidth sets the minimum width of the context menu.
func (cm *ContextMenu) SetMinWidth(width float64) error {
	if !validLength(width) {
		return ErrInvalidLength
	}
	cm.minWidth = width
	cm.place()
	return nil
}

// SetItems replaces all menu items. The supplied slice is copied so callers
// can safely reuse it after the call.
func (cm *ContextMenu) SetItems(items []ContextMenuItem) {
	cm.items = append(cm.items[:0], items...)
	cm.hoveredIndex = -1
	cm.measuredW = 0
	cm.scrollY = 0
	cm.place()
}

func validLength(value float64) bool {
	return value >= 0 && finiteValue(value)
}

func validPositiveLength(value float64) bool { return value > 0 && validLength(value) }

func finiteValue(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

// AddItem adds an action item to the context menu.
func (cm *ContextMenu) AddItem(label string, action Action) *ContextMenu {
	cm.items = append(cm.items, ContextMenuItem{
		Kind:   ContextMenuItemAction,
		Label:  label,
		action: action,
	})
	cm.measuredW = 0
	cm.place()
	return cm
}

// AddItemWithHint adds an action item with a right-aligned hint.
func (cm *ContextMenu) AddItemWithHint(label, hint string, action Action) *ContextMenu {
	cm.items = append(cm.items, NewContextMenuActionWithHint(label, hint, action))
	cm.measuredW = 0
	cm.place()
	return cm
}

// SetOnAction assigns a callback for all activated context menu actions.
func (cm *ContextMenu) SetOnAction(onAction func(ContextMenuActionEvent)) {
	cm.onAction = onAction
}

// AddSeparator adds a separator to the context menu.
func (cm *ContextMenu) AddSeparator() *ContextMenu {
	cm.items = append(cm.items, NewContextMenuSeparator())
	cm.place()
	return cm
}

// ScrollWheel scrolls an oversized menu by vertical SDL wheel units. Positive
// values move toward the first item and negative values toward the last item.
func (cm *ContextMenu) ScrollWheel(y float64) {
	if !cm.isOpen || y == 0 {
		return
	}
	step := max(cm.itemHeight, cm.separatorH)
	cm.scrollY = min(max(cm.scrollY-y*step, 0), cm.maxScroll())
}

func (cm *ContextMenu) Draw(renderer *sdl.Renderer, font *rendering.Font, theme ContextMenuTheme) {
	if !cm.isOpen {
		return
	}

	cm.measuredW = cm.measureWidth(func(text string) float64 {
		width, _ := font.Measure(text)
		return width
	})
	cm.place()
	bounds := cm.Bounds()

	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1.0, theme.Stroke)

	contentBounds := cm.contentBounds(bounds)
	if contentBounds.W <= 0 || contentBounds.H <= 0 {
		return
	}
	clip := sdl.Rect{
		X: int32(contentBounds.X),
		Y: int32(contentBounds.Y),
		W: int32(contentBounds.W),
		H: int32(contentBounds.H),
	}
	_ = renderer.SetClipRect(&clip)

	currentY := bounds.Y + cm.paddingY - cm.scrollY
	actionIndex := 0
	for _, item := range cm.items {
		if item.Kind == ContextMenuItemSeparator {
			sepY := currentY + cm.separatorH/2
			rendering.DrawLine(renderer, bounds.X+cm.paddingX, sepY,
				bounds.W-cm.paddingX*2, 1, theme.Separator, true)
			currentY += cm.separatorH
		} else {
			if actionIndex == cm.hoveredIndex && !item.Disabled {
				rendering.FillRect(renderer, bounds.X+1, currentY+1,
					bounds.W-2, max(0, cm.itemHeight-2), theme.Hover)
				if theme.Accent.A != 0 {
					rendering.FillRect(renderer, bounds.X+1, currentY+1,
						2, max(0, cm.itemHeight-2), theme.Accent)
				}
			}

			textColor := theme.Text
			if item.Disabled {
				textColor = theme.DisabledText
			}
			textY := textTopY(item.Label, font, currentY, cm.itemHeight)
			rendering.DrawText(renderer, item.Label, font, bounds.X+cm.paddingX, textY, textColor)
			if item.Hint != "" {
				hintWidth, _ := font.Measure(item.Hint)
				rendering.DrawText(renderer, item.Hint, font,
					bounds.X+bounds.W-cm.paddingX-hintWidth, textY, textColor)
			}

			currentY += cm.itemHeight
			actionIndex++
		}
	}
	_ = renderer.SetClipRect(nil)
}

func (cm *ContextMenu) Bounds() layout.Rect {
	if !cm.isOpen {
		return layout.Rect{}
	}

	width := max(cm.minWidth, cm.measuredW)
	if cm.hasViewport {
		width = min(width, cm.viewport.W)
	}
	height := cm.menuHeight()
	if cm.hasViewport {
		height = min(height, cm.viewport.H)
	}

	return layout.Rect{X: cm.x, Y: cm.y, W: width, H: height}
}

func (cm *ContextMenu) measureWidth(measure func(string) float64) float64 {
	width := cm.minWidth
	if measure == nil {
		return width
	}
	for _, item := range cm.items {
		if item.Kind != ContextMenuItemAction {
			continue
		}
		rowWidth := cm.paddingX*2 + measure(item.Label)
		if item.Hint != "" {
			rowWidth += cm.hintGap + measure(item.Hint)
		}
		width = max(width, rowWidth)
	}
	return width
}

func (cm *ContextMenu) place() {
	if !cm.hasViewport {
		cm.x, cm.y = cm.anchorX, cm.anchorY
		return
	}
	width := max(cm.minWidth, cm.measuredW)
	width = min(width, cm.viewport.W)
	height := cm.menuHeight()
	height = min(height, cm.viewport.H)
	maxX := max(cm.viewport.X, cm.viewport.X+cm.viewport.W-width)
	maxY := max(cm.viewport.Y, cm.viewport.Y+cm.viewport.H-height)
	cm.x = min(max(cm.anchorX, cm.viewport.X), maxX)
	cm.y = min(max(cm.anchorY, cm.viewport.Y), maxY)
	cm.scrollY = min(max(cm.scrollY, 0), cm.maxScroll())
}

func (cm *ContextMenu) menuHeight() float64 {
	return cm.paddingY*2 + cm.contentHeight()
}

func (cm *ContextMenu) contentHeight() float64 {
	height := 0.0
	for _, item := range cm.items {
		if item.Kind == ContextMenuItemSeparator {
			height += cm.separatorH
		} else {
			height += cm.itemHeight
		}
	}
	return height
}

func (cm *ContextMenu) contentBounds(bounds layout.Rect) layout.Rect {
	return layout.Rect{
		X: bounds.X,
		Y: bounds.Y + cm.paddingY,
		W: bounds.W,
		H: max(0, bounds.H-cm.paddingY*2),
	}
}

func (cm *ContextMenu) maxScroll() float64 {
	if !cm.hasViewport {
		return 0
	}
	return max(0, cm.contentHeight()-max(0, cm.Bounds().H-cm.paddingY*2))
}

// HitTest returns the action index at the given point, or -1.
// Skips separators and disabled items.
func (cm *ContextMenu) HitTest(x, y float64) int {
	if !cm.isOpen {
		return -1
	}

	bounds := cm.Bounds()
	if !rendering.PointWithinBounds(x, y, cm.contentBounds(bounds)) {
		return -1
	}

	currentY := cm.y + cm.paddingY - cm.scrollY
	actionIndex := 0
	for _, item := range cm.items {
		if item.Kind == ContextMenuItemSeparator {
			currentY += cm.separatorH
		} else {
			if y >= currentY && y < currentY+cm.itemHeight {
				if !item.Disabled {
					return actionIndex
				}
				return -1
			}
			currentY += cm.itemHeight
			actionIndex++
		}
	}

	return -1
}

// SetHovered sets which action index is hovered (-1 for none).
func (cm *ContextMenu) SetHovered(actionIndex int) {
	if _, ok := cm.enabledItem(actionIndex); ok {
		cm.hoveredIndex = actionIndex
		return
	}
	cm.hoveredIndex = -1
}

// HoveredIndex returns the currently hovered action index, or -1.
func (cm *ContextMenu) HoveredIndex() int { return cm.hoveredIndex }

// HandleKey handles semantic menu navigation keys. It recognizes up, down,
// home, end, return, space, and escape, and skips separators and disabled
// actions. It returns whether the key was consumed.
func (cm *ContextMenu) HandleKey(key string) bool {
	if !cm.isOpen {
		return false
	}
	switch key {
	case "up":
		cm.moveHover(-1)
	case "down":
		cm.moveHover(1)
	case "home":
		cm.hoveredIndex = cm.firstEnabled()
		cm.revealHovered()
	case "end":
		cm.hoveredIndex = cm.lastEnabled()
		cm.revealHovered()
	case "return", "space":
		cm.Click(cm.hoveredIndex)
	case "escape":
		cm.Close()
	default:
		return false
	}
	return true
}

func (cm *ContextMenu) moveHover(delta int) {
	count := cm.actionCount()
	if count == 0 || delta == 0 {
		return
	}
	index := cm.hoveredIndex
	if index < 0 {
		if delta < 0 {
			index = 0
		} else {
			index = -1
		}
	}
	for range count {
		index = (index + delta + count) % count
		if _, ok := cm.enabledItem(index); ok {
			cm.hoveredIndex = index
			cm.revealHovered()
			return
		}
	}
}

func (cm *ContextMenu) revealHovered() {
	top, ok := cm.actionOffset(cm.hoveredIndex)
	if !ok {
		return
	}
	visibleHeight := max(0, cm.Bounds().H-cm.paddingY*2)
	if top < cm.scrollY {
		cm.scrollY = top
	} else if bottom := top + cm.itemHeight; bottom > cm.scrollY+visibleHeight {
		cm.scrollY = bottom - visibleHeight
	}
	cm.scrollY = min(max(cm.scrollY, 0), cm.maxScroll())
}

func (cm *ContextMenu) actionOffset(actionIndex int) (float64, bool) {
	offset := 0.0
	index := 0
	for _, item := range cm.items {
		if item.Kind == ContextMenuItemSeparator {
			offset += cm.separatorH
			continue
		}
		if index == actionIndex {
			return offset, true
		}
		offset += cm.itemHeight
		index++
	}
	return 0, false
}

func (cm *ContextMenu) firstEnabled() int {
	for index := 0; index < cm.actionCount(); index++ {
		if _, ok := cm.enabledItem(index); ok {
			return index
		}
	}
	return -1
}

func (cm *ContextMenu) lastEnabled() int {
	for index := cm.actionCount() - 1; index >= 0; index-- {
		if _, ok := cm.enabledItem(index); ok {
			return index
		}
	}
	return -1
}

func (cm *ContextMenu) actionCount() int {
	count := 0
	for _, item := range cm.items {
		if item.Kind == ContextMenuItemAction {
			count++
		}
	}
	return count
}

func (cm *ContextMenu) enabledItem(actionIndex int) (ContextMenuItem, bool) {
	item, ok := cm.item(actionIndex)
	return item, ok && !item.Disabled
}

func (cm *ContextMenu) item(actionIndex int) (ContextMenuItem, bool) {
	if actionIndex < 0 {
		return ContextMenuItem{}, false
	}
	count := 0
	for _, item := range cm.items {
		if item.Kind != ContextMenuItemAction {
			continue
		}
		if count == actionIndex {
			return item, true
		}
		count++
	}
	return ContextMenuItem{}, false
}

func (cm *ContextMenu) Click(actionIndex int) {
	item, ok := cm.item(actionIndex)
	if !ok {
		return
	}
	if !item.Disabled {
		item.action.Invoke()
		if cm.onAction != nil {
			cm.onAction(ContextMenuActionEvent{
				Menu:  cm,
				Index: actionIndex,
				Item:  item,
			})
		}
	}

	cm.Close()
}
