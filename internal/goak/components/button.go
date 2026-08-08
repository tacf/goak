package components

import (
	"github.com/tacf/goak/internal/goak/layout"
	"github.com/tacf/goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Button is a clickable control with a label.
type Button struct {
	c       *layout.Container
	label   string
	action  Action
	onClick func(ButtonClickEvent)
}

// NewButton creates a standalone button.
func NewButton(width, height layout.Size, label string) *Button {
	return &Button{c: layout.NewContainer(width, height), label: label}
}

// Bounds returns the computed layout rect after Layout.
func (b *Button) Bounds() layout.Rect { return b.c.Bounds }

// Label returns the button text.
func (b *Button) Label() string { return b.label }

// SetLabel updates the button text.
func (b *Button) SetLabel(label string) { b.label = label }

// SetAction assigns a reusable semantic action to the button.
func (b *Button) SetAction(action Action) { b.action = action }

// SetOnClick assigns a typed button event callback.
func (b *Button) SetOnClick(onClick func(ButtonClickEvent)) { b.onClick = onClick }

// Click activates the button action and callback.
func (b *Button) Click() {
	b.action.Invoke()
	if b.onClick != nil {
		b.onClick(ButtonClickEvent{Button: b})
	}
}

func (b *Button) Draw(renderer *sdl.Renderer, font *rendering.Font, theme ButtonTheme) {
	bound := b.Bounds()
	rendering.FillRect(renderer, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	tw, th := font.Measure(b.label)
	tx := bound.X + (bound.W-tw)/2
	ty := bound.Y + (bound.H-th)/2

	rendering.DrawText(renderer, b.label, font, tx, ty, theme.Text)
}
