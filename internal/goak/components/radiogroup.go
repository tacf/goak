package components

import (
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// RadioOption represents a single option in a radio group.
type RadioOption struct {
	Label string
	Value string
}

// RadioGroup is a group of mutually exclusive radio buttons.
type RadioGroup struct {
	c             *layout.Container
	options       []RadioOption
	selectedIndex int
	onChanged     func(RadioChangedEvent)
	itemHeight    float64
	hoveredIndex  int
}

// NewRadioGroup creates a standalone radio group. Add it with panel.AddRadioGroup(rg).
func NewRadioGroup(width, height layout.Size, options []RadioOption) *RadioGroup {
	return &RadioGroup{
		c:             layout.NewContainer(width, height),
		options:       append([]RadioOption(nil), options...),
		selectedIndex: -1,
		itemHeight:    24.0,
		hoveredIndex:  -1,
	}
}

// Bounds returns the computed layout rect after Layout.
func (rg *RadioGroup) Bounds() layout.Rect { return rg.c.Bounds }

// Container returns the layout node for this radio group (internal use).
func (rg *RadioGroup) Container() *layout.Container { return rg.c }

// Options returns a copy of the radio options.
func (rg *RadioGroup) Options() []RadioOption {
	return append([]RadioOption(nil), rg.options...)
}

// SetOptions replaces the options and clears the selection.
func (rg *RadioGroup) SetOptions(options []RadioOption) {
	rg.options = append(rg.options[:0], options...)
	rg.selectedIndex = -1
	rg.hoveredIndex = -1
}

// SetItemHeight sets the height of each radio option.
func (rg *RadioGroup) SetItemHeight(height float64) {
	rg.itemHeight = height
}

// SelectedIndex returns the selected option index, or -1.
func (rg *RadioGroup) SelectedIndex() int { return rg.selectedIndex }

// SetOnChanged assigns the selection change callback.
func (rg *RadioGroup) SetOnChanged(onChanged func(RadioChangedEvent)) {
	rg.onChanged = onChanged
}

func (rg *RadioGroup) Draw(renderer *sdl.Renderer, font *rendering.Font, theme RadioTheme) {
	bound := rg.Bounds()
	circleSize := 14.0
	circleRadius := circleSize / 2

	for i, opt := range rg.options {
		y := bound.Y + float64(i)*rg.itemHeight
		circleY := y + (rg.itemHeight-circleSize)/2
		circleCenterX := bound.X + circleRadius
		circleCenterY := circleY + circleRadius

		rendering.DrawFilledCircle(renderer, circleCenterX, circleCenterY, circleRadius, theme.CircleFill)
		rendering.DrawCircleStroke(renderer, circleCenterX, circleCenterY, circleRadius, 1.0, theme.CircleStroke)

		if i == rg.selectedIndex {
			innerRadius := circleRadius - 3.0
			rendering.DrawFilledCircle(renderer, circleCenterX, circleCenterY, innerRadius, theme.SelectedFill)
		}

		if i == rg.hoveredIndex {
			rendering.DrawFilledCircle(renderer, circleCenterX, circleCenterY, circleRadius, theme.HoverOverlay)
		}
		labelX := int(bound.X + circleSize + 8)
		labelY := textTopY(opt.Label, font, y, rg.itemHeight)
		rendering.DrawText(renderer, opt.Label, font, float64(labelX), labelY, theme.Text)
	}
}

// HitTest returns the index of the option at the given point, or -1.
func (rg *RadioGroup) HitTest(x, y float64) int {
	bound := rg.Bounds()
	if x < bound.X || x >= bound.X+bound.W {
		return -1
	}
	for i := range rg.options {
		itemY := bound.Y + float64(i)*rg.itemHeight
		if y >= itemY && y < itemY+rg.itemHeight {
			return i
		}
	}
	return -1
}

// SetHovered sets which option index is hovered (-1 for none).
func (rg *RadioGroup) SetHovered(index int) {
	rg.hoveredIndex = index
}

// SetSelectedIndex selects an option and emits a change event when necessary.
func (rg *RadioGroup) SetSelectedIndex(index int) {
	if index < 0 || index >= len(rg.options) {
		return
	}
	if rg.selectedIndex == index {
		return
	}
	previous := rg.selectedIndex
	rg.selectedIndex = index
	if rg.onChanged != nil {
		rg.onChanged(RadioChangedEvent{
			RadioGroup:    rg,
			PreviousIndex: previous,
			Index:         index,
			Option:        rg.options[index],
		})
	}
}
