package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

// ButtonTheme controls button drawing colors.
type ButtonTheme struct {
	Fill   colors.Color
	Stroke colors.Color
	Text   colors.Color
}

// DefaultButtonTheme returns the default button theme.
func DefaultButtonTheme() ButtonTheme {
	return ButtonTheme{
		Fill:   colors.HexOr("#404040", colors.RGB(64, 64, 64)),
		Stroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
		Text:   colors.HexOr("#eee", colors.RGB(238, 238, 238)),
	}
}

func (b *Button) Draw(dst *ebiten.Image, face text.GoTextFace, theme ButtonTheme) {
	bound := b.Bounds()
	rendering.FillRect(dst, bound.X, bound.Y, bound.W, bound.H, theme.Fill)
	rendering.DrawStrokeRect(dst, bound.X, bound.Y, bound.W, bound.H, 1.0, theme.Stroke)

	tw, th := text.Measure(b.Label, &face, 0)
	tx := bound.X + (bound.W-tw)/2
	ty := bound.Y + (bound.H-th)/2

	rendering.DrawText(dst, b.Label, face, int(tx), int(ty), theme.Text)
}
