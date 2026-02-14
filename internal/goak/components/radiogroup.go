package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

// RadioOption represents a single option in a radio group.
type RadioOption struct {
	Label string
	Value string
}

// RadioGroup is a group of mutually exclusive radio buttons.
type RadioGroup struct {
	c              *layout.Container
	Options        []RadioOption
	SelectedIndex  int
	OnChanged      func(int, string)
	itemHeight     float64
	hoveredIndex   int
}

// NewRadioGroup creates a standalone radio group. Add it with panel.AddRadioGroup(rg).
func NewRadioGroup(width, height layout.Size, options []RadioOption) *RadioGroup {
	return &RadioGroup{
		c:             layout.NewContainer(width, height),
		Options:       options,
		SelectedIndex: -1,
		itemHeight:    24.0,
		hoveredIndex:  -1,
	}
}

// Bounds returns the computed layout rect after Layout.
func (rg *RadioGroup) Bounds() layout.Rect { return rg.c.Bounds }

// Container returns the layout node for this radio group (internal use).
func (rg *RadioGroup) Container() *layout.Container { return rg.c }

// SetItemHeight sets the height of each radio option.
func (rg *RadioGroup) SetItemHeight(height float64) {
	rg.itemHeight = height
}

// RadioTheme controls radio group drawing colors.
type RadioTheme struct {
	CircleFill   colors.Color
	CircleStroke colors.Color
	SelectedFill colors.Color
	Text         colors.Color
	HoverOverlay colors.Color
}

// DefaultRadioTheme returns the default radio theme.
func DefaultRadioTheme() RadioTheme {
	return RadioTheme{
		CircleFill:   colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
		CircleStroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
		SelectedFill: colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
		Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		HoverOverlay: colors.RGBA(255, 255, 255, 20),
	}
}

func (rg *RadioGroup) Draw(dst *ebiten.Image, face font.Face, theme RadioTheme) {
	bound := rg.Bounds()
	circleSize := 14.0
	circleRadius := circleSize / 2

	for i, opt := range rg.Options {
		y := bound.Y + float64(i)*rg.itemHeight
		circleY := y + (rg.itemHeight-circleSize)/2
		circleCenterX := bound.X + circleRadius
		circleCenterY := circleY + circleRadius

		rendering.DrawFilledCircle(dst, circleCenterX, circleCenterY, circleRadius, theme.CircleFill)
		rendering.DrawCircleStroke(dst, circleCenterX, circleCenterY, circleRadius, 1.0, theme.CircleStroke)

		if i == rg.SelectedIndex {
			innerRadius := circleRadius - 3.0
			rendering.DrawFilledCircle(dst, circleCenterX, circleCenterY, innerRadius, theme.SelectedFill)
		}

		if i == rg.hoveredIndex {
			rendering.DrawFilledCircle(dst, circleCenterX, circleCenterY, circleRadius, theme.HoverOverlay)
		}
		labelX := int(bound.X + circleSize + 8)
		th := face.Metrics().Height.Ceil()
		labelY := int(y+rg.itemHeight/2) + th/2 - 2
		rendering.DrawText(dst, opt.Label, face, labelX, labelY, theme.Text)
	}
}

// HitTest returns the index of the option at the given point, or -1.
func (rg *RadioGroup) HitTest(x, y float64) int {
	bound := rg.Bounds()
	if x < bound.X || x >= bound.X+bound.W {
		return -1
	}
	for i := range rg.Options {
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

// Select selects the option at the given index and calls OnChanged if set.
func (rg *RadioGroup) Select(index int) {
	if index < 0 || index >= len(rg.Options) {
		return
	}
	rg.SelectedIndex = index
	if rg.OnChanged != nil {
		rg.OnChanged(index, rg.Options[index].Value)
	}
}
