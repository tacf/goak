package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
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

// Draw renders the button background, border, and label text.
func (b *Button) Draw(dst *ebiten.Image, face font.Face, theme ButtonTheme) {
	bound := b.Bounds()
	vector.FillRect(dst, float32(bound.X), float32(bound.Y), float32(bound.W), float32(bound.H), theme.Fill, true)
	drawButtonStrokeRect(dst, bound.X, bound.Y, bound.W, bound.H, theme.Stroke)

	tb := text.BoundString(face, b.Label)
	tw := tb.Dx()
	th := tb.Dy()
	tx := int(bound.X+bound.W/2) - tw/2
	ty := int(bound.Y+bound.H/2) + th/2 - 2
	text.Draw(dst, b.Label, face, tx, ty, theme.Text)
}

func drawButtonStrokeRect(dst *ebiten.Image, x, y, w, h float64, c colors.Color) {
	const t = 1.0
	ebitenutil.DrawRect(dst, x, y, w, t, c)
	ebitenutil.DrawRect(dst, x, y+h-t, w, t, c)
	ebitenutil.DrawRect(dst, x, y, t, h, c)
	ebitenutil.DrawRect(dst, x+w-t, y, t, h, c)
}
