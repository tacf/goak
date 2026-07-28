package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Label displays non-interactive text.
type Label struct {
	c               *layout.Container
	Text            string
	Color           *colors.Color
	HorizontalAlign layout.Alignment
	VerticalAlign   layout.Alignment
}

// NewLabel creates a label with the requested layout size and text.
func NewLabel(width, height layout.Size, text string) *Label {
	return &Label{
		c:             layout.NewContainer(width, height),
		Text:          text,
		VerticalAlign: layout.AlignCenter,
	}
}

// Container returns the label's underlying layout node.
func (l *Label) Container() *layout.Container { return l.c }

// Bounds returns the computed label bounds.
func (l *Label) Bounds() layout.Rect { return l.c.Bounds }

// SetColor overrides the default label text color.
func (l *Label) SetColor(color colors.Color) {
	l.Color = &color
}

// SetAlignment controls text placement within the label bounds.
func (l *Label) SetAlignment(horizontal, vertical layout.Alignment) {
	l.HorizontalAlign = horizontal
	l.VerticalAlign = vertical
}

// Draw renders the label text.
func (l *Label) Draw(renderer *sdl.Renderer, font *rendering.Font, theme LabelTheme) {
	if l == nil || font == nil || l.Text == "" {
		return
	}

	bounds := l.Bounds()
	textW, textH := font.Measure(l.Text)
	x := bounds.X
	y := bounds.Y

	switch l.HorizontalAlign {
	case layout.AlignCenter:
		x += (bounds.W - textW) / 2
	case layout.AlignEnd:
		x += bounds.W - textW
	}
	switch l.VerticalAlign {
	case layout.AlignCenter:
		y += (bounds.H - textH) / 2
	case layout.AlignEnd:
		y += bounds.H - textH
	}

	color := theme.Text
	if l.Color != nil {
		color = *l.Color
	}
	rendering.DrawText(renderer, l.Text, font, x, y, color)
}
