package components

import (
	"github.com/tacf/goak/internal/goak/layout"
	"github.com/tacf/goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Checkbox is a toggleable control with a label.
type Checkbox struct {
	c         *layout.Container
	label     string
	checked   bool
	onChanged func(CheckboxChangedEvent)
}

// NewCheckbox creates a standalone checkbox.
func NewCheckbox(width, height layout.Size, label string) *Checkbox {
	return &Checkbox{c: layout.NewContainer(width, height), label: label}
}

// Bounds returns the computed layout rect after Layout.
func (cb *Checkbox) Bounds() layout.Rect { return cb.c.Bounds }

// Container returns the layout node for this checkbox (internal use).
func (cb *Checkbox) Container() *layout.Container { return cb.c }

// Label returns the checkbox text.
func (cb *Checkbox) Label() string { return cb.label }

// SetLabel updates the checkbox text.
func (cb *Checkbox) SetLabel(label string) { cb.label = label }

// Checked reports whether the checkbox is selected.
func (cb *Checkbox) Checked() bool { return cb.checked }

// SetChecked updates the checkbox and emits a change event when necessary.
func (cb *Checkbox) SetChecked(checked bool) {
	if cb.checked == checked {
		return
	}
	previous := cb.checked
	cb.checked = checked
	if cb.onChanged != nil {
		cb.onChanged(CheckboxChangedEvent{
			Checkbox: cb,
			Previous: previous,
			Checked:  checked,
		})
	}
}

// SetOnChanged assigns the checkbox change callback.
func (cb *Checkbox) SetOnChanged(onChanged func(CheckboxChangedEvent)) {
	cb.onChanged = onChanged
}

func (cb *Checkbox) Draw(renderer *sdl.Renderer, font *rendering.Font, theme CheckboxTheme, hovered bool) {
	bound := cb.Bounds()
	boxSize := 16.0
	boxY := bound.Y + (bound.H-boxSize)/2

	rendering.FillRect(renderer, bound.X, boxY, boxSize, boxSize, theme.BoxFill)
	rendering.DrawStrokeRect(renderer, bound.X, boxY, boxSize, boxSize, 1.0, theme.BoxStroke)

	if cb.checked {
		padding := 3.0
		rendering.FillRect(renderer, bound.X+padding, boxY+padding, boxSize-padding*2, boxSize-padding*2, theme.CheckFill)
	}

	if hovered {
		rendering.FillRect(renderer, bound.X, boxY, boxSize, boxSize, theme.HoverOverlay)
	}

	labelX := int(bound.X + boxSize + 8)
	labelY := textTopY(cb.label, font, bound.Y, bound.H)
	rendering.DrawText(renderer, cb.label, font, float64(labelX), labelY, theme.Text)
}

// Toggle switches the checkbox state.
func (cb *Checkbox) Toggle() {
	cb.SetChecked(!cb.checked)
}
