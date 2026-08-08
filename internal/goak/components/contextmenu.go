package components

import (
	"github.com/tacf/goak/internal/goak/layout"
	"github.com/tacf/goak/internal/goak/rendering"

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

// NewDisabledContextMenuAction creates a disabled context menu action.
func NewDisabledContextMenuAction(label string) ContextMenuItem {
	return ContextMenuItem{
		Kind:     ContextMenuItemAction,
		Label:    label,
		Disabled: true,
	}
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
	Items        []ContextMenuItem
	isOpen       bool
	x            float64
	y            float64
	hoveredIndex int
	itemHeight   float64
	separatorH   float64
	minWidth     float64
	onAction     func(ContextMenuActionEvent)
}

// NewContextMenu creates a context menu with the given items.
func NewContextMenu(items []ContextMenuItem) *ContextMenu {
	return &ContextMenu{
		Items:        items,
		hoveredIndex: -1,
		itemHeight:   24.0,
		separatorH:   8.0,
		minWidth:     150.0,
	}
}

// IsOpen returns whether the context menu is currently visible.
func (cm *ContextMenu) IsOpen() bool { return cm.isOpen }

// Open displays the context menu at the given position.
func (cm *ContextMenu) Open(x, y float64) {
	cm.isOpen = true
	cm.x = x
	cm.y = y
	cm.hoveredIndex = -1
}

// Close hides the context menu.
func (cm *ContextMenu) Close() {
	cm.isOpen = false
	cm.hoveredIndex = -1
}

// SetItemHeight sets the height of each menu item.
func (cm *ContextMenu) SetItemHeight(height float64) {
	cm.itemHeight = height
}

// SetMinWidth sets the minimum width of the context menu.
func (cm *ContextMenu) SetMinWidth(width float64) {
	cm.minWidth = width
}

// AddItem adds an action item to the context menu.
func (cm *ContextMenu) AddItem(label string, action Action) *ContextMenu {
	cm.Items = append(cm.Items, ContextMenuItem{
		Kind:   ContextMenuItemAction,
		Label:  label,
		action: action,
	})
	return cm
}

// SetOnAction assigns a callback for all activated context menu actions.
func (cm *ContextMenu) SetOnAction(onAction func(ContextMenuActionEvent)) {
	cm.onAction = onAction
}

// AddSeparator adds a separator to the context menu.
func (cm *ContextMenu) AddSeparator() *ContextMenu {
	cm.Items = append(cm.Items, NewContextMenuSeparator())
	return cm
}

func (cm *ContextMenu) Draw(renderer *sdl.Renderer, font *rendering.Font, theme ContextMenuTheme) {
	if !cm.isOpen {
		return
	}

	bounds := cm.Bounds()

	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1.0, theme.Stroke)

	currentY := cm.y
	actionIndex := 0
	for _, item := range cm.Items {
		if item.Kind == ContextMenuItemSeparator {
			sepY := currentY + cm.separatorH/2
			rendering.DrawLine(renderer, cm.x+6, sepY, bounds.W-12, 1, theme.Separator, true)
			currentY += cm.separatorH
		} else {
			if actionIndex == cm.hoveredIndex && !item.Disabled {
				rendering.FillRect(renderer, cm.x+1, currentY+1, bounds.W-2, cm.itemHeight-2, theme.Hover)
			}

			textColor := theme.Text
			if item.Disabled {
				textColor = theme.DisabledText
			}
			textY := textTopY(item.Label, font, currentY, cm.itemHeight)
			rendering.DrawText(renderer, item.Label, font, cm.x+10, textY, textColor)

			currentY += cm.itemHeight
			actionIndex++
		}
	}
}

func (cm *ContextMenu) Bounds() layout.Rect {
	if !cm.isOpen {
		return layout.Rect{}
	}

	width := cm.minWidth
	height := 0.0
	for _, item := range cm.Items {
		if item.Kind == ContextMenuItemSeparator {
			height += cm.separatorH
		} else {
			height += cm.itemHeight
		}
	}

	return layout.Rect{X: cm.x, Y: cm.y, W: width, H: height}
}

// HitTest returns the action index at the given point, or -1.
// Skips separators and disabled items.
func (cm *ContextMenu) HitTest(x, y float64) int {
	if !cm.isOpen {
		return -1
	}

	bounds := cm.Bounds()
	if !rendering.PointWithinBounds(x, y, bounds) {
		return -1
	}

	currentY := cm.y
	actionIndex := 0
	for _, item := range cm.Items {
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
	cm.hoveredIndex = actionIndex
}

func (cm *ContextMenu) Click(actionIndex int) {
	if actionIndex < 0 {
		return
	}

	realIndex := 0
	count := 0
	for i, item := range cm.Items {
		if item.Kind == ContextMenuItemAction {
			if count == actionIndex {
				realIndex = i
				break
			}
			count++
		}
	}

	if realIndex < len(cm.Items) {
		item := cm.Items[realIndex]
		if item.Kind == ContextMenuItemAction && !item.Disabled {
			item.action.Invoke()
			if cm.onAction != nil {
				cm.onAction(ContextMenuActionEvent{
					Menu:  cm,
					Index: actionIndex,
					Item:  item,
				})
			}
		}
	}

	cm.Close()
}
