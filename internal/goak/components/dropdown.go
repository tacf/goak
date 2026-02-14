package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

// DropdownOption represents a single option in a dropdown.
type DropdownOption struct {
	Label string
	Value string
}

// Dropdown is a collapsible list of options.
type Dropdown struct {
	c             *layout.Container
	Label         string
	Options       []DropdownOption
	SelectedIndex int
	OnChanged     func(int, string)
	isOpen        bool
	hoveredIndex  int
	itemHeight    float64
}

// NewDropdown creates a standalone dropdown. Add it with panel.AddDropdown(dd).
func NewDropdown(width, height layout.Size, label string, options []DropdownOption) *Dropdown {
	return &Dropdown{
		c:             layout.NewContainer(width, height),
		Label:         label,
		Options:       options,
		SelectedIndex: -1,
		itemHeight:    24.0,
		hoveredIndex:  -1,
	}
}

// Bounds returns the computed layout rect after Layout.
func (dd *Dropdown) Bounds() layout.Rect { return dd.c.Bounds }

// Container returns the layout node for this dropdown (internal use).
func (dd *Dropdown) Container() *layout.Container { return dd.c }

// IsOpen returns whether the dropdown is currently expanded.
func (dd *Dropdown) IsOpen() bool { return dd.isOpen }

// Open expands the dropdown list.
func (dd *Dropdown) Open() { dd.isOpen = true }

// Close collapses the dropdown list.
func (dd *Dropdown) Close() {
	dd.isOpen = false
	dd.hoveredIndex = -1
}

// Toggle toggles the dropdown open/closed state.
func (dd *Dropdown) Toggle() {
	if dd.isOpen {
		dd.Close()
	} else {
		dd.Open()
	}
}

// SetItemHeight sets the height of each dropdown option.
func (dd *Dropdown) SetItemHeight(height float64) {
	dd.itemHeight = height
}

// DropdownTheme controls dropdown drawing colors.
type DropdownTheme struct {
	Fill      colors.Color
	Stroke    colors.Color
	Hover     colors.Color
	Selected  colors.Color
	Text      colors.Color
	ArrowFill colors.Color
}

// DefaultDropdownTheme returns the default dropdown theme.
func DefaultDropdownTheme() DropdownTheme {
	return DropdownTheme{
		Fill:      colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
		Stroke:    colors.HexOr("#666", colors.RGB(102, 102, 102)),
		Hover:     colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
		Selected:  colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
		Text:      colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		ArrowFill: colors.HexOr("#aaa", colors.RGB(170, 170, 170)),
	}
}

func (dd *Dropdown) Draw(dst *ebiten.Image, face font.Face, theme DropdownTheme) {
	bound := dd.Bounds()

	rendering.FillRect(dst, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(dst, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	displayText := dd.Label
	if dd.SelectedIndex >= 0 && dd.SelectedIndex < len(dd.Options) {
		displayText = dd.Options[dd.SelectedIndex].Label
	}
	th := face.Metrics().Height.Ceil()
	textY := int(bound.Y+bound.H/2) + th/2 - 2
	rendering.DrawText(dst, displayText, face, int(bound.X+8), textY, theme.Text)

	arrowSize := 6.0
	arrowX := bound.X + bound.W - arrowSize - 8
	arrowY := bound.Y + (bound.H-arrowSize)/2
	if dd.isOpen {
		// Up arrow (triangle)
		rendering.FillRect(dst, arrowX, arrowY+arrowSize, arrowSize, 1, theme.ArrowFill)
		rendering.FillRect(dst, arrowX+1, arrowY+arrowSize-2, arrowSize-2, 1, theme.ArrowFill)
		rendering.FillRect(dst, arrowX+2, arrowY+arrowSize-4, arrowSize-4, 1, theme.ArrowFill)
	} else {
		// Down arrow (triangle)
		rendering.FillRect(dst, arrowX, arrowY, arrowSize, 1, theme.ArrowFill)
		rendering.FillRect(dst, arrowX+1, arrowY+2, arrowSize-2, 1, theme.ArrowFill)
		rendering.FillRect(dst, arrowX+2, arrowY+4, arrowSize-4, 1, theme.ArrowFill)
	}

	if dd.isOpen {
		dd.drawList(dst, face, theme)
	}
}

func (dd *Dropdown) drawList(dst *ebiten.Image, face font.Face, theme DropdownTheme) {
	bound := dd.Bounds()
	listY := bound.Y + bound.H
	listHeight := float64(len(dd.Options)) * dd.itemHeight

	rendering.FillRect(dst, bound.X, listY, bound.W, listHeight, theme.Fill)
	rendering.DrawStrokeRect(dst, bound.X, listY, bound.W, listHeight, 1.0, theme.Stroke)

	for i, opt := range dd.Options {
		itemY := listY + float64(i)*dd.itemHeight

		// Highlight selected or hovered
		if i == dd.SelectedIndex {
			rendering.FillRect(dst, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Selected)
		} else if i == dd.hoveredIndex {
			rendering.FillRect(dst, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Hover)
		}

		th := face.Metrics().Height.Ceil()
		textY := int(itemY+dd.itemHeight/2) + th/2 - 2
		rendering.DrawText(dst, opt.Label, face, int(bound.X+8), textY, theme.Text)
	}
}

// ListBounds returns the bounds of the expanded list when open.
func (dd *Dropdown) ListBounds() layout.Rect {
	if !dd.isOpen {
		return layout.Rect{}
	}
	bound := dd.Bounds()
	listY := bound.Y + bound.H
	listHeight := float64(len(dd.Options)) * dd.itemHeight
	return layout.Rect{X: bound.X, Y: listY, W: bound.W, H: listHeight}
}

// HitTestList returns the index of the option at the given point in the list, or -1.
func (dd *Dropdown) HitTestList(x, y float64) int {
	if !dd.isOpen {
		return -1
	}
	listBounds := dd.ListBounds()
	if !rendering.PointWithinBounds(x, y, listBounds) {
		return -1
	}
	relY := y - listBounds.Y
	index := int(relY / dd.itemHeight)
	if index >= 0 && index < len(dd.Options) {
		return index
	}
	return -1
}

// SetHovered sets which option index is hovered (-1 for none).
func (dd *Dropdown) SetHovered(index int) {
	dd.hoveredIndex = index
}

// Select selects the option at the given index and calls OnChanged if set.
func (dd *Dropdown) Select(index int) {
	if index < 0 || index >= len(dd.Options) {
		return
	}
	dd.SelectedIndex = index
	if dd.OnChanged != nil {
		dd.OnChanged(index, dd.Options[index].Value)
	}
	dd.Close()
}
