package components

import (
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Button is a clickable control with a label.
// Create with NewButton for reuse; add with panel.AddButton(btn) and set OnClick per instance.
type Button struct {
	c       *layout.Container
	Label   string
	OnClick func()
}

// NewButton creates a standalone button (not in the tree). Add it with panel.AddButton(btn), then set OnClick.
func NewButton(width, height layout.Size, label string) *Button {
	return &Button{c: layout.NewContainer(width, height), Label: label}
}

// Bounds returns the computed layout rect after Layout.
func (b *Button) Bounds() layout.Rect { return b.c.Bounds }

func (b *Button) Draw(renderer *sdl.Renderer, font *rendering.Font, theme ButtonTheme) {
	bound := b.Bounds()
	rendering.FillRect(renderer, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(renderer, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	tw, th := font.Measure(b.Label)
	tx := bound.X + (bound.W-tw)/2
	ty := bound.Y + (bound.H-th)/2

	rendering.DrawText(renderer, b.Label, font, tx, ty, theme.Text)
}
