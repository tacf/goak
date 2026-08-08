package components

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Label displays non-interactive text.
type Label struct {
	c               *layout.Container
	text            string
	color           *colors.Color
	horizontalAlign layout.Alignment
	verticalAlign   layout.Alignment
}

// NewLabel creates a label with the requested layout size and text.
func NewLabel(width, height layout.Size, text string) *Label {
	return &Label{
		c:             layout.NewContainer(width, height),
		text:          text,
		verticalAlign: layout.AlignCenter,
	}
}

// Container returns the label's underlying layout node.
func (l *Label) Container() *layout.Container { return l.c }

// Bounds returns the computed label bounds.
func (l *Label) Bounds() layout.Rect { return l.c.Bounds }

// Text returns the label text.
func (l *Label) Text() string { return l.text }

// SetText updates the label text.
func (l *Label) SetText(text string) { l.text = text }

// SetColor overrides the default label text color.
func (l *Label) SetColor(color colors.Color) {
	l.color = &color
}

// SetAlignment controls text placement within the label bounds.
func (l *Label) SetAlignment(horizontal, vertical layout.Alignment) {
	l.horizontalAlign = horizontal
	l.verticalAlign = vertical
}

// Draw renders the label text.
func (l *Label) Draw(renderer *sdl.Renderer, font *rendering.Font, theme LabelTheme) {
	if l == nil || font == nil || l.text == "" {
		return
	}

	bounds := l.Bounds()
	textW, textH := font.Measure(l.text)
	x := bounds.X
	y := bounds.Y

	switch l.horizontalAlign {
	case layout.AlignCenter:
		x += (bounds.W - textW) / 2
	case layout.AlignEnd:
		x += bounds.W - textW
	}
	switch l.verticalAlign {
	case layout.AlignCenter:
		y += (bounds.H - textH) / 2
	case layout.AlignEnd:
		y += bounds.H - textH
	}

	color := theme.Text
	if l.color != nil {
		color = *l.color
	}
	rendering.DrawText(renderer, l.text, font, x, y, color)
}
