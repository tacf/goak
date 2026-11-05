package goak

import (
	"fmt"
	"strings"

	"github.com/tacf/goak/colors"
)

func (c *Context) fade(col colors.Color) colors.Color {
	if c.disabled > 0 {
		col.A /= 2
	}
	return col
}

// TextColored is Text in the given color.
func (c *Context) TextColored(col colors.Color, s string) {
	r := c.ItemAdd(c.TextSize(s))
	if c.cur.skip {
		return
	}
	c.DrawText(r.Min, s, c.fade(col))
}

// TextDisabled is Text in the disabled text color.
func (c *Context) TextDisabled(s string) { c.TextColored(c.col(ColTextDisabled), s) }

// wrapText breaks s at spaces so no line is wider than width, keeping
// existing newlines. A word wider than width stays on its own line.
func (c *Context) wrapText(s string, width float32) string {
	if width <= 0 {
		return s
	}
	var out strings.Builder
	for li, line := range strings.Split(s, "\n") {
		if li > 0 {
			out.WriteByte('\n')
		}
		lineW := float32(0)
		for wi, word := range strings.Split(line, " ") {
			ww := c.TextSize(word).X
			sp := c.TextSize(" ").X
			switch {
			case wi == 0:
			case lineW+sp+ww <= width:
				out.WriteByte(' ')
				lineW += sp
			default:
				out.WriteByte('\n')
				lineW = 0
			}
			out.WriteString(word)
			lineW += ww
		}
	}
	return out.String()
}

// TextWrapped is Text wrapped to the space left to the content edge.
func (c *Context) TextWrapped(s string) {
	w := c.cur
	c.Text(c.wrapText(s, w.contentRight-w.cursor.X))
}

// Bullet draws a bullet without text.
func (c *Context) Bullet() {
	lh := c.lineHeight()
	r := c.ItemAdd(Vec2{X: lh, Y: lh})
	if c.cur.skip {
		return
	}
	c.cur.DL.AddCircleFilled(Vec2{X: r.Min.X + lh*0.5, Y: r.Min.Y + lh*0.5}, lh*0.15, c.col(ColText))
}

// Text draws a possibly multi-line label.
func (c *Context) Text(s string) {
	sz := c.TextSize(s)
	r := c.ItemAdd(sz)
	if c.cur.skip {
		return
	}
	c.DrawText(r.Min, s, c.col(ColText))
}

// Textf is Text with formatting.
func (c *Context) Textf(format string, a ...any) { c.Text(fmt.Sprintf(format, a...)) }

// LabelText draws a value on the left and a label after it.
func (c *Context) LabelText(label, value string) {
	text := DisplayLabel(label)
	vs := c.TextSize(value)
	ls := c.TextSize(text)
	fp := c.Style.FramePadding
	vw := max(c.CalcItemWidth(), vs.X)
	r := c.ItemAdd(Vec2{X: vw + c.Style.ItemInnerSpacing.X + ls.X, Y: max(vs.Y, ls.Y) + 2*fp.Y})
	if c.cur.skip {
		return
	}
	c.DrawText(Vec2{X: r.Min.X, Y: r.Min.Y + fp.Y}, value, c.col(ColText))
	c.DrawText(Vec2{X: r.Min.X + vw + c.Style.ItemInnerSpacing.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
}

// BulletText draws a bullet followed by text.
func (c *Context) BulletText(s string) {
	ts := c.TextSize(s)
	lh := c.lineHeight()
	r := c.ItemAdd(Vec2{X: lh + ts.X, Y: ts.Y})
	if c.cur.skip {
		return
	}
	c.cur.DL.AddCircleFilled(Vec2{X: r.Min.X + lh*0.5, Y: r.Min.Y + lh*0.5}, lh*0.15, c.col(ColText))
	c.DrawText(Vec2{X: r.Min.X + lh, Y: r.Min.Y}, s, c.col(ColText))
}

// SeparatorText draws a rule with a label on its left.
func (c *Context) SeparatorText(label string) {
	w := c.cur
	text := DisplayLabel(label)
	ts := c.TextSize(text)
	width := max(w.contentRight-w.lineStartX(), 1)
	r := c.ItemAdd(Vec2{X: width, Y: ts.Y + c.Style.FramePadding.Y})
	if w.skip {
		return
	}
	y := r.Min.Y + ts.Y/2
	col := c.col(ColSeparator)
	gap := c.Style.ItemInnerSpacing.X
	if text == "" {
		w.DL.AddLine(Vec2{X: r.Min.X, Y: y}, Vec2{X: r.Max.X, Y: y}, col, 1)
		return
	}
	lead := float32(12)
	w.DL.AddLine(Vec2{X: r.Min.X, Y: y}, Vec2{X: r.Min.X + lead, Y: y}, col, 1)
	c.DrawText(Vec2{X: r.Min.X + lead + gap, Y: r.Min.Y}, text, c.col(ColText))
	x := r.Min.X + lead + 2*gap + ts.X
	if x < r.Max.X {
		w.DL.AddLine(Vec2{X: x, Y: y}, Vec2{X: r.Max.X, Y: y}, col, 1)
	}
}
