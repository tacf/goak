package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
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
	OnClick  func()
	Disabled bool
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
func (cm *ContextMenu) AddItem(label string, onClick func()) *ContextMenu {
	cm.Items = append(cm.Items, ContextMenuItem{
		Kind:    ContextMenuItemAction,
		Label:   label,
		OnClick: onClick,
	})
	return cm
}

// AddSeparator adds a separator to the context menu.
func (cm *ContextMenu) AddSeparator() *ContextMenu {
	cm.Items = append(cm.Items, ContextMenuItem{Kind: ContextMenuItemSeparator})
	return cm
}

// ContextMenuTheme controls context menu drawing colors.
type ContextMenuTheme struct {
	Fill         colors.Color
	Stroke       colors.Color
	Hover        colors.Color
	Text         colors.Color
	DisabledText colors.Color
	Separator    colors.Color
}

// DefaultContextMenuTheme returns the default context menu theme.
func DefaultContextMenuTheme() ContextMenuTheme {
	return ContextMenuTheme{
		Fill:         colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
		Stroke:       colors.HexOr("#666", colors.RGB(102, 102, 102)),
		Hover:        colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
		Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		DisabledText: colors.HexOr("#777", colors.RGB(119, 119, 119)),
		Separator:    colors.HexOr("#555", colors.RGB(85, 85, 85)),
	}
}

func (cm *ContextMenu) Draw(dst *ebiten.Image, face font.Face, theme ContextMenuTheme) {
	if !cm.isOpen {
		return
	}

	bounds := cm.Bounds()

	rendering.FillRect(dst, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	rendering.DrawStrokeRect(dst, bounds.X, bounds.Y, bounds.W, bounds.H, 1.0, theme.Stroke)

	currentY := cm.y
	actionIndex := 0
	for _, item := range cm.Items {
		if item.Kind == ContextMenuItemSeparator {
			sepY := currentY + cm.separatorH/2
			rendering.DrawLine(dst, cm.x+6, sepY, bounds.W-12, 1, theme.Separator, true)
			currentY += cm.separatorH
		} else {
			if actionIndex == cm.hoveredIndex && !item.Disabled {
				rendering.FillRect(dst, cm.x+1, currentY+1, bounds.W-2, cm.itemHeight-2, theme.Hover)
			}

			textColor := theme.Text
			if item.Disabled {
				textColor = theme.DisabledText
			}
			th := face.Metrics().Height.Ceil()
			textY := int(currentY+cm.itemHeight/2) + th/2 - 2
			rendering.DrawText(dst, item.Label, face, int(cm.x+10), textY, textColor)

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
		if item.Kind == ContextMenuItemAction && !item.Disabled && item.OnClick != nil {
			item.OnClick()
		}
	}

	cm.Close()
}
