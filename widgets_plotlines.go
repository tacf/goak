package goak

import (
	"fmt"
	"math"

	"github.com/tacf/goak/internal/mathx"
)

// AutoScale as a plot scale bound means take it from the data.
const AutoScale = float32(math.MaxFloat32)

// plotRange returns the value range of values. Bounds other than AutoScale
// are used as given. Non-finite samples are ignored.
func plotRange(values []float32, lo, hi float32) (float32, float32) {
	if lo == AutoScale || hi == AutoScale {
		dlo, dhi := float32(math.MaxFloat32), float32(-math.MaxFloat32)
		for _, v := range values {
			if mathx.Finite32(v) {
				dlo, dhi = min(dlo, v), max(dhi, v)
			}
		}
		if dlo > dhi {
			dlo, dhi = 0, 1
		}
		if lo == AutoScale {
			lo = dlo
		}
		if hi == AutoScale {
			hi = dhi
		}
	}
	if hi <= lo {
		hi = lo + 1
	}
	return lo, hi
}

// plotFrame reserves the plot frame and draws the background, label and
// overlay. It returns the inner rect, or ok false when the plot is skipped.
func (c *Context) plotFrame(label string, size Vec2, overlay string) (inner Rect, hovered, ok bool) {
	text := DisplayLabel(label)
	if size.X <= 0 {
		size.X = c.CalcItemWidth()
	}
	if size.Y <= 0 {
		size.Y = 3*c.lineHeight() + 2*c.Style.FramePadding.Y
	}
	total := size.X
	if text != "" {
		total += c.Style.ItemInnerSpacing.X + c.TextSize(text).X
	}
	id := c.ItemID(label)
	r := c.ItemAdd(Vec2{X: total, Y: size.Y})
	if c.cur.skip {
		return Rect{}, false, false
	}
	frame := Rect{Min: r.Min, Max: Vec2{X: r.Min.X + size.X, Y: r.Max.Y}}
	s := c.ButtonBehaviorEx(frame, id, BtnNoFocus)
	c.drawFrame(frame, ColFrameBg, false)
	c.frameLabel(frame, text)
	pad := c.Style.FramePadding
	inner = Rect{Min: frame.Min.Add(Vec2{X: 1, Y: 1}), Max: frame.Max.Sub(Vec2{X: 1, Y: 1})}
	if overlay != "" {
		ts := c.TextSize(overlay)
		c.DrawText(Vec2{X: frame.Min.X + (size.X-ts.X)/2, Y: frame.Min.Y + pad.Y}, overlay, c.col(ColText))
	}
	return inner, s.Hovered, true
}

// plot draws values as a line (histogram false) or bars.
func (c *Context) plot(label string, values []float32, offset int, overlay string, lo, hi float32, size Vec2, histogram bool) {
	inner, hovered, ok := c.plotFrame(label, size, overlay)
	n := len(values)
	if !ok || n == 0 {
		return
	}
	lo, hi = plotRange(values, lo, hi)
	at := func(i int) float32 { return values[((i+offset)%n+n)%n] }
	w, h := inner.Max.X-inner.Min.X, inner.Max.Y-inner.Min.Y
	ypos := func(v float32) float32 {
		t := (min(max(v, lo), hi) - lo) / (hi - lo)
		if !mathx.Finite32(t) {
			t = 0
		}
		return inner.Max.Y - t*h
	}
	hot := -1
	if hovered {
		mx := c.mouse().X
		if histogram {
			hot = int((mx - inner.Min.X) / w * float32(n))
		} else if n > 1 {
			hot = int(math.Round(float64((mx - inner.Min.X) / w * float32(n-1))))
		} else {
			hot = 0
		}
		hot = min(max(hot, 0), n-1)
	}
	dl := c.cur.DL
	dl.PushClipRect(inner.Intersect(c.cur.clip))
	if histogram {
		bw := w / float32(n)
		base := ypos(max(lo, min(0, hi)))
		for i := 0; i < n; i++ {
			col := c.col(ColPlotHistogram)
			if i == hot {
				col = c.col(ColPlotHistogramHovered)
			}
			x0 := inner.Min.X + float32(i)*bw
			y := ypos(at(i))
			dl.AddRectFilled(Vec2{X: x0 + 0.5, Y: min(y, base)}, Vec2{X: x0 + bw - 0.5, Y: max(y, base) + 1}, c.fade(col), 0)
		}
	} else {
		step := float32(0)
		if n > 1 {
			step = w / float32(n-1)
		}
		pos := func(i int) Vec2 { return Vec2{X: inner.Min.X + float32(i)*step, Y: ypos(at(i))} }
		for i := 0; i+1 < n; i++ {
			col := c.col(ColPlotLines)
			if i == hot || i+1 == hot {
				col = c.col(ColPlotLinesHovered)
			}
			dl.AddLine(pos(i), pos(i+1), c.fade(col), 1.5)
		}
		if hot >= 0 {
			dl.AddCircleFilled(pos(hot), 3, c.fade(c.col(ColPlotLinesHovered)))
		}
	}
	dl.PopClipRect()
	if hot >= 0 {
		c.SetTooltip(fmt.Sprintf("%d: %.3f", hot, at(hot)))
	}
}

// PlotLines draws values as a line graph. offset rotates the start of the
// ring buffer. Pass AutoScale for scaleMin or scaleMax to fit the data. A
// zero size uses the item width and three lines of height.
func (c *Context) PlotLines(label string, values []float32, offset int, overlay string, scaleMin, scaleMax float32, size Vec2) {
	c.plot(label, values, offset, overlay, scaleMin, scaleMax, size, false)
}

// PlotHistogram draws values as bars. See PlotLines for the arguments.
func (c *Context) PlotHistogram(label string, values []float32, offset int, overlay string, scaleMin, scaleMax float32, size Vec2) {
	c.plot(label, values, offset, overlay, scaleMin, scaleMax, size, true)
}
