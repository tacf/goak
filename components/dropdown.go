package components

import (
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// DropdownOption represents a single option in a dropdown.
type DropdownOption struct {
	Label string
	Value string
}

// Dropdown is a collapsible list of options.
type Dropdown struct {
	Element
	c             *layout.Container
	label         string
	options       []DropdownOption
	selectedIndex int
	onChanged     func(DropdownChangedEvent)
	isOpen        bool
	hoveredIndex  int
	itemHeight    float64
	viewport      layout.Rect
	hasViewport   bool
	listBounds    layout.Rect
	scrollY       float64
}

// NewDropdown creates a standalone dropdown. Add it with panel.AddDropdown(dd).
func NewDropdown(width, height layout.Size, label string, options []DropdownOption) *Dropdown {
	dropdown := &Dropdown{
		c:             layout.NewContainer(width, height),
		label:         label,
		options:       append([]DropdownOption(nil), options...),
		selectedIndex: -1,
		itemHeight:    24.0,
		hoveredIndex:  -1,
	}
	dropdown.Element.init(dropdown)
	return dropdown
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
func (dd *Dropdown) Open() {
	dd.isOpen = true
	dd.hoveredIndex = dd.selectedIndex
	if dd.hoveredIndex < 0 && len(dd.options) > 0 {
		dd.hoveredIndex = 0
	}
	dd.place()
	dd.revealHovered()
}

// Close collapses the dropdown list.
func (dd *Dropdown) Close() {
	dd.isOpen = false
	dd.hoveredIndex = -1
	dd.scrollY = 0
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
func (dd *Dropdown) SetItemHeight(height float64) error {
	if !validPositiveLength(height) {
		return ErrInvalidLength
	}
	dd.itemHeight = height
	dd.place()
	return nil
}

// Place constrains the popup to a logical UI viewport. It prefers opening
// downward and flips upward when that provides more usable space.
func (dd *Dropdown) Place(viewport layout.Rect) {
	dd.viewport = viewport
	dd.hasViewport = finiteValue(viewport.X) && finiteValue(viewport.Y) &&
		validPositiveLength(viewport.W) && validPositiveLength(viewport.H)
	dd.place()
}

func (dd *Dropdown) Draw(renderer *sdl.Renderer, font *rendering.Font, theme DropdownTheme) {
	dd.DrawControl(renderer, font, theme)
	dd.DrawList(renderer, font, theme)
}

// DrawControl draws the collapsed dropdown control without its popup list.
func (dd *Dropdown) DrawControl(renderer *sdl.Renderer, font *rendering.Font, theme DropdownTheme) {
	bound := dd.Bounds()

	rendering.FillRect(renderer, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	displayText := dd.label
	if dd.selectedIndex >= 0 && dd.selectedIndex < len(dd.options) {
		displayText = dd.options[dd.selectedIndex].Label
	}
	textY := textTopY(displayText, font, bound.Y, bound.H)
	rendering.DrawText(renderer, displayText, font, bound.X+8, textY, theme.Text)

	arrow := dropdownArrowGeometry(bound, dd.isOpen)
	rendering.FillTriangle(renderer,
		arrow[0].x, arrow[0].y,
		arrow[1].x, arrow[1].y,
		arrow[2].x, arrow[2].y,
		theme.ArrowFill,
	)
}

type dropdownArrowPoint struct{ x, y float64 }

func dropdownArrowGeometry(bounds layout.Rect, open bool) [3]dropdownArrowPoint {
	// Use logical control dimensions so the indicator scales with both widget
	// size and the renderer's HiDPI transform. The width-relative ceiling keeps
	// the triangle inside very narrow controls.
	width := min(max(5.0, bounds.H*0.32), max(0, bounds.W*0.16))
	height := width * 0.55
	rightInset := min(max(6.0, bounds.H*0.22), max(0, bounds.W-width))
	centerX := bounds.X + bounds.W - rightInset - width/2
	centerY := bounds.Y + bounds.H/2
	leftX, rightX := centerX-width/2, centerX+width/2
	topY, bottomY := centerY-height/2, centerY+height/2
	if open {
		return [3]dropdownArrowPoint{{leftX, bottomY}, {rightX, bottomY}, {centerX, topY}}
	}
	return [3]dropdownArrowPoint{{leftX, topY}, {rightX, topY}, {centerX, bottomY}}
}

// DrawList draws the popup option list when the dropdown is open.
func (dd *Dropdown) DrawList(renderer *sdl.Renderer, font *rendering.Font, theme DropdownTheme) {
	if !dd.isOpen {
		return
	}
	bound := dd.ListBounds()
	if bound.W <= 0 || bound.H <= 0 {
		return
	}

	rendering.FillRect(renderer, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, bound.Y, bound.W, bound.H, 1, theme.Stroke)
	clip := sdl.Rect{X: int32(bound.X), Y: int32(bound.Y), W: int32(bound.W), H: int32(bound.H)}
	_ = renderer.SetClipRect(&clip)

	for i, opt := range dd.options {
		itemY := bound.Y + float64(i)*dd.itemHeight - dd.scrollY

		// Highlight selected or hovered
		if i == dd.selectedIndex {
			rendering.FillRect(renderer, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Selected)
		} else if i == dd.hoveredIndex {
			rendering.FillRect(renderer, bound.X+1, itemY+1, bound.W-2, dd.itemHeight-2, theme.Hover)
		}

		textY := textTopY(opt.Label, font, itemY, dd.itemHeight)
		rendering.DrawText(renderer, opt.Label, font, bound.X+8, textY, theme.Text)
	}
	_ = renderer.SetClipRect(nil)
}

// ListBounds returns the bounds of the expanded list when open.
func (dd *Dropdown) ListBounds() layout.Rect {
	if !dd.isOpen {
		return layout.Rect{}
	}
	return dd.listBounds
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
	relY := y - listBounds.Y + dd.scrollY
	index := int(relY / dd.itemHeight)
	if index >= 0 && index < len(dd.options) {
		return index
	}
	return -1
}

// SetHovered sets which option index is hovered (-1 for none).
func (dd *Dropdown) SetHovered(index int) {
	if index < 0 || index >= len(dd.options) {
		dd.hoveredIndex = -1
		return
	}
	dd.hoveredIndex = index
}

// HoveredIndex returns the keyboard or pointer-highlighted option.
func (dd *Dropdown) HoveredIndex() int { return dd.hoveredIndex }

// HandleKey provides keyboard navigation while the list is open.
func (dd *Dropdown) HandleKey(key string) bool {
	if !dd.isOpen {
		return false
	}
	switch key {
	case "up":
		dd.moveHover(-1)
	case "down":
		dd.moveHover(1)
	case "home":
		if len(dd.options) > 0 {
			dd.hoveredIndex = 0
			dd.revealHovered()
		}
	case "end":
		if len(dd.options) > 0 {
			dd.hoveredIndex = len(dd.options) - 1
			dd.revealHovered()
		}
	case "return", "space":
		dd.SetSelectedIndex(dd.hoveredIndex)
	case "escape":
		dd.Close()
	default:
		return false
	}
	return true
}

// ScrollWheel scrolls an oversized popup list.
func (dd *Dropdown) ScrollWheel(y float64) {
	if !dd.isOpen || y == 0 {
		return
	}
	dd.scrollY = min(max(dd.scrollY-y*dd.itemHeight, 0), dd.maxScroll())
}

func (dd *Dropdown) moveHover(delta int) {
	if len(dd.options) == 0 || delta == 0 {
		return
	}
	if dd.hoveredIndex < 0 {
		dd.hoveredIndex = 0
	} else {
		dd.hoveredIndex = (dd.hoveredIndex + delta + len(dd.options)) % len(dd.options)
	}
	dd.revealHovered()
}

func (dd *Dropdown) revealHovered() {
	if dd.hoveredIndex < 0 || dd.hoveredIndex >= len(dd.options) {
		return
	}
	top := float64(dd.hoveredIndex) * dd.itemHeight
	if top < dd.scrollY {
		dd.scrollY = top
	} else if bottom := top + dd.itemHeight; bottom > dd.scrollY+dd.listBounds.H {
		dd.scrollY = bottom - dd.listBounds.H
	}
	dd.scrollY = min(max(dd.scrollY, 0), dd.maxScroll())
}

func (dd *Dropdown) maxScroll() float64 {
	return max(0, float64(len(dd.options))*dd.itemHeight-dd.listBounds.H)
}

func (dd *Dropdown) place() {
	control := dd.Bounds()
	contentHeight := float64(len(dd.options)) * dd.itemHeight
	dd.listBounds = layout.Rect{X: control.X, Y: control.Y + control.H, W: control.W, H: contentHeight}
	if dd.hasViewport {
		width := min(control.W, dd.viewport.W)
		x := min(max(control.X, dd.viewport.X), dd.viewport.X+dd.viewport.W-width)
		belowY := control.Y + control.H
		below := max(0, dd.viewport.Y+dd.viewport.H-belowY)
		above := max(0, control.Y-dd.viewport.Y)
		openAbove := contentHeight > below && above > below
		available := below
		y := belowY
		if openAbove {
			available = above
			y = control.Y - min(contentHeight, available)
		}
		height := min(contentHeight, available)
		dd.listBounds = layout.Rect{X: x, Y: y, W: width, H: height}
	}
	dd.scrollY = min(max(dd.scrollY, 0), dd.maxScroll())
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
