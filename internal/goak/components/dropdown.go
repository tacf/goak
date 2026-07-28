package components

import (
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// DropdownOption represents a single option in a dropdown.
type DropdownOption struct {
	Label string
	Value string
}

// Dropdown is a collapsible list of options.
type Dropdown struct {
	c             *layout.Container
	label         string
	options       []DropdownOption
	selectedIndex int
	onChanged     func(DropdownChangedEvent)
	isOpen        bool
	hoveredIndex  int
	itemHeight    float64
}

// NewDropdown creates a standalone dropdown. Add it with panel.AddDropdown(dd).
func NewDropdown(width, height layout.Size, label string, options []DropdownOption) *Dropdown {
	return &Dropdown{
		c:             layout.NewContainer(width, height),
		label:         label,
		options:       append([]DropdownOption(nil), options...),
		selectedIndex: -1,
		itemHeight:    24.0,
		hoveredIndex:  -1,
	}
}

// Bounds returns the computed layout rect after Layout.
func (dd *Dropdown) Bounds() layout.Rect { return dd.c.Bounds }

// Container returns the layout node for this dropdown (internal use).
func (dd *Dropdown) Container() *layout.Container { return dd.c }

// Label returns the dropdown placeholder text.
func (dd *Dropdown) Label() string { return dd.label }

// SetLabel updates the dropdown placeholder text.
func (dd *Dropdown) SetLabel(label string) { dd.label = label }

// Options returns a copy of the dropdown options.
func (dd *Dropdown) Options() []DropdownOption {
	return append([]DropdownOption(nil), dd.options...)
}

// SetOptions replaces the dropdown options and clears the selection.
func (dd *Dropdown) SetOptions(options []DropdownOption) {
	dd.options = append(dd.options[:0], options...)
	dd.selectedIndex = -1
	dd.Close()
}

// SelectedIndex returns the selected option index, or -1.
func (dd *Dropdown) SelectedIndex() int { return dd.selectedIndex }

// SetOnChanged assigns the dropdown selection callback.
func (dd *Dropdown) SetOnChanged(onChanged func(DropdownChangedEvent)) {
	dd.onChanged = onChanged
}

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

func (dd *Dropdown) Draw(renderer *sdl.Renderer, font *rendering.Font, theme DropdownTheme) {
	bound := dd.Bounds()

	rendering.FillRect(renderer, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	displayText := dd.label
	if dd.selectedIndex >= 0 && dd.selectedIndex < len(dd.options) {
		displayText = dd.options[dd.selectedIndex].Label
	}
	textY := textTopY(displayText, font, bound.Y, bound.H)
	rendering.DrawText(renderer, displayText, font, bound.X+8, textY, theme.Text)

	arrowSize := 6.0
	arrowX := bound.X + bound.W - arrowSize - 8
	arrowY := bound.Y + (bound.H-arrowSize)/2
	if dd.isOpen {
		// Up arrow (triangle)
		rendering.FillRect(renderer, arrowX, arrowY+arrowSize, arrowSize, 1, theme.ArrowFill)
		rendering.FillRect(renderer, arrowX+1, arrowY+arrowSize-2, arrowSize-2, 1, theme.ArrowFill)
		rendering.FillRect(renderer, arrowX+2, arrowY+arrowSize-4, arrowSize-4, 1, theme.ArrowFill)
	} else {
		// Down arrow (triangle)
		rendering.FillRect(renderer, arrowX, arrowY, arrowSize, 1, theme.ArrowFill)
		rendering.FillRect(renderer, arrowX+1, arrowY+2, arrowSize-2, 1, theme.ArrowFill)
		rendering.FillRect(renderer, arrowX+2, arrowY+4, arrowSize-4, 1, theme.ArrowFill)
	}

	if dd.isOpen {
		dd.drawList(renderer, font, theme)
	}
}

func (dd *Dropdown) drawList(renderer *sdl.Renderer, font *rendering.Font, theme DropdownTheme) {
	bound := dd.Bounds()
	listY := bound.Y + bound.H
	listHeight := float64(len(dd.options)) * dd.itemHeight

	rendering.FillRect(renderer, bound.X, listY, bound.W, listHeight, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, listY, bound.W, listHeight, 1.0, theme.Stroke)

	for i, opt := range dd.options {
		itemY := listY + float64(i)*dd.itemHeight

		// Highlight selected or hovered
		if i == dd.selectedIndex {
			rendering.FillRect(renderer, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Selected)
		} else if i == dd.hoveredIndex {
			rendering.FillRect(renderer, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Hover)
		}

		textY := textTopY(opt.Label, font, itemY, dd.itemHeight)
		rendering.DrawText(renderer, opt.Label, font, bound.X+8, textY, theme.Text)
	}
}

// ListBounds returns the bounds of the expanded list when open.
func (dd *Dropdown) ListBounds() layout.Rect {
	if !dd.isOpen {
		return layout.Rect{}
	}
	bound := dd.Bounds()
	listY := bound.Y + bound.H
	listHeight := float64(len(dd.options)) * dd.itemHeight
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
	if index >= 0 && index < len(dd.options) {
		return index
	}
	return -1
}

// SetHovered sets which option index is hovered (-1 for none).
func (dd *Dropdown) SetHovered(index int) {
	dd.hoveredIndex = index
}

// SetSelectedIndex selects an option and emits a change event when necessary.
func (dd *Dropdown) SetSelectedIndex(index int) {
	if index < 0 || index >= len(dd.options) {
		return
	}
	if dd.selectedIndex == index {
		dd.Close()
		return
	}
	previous := dd.selectedIndex
	dd.selectedIndex = index
	if dd.onChanged != nil {
		dd.onChanged(DropdownChangedEvent{
			Dropdown:      dd,
			PreviousIndex: previous,
			Index:         index,
			Option:        dd.options[index],
		})
	}
	dd.Close()
}
