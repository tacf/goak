package components

import (
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
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

func (cb *Checkbox) Draw(renderer *sdl.Renderer, font *rendering.Font, theme CheckboxTheme, hovered bool) {
	bound := cb.Bounds()
	boxSize := 16.0
	boxY := bound.Y + (bound.H-boxSize)/2

	rendering.FillRect(renderer, bound.X, boxY, boxSize, boxSize, theme.BoxFill)
	rendering.DrawStrokeRect(renderer, bound.X, boxY, boxSize, boxSize, 1.0, theme.BoxStroke)

	if cb.Checked {
		padding := 3.0
		rendering.FillRect(renderer, bound.X+padding, boxY+padding, boxSize-padding*2, boxSize-padding*2, theme.CheckFill)
	}

	if hovered {
		rendering.FillRect(renderer, bound.X, boxY, boxSize, boxSize, theme.HoverOverlay)
	}

	labelX := int(bound.X + boxSize + 8)
	labelY := textTopY(cb.Label, font, bound.Y, bound.H)
	rendering.DrawText(renderer, cb.Label, font, float64(labelX), labelY, theme.Text)
}

// Toggle switches the checkbox state and calls OnChanged if set.
func (cb *Checkbox) Toggle() {
	cb.Checked = !cb.Checked
	if cb.OnChanged != nil {
		cb.OnChanged(cb.Checked)
	}
}
