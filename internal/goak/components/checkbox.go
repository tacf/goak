package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

// Checkbox is a toggleable control with a label.
type Checkbox struct {
	c         *layout.Container
	Label     string
	Checked   bool
	OnChanged func(bool)
}

// NewCheckbox creates a standalone checkbox. Add it with panel.AddCheckbox(cb), then set OnChanged.
func NewCheckbox(width, height layout.Size, label string) *Checkbox {
	return &Checkbox{c: layout.NewContainer(width, height), Label: label}
}

// Bounds returns the computed layout rect after Layout.
func (cb *Checkbox) Bounds() layout.Rect { return cb.c.Bounds }

// Container returns the layout node for this checkbox (internal use).
func (cb *Checkbox) Container() *layout.Container { return cb.c }

// CheckboxTheme controls checkbox drawing colors.
type CheckboxTheme struct {
	BoxFill      colors.Color
	BoxStroke    colors.Color
	CheckFill    colors.Color
	Text         colors.Color
	HoverOverlay colors.Color
}

// DefaultCheckboxTheme returns the default checkbox theme.
func DefaultCheckboxTheme() CheckboxTheme {
	return CheckboxTheme{
		BoxFill:      colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
		BoxStroke:    colors.HexOr("#666", colors.RGB(102, 102, 102)),
		CheckFill:    colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
		Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		HoverOverlay: colors.RGBA(255, 255, 255, 20),
	}
}

func (cb *Checkbox) Draw(dst *ebiten.Image, face font.Face, theme CheckboxTheme, hovered bool) {
	bound := cb.Bounds()
	boxSize := 16.0
	boxY := bound.Y + (bound.H-boxSize)/2

	rendering.FillRect(dst, bound.X, boxY, boxSize, boxSize, theme.BoxFill)
	rendering.DrawStrokeRect(dst, bound.X, boxY, boxSize, boxSize, 1.0, theme.BoxStroke)

	if cb.Checked {
		padding := 3.0
		rendering.FillRect(dst, bound.X+padding, boxY+padding, boxSize-padding*2, boxSize-padding*2, theme.CheckFill)
	}

	if hovered {
		rendering.FillRect(dst, bound.X, boxY, boxSize, boxSize, theme.HoverOverlay)
	}

	labelX := int(bound.X + boxSize + 8)
	th := face.Metrics().Height.Ceil()
	labelY := int(bound.Y+bound.H/2) + th/2 - 2
	rendering.DrawText(dst, cb.Label, face, labelX, labelY, theme.Text)
}

// Toggle switches the checkbox state and calls OnChanged if set.
func (cb *Checkbox) Toggle() {
	cb.Checked = !cb.Checked
	if cb.OnChanged != nil {
		cb.OnChanged(cb.Checked)
	}
}
