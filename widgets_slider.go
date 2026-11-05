package goak

import (
	"fmt"
	"math"

	"github.com/tacf/goak/internal/mathx"
)

// sliderFrame lays out the frame and label, and returns the frame rect.
func (c *Context) sliderFrame(label string) (frame Rect, text string, ok bool) {
	text = DisplayLabel(label)
	w := c.CalcItemWidth()
	h := c.frameHeight()
	total := w
	if text != "" {
		total += c.Style.ItemInnerSpacing.X + c.TextSize(text).X
	}
	r := c.ItemAdd(Vec2{X: total, Y: h})
	frame = Rect{Min: r.Min, Max: Vec2{X: r.Min.X + w, Y: r.Max.Y}}
	return frame, text, !c.cur.skip
}

func (c *Context) drawFrame(r Rect, col Col, focused bool) {
	dl := c.cur.DL
	dl.AddRectFilled(r.Min, r.Max, c.col(col), c.Style.FrameRounding)
	bc := ColBorder
	if focused {
		bc = ColCheckMark
	}
	dl.AddRect(r.Min, r.Max, c.col(bc), c.Style.FrameRounding, 1)
}

func (c *Context) frameLabel(frame Rect, text string) {
	if text != "" {
		c.DrawText(Vec2{X: frame.Max.X + c.Style.ItemInnerSpacing.X, Y: frame.Min.Y + c.Style.FramePadding.Y}, text, c.col(ColText))
	}
}

// sliderT maps a mouse x to 0..1 along the grab travel of frame.
func (c *Context) sliderT(frame Rect) float32 {
	g := c.Style.GrabMinSize
	span := frame.Max.X - frame.Min.X - g
	if span <= 0 {
		return 0
	}
	t := (c.mouse().X - frame.Min.X - g/2) / span
	return min(max(t, 0), 1)
}

// slider runs the shared behavior. It returns the new normalized value and
// whether the user moved it.
func (c *Context) slider(label string, t float32, step float32) (frame Rect, text string, nt float32, moved, drawn bool) {
	id := c.ItemID(label)
	frame, text, ok := c.sliderFrame(label)
	if !ok {
		return frame, text, t, false, false
	}
	s := c.ButtonBehaviorEx(frame, id, 0)
	nt = t
	if c.activeID == id && s.Held {
		nt = c.sliderT(frame)
	}
	if c.focusID == id && step != 0 {
		if c.In.KeyPressed[KeyLeft] {
			nt = min(max(t-step, 0), 1)
		}
		if c.In.KeyPressed[KeyRight] {
			nt = min(max(t+step, 0), 1)
		}
	}
	moved = nt != t
	col := c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive)
	c.drawFrame(frame, col, c.focusID == id)
	g := c.Style.GrabMinSize
	gx := frame.Min.X + (frame.Max.X-frame.Min.X-g)*nt
	gc := ColSliderGrab
	if c.activeID == id {
		gc = ColSliderGrabActive
	}
	c.cur.DL.AddRectFilled(Vec2{X: gx + 2, Y: frame.Min.Y + 2}, Vec2{X: gx + g - 2, Y: frame.Max.Y - 2}, c.col(gc), c.Style.GrabRounding)
	c.frameLabel(frame, text)
	return frame, text, nt, moved, true
}

func (c *Context) centerValue(frame Rect, s string) {
	ts := c.TextSize(s)
	c.DrawText(Vec2{X: frame.Min.X + (frame.Max.X-frame.Min.X-ts.X)/2, Y: frame.Min.Y + c.Style.FramePadding.Y}, s, c.col(ColText))
}

// SliderFloat edits *v within [lo, hi]. format defaults to "%.3f". It
// returns true when the value changed. Left and Right nudge a focused slider.
func (c *Context) SliderFloat(label string, v *float32, lo, hi float32, format string) bool {
	if v == nil {
		return false
	}
	if format == "" {
		format = "%.3f"
	}
	valid := mathx.Finite32(lo) && mathx.Finite32(hi) && hi != lo
	t := float32(0)
	if valid && mathx.Finite32(*v) {
		t = min(max((*v-lo)/(hi-lo), 0), 1)
	}
	frame, _, nt, moved, drawn := c.slider(label, t, 0.01)
	changed := false
	if moved && valid {
		nv := lo + nt*(hi-lo)
		if mathx.Finite32(nv) && nv != *v {
			*v, changed = nv, true
		}
	}
	if drawn {
		c.centerValue(frame, fmt.Sprintf(format, *v))
	}
	return changed
}

// SliderInt edits *v within [lo, hi]. format defaults to "%d".
func (c *Context) SliderInt(label string, v *int, lo, hi int, format string) bool {
	if v == nil {
		return false
	}
	if format == "" {
		format = "%d"
	}
	t := float32(0)
	if hi != lo {
		t = min(max(float32(*v-lo)/float32(hi-lo), 0), 1)
	}
	step := float32(0)
	if hi != lo {
		step = 1 / float32(hi-lo)
	}
	frame, _, nt, moved, drawn := c.slider(label, t, step)
	changed := false
	if moved {
		nv := lo + int(math.Round(float64(nt)*float64(hi-lo)))
		if nv != *v {
			*v, changed = nv, true
		}
	}
	if drawn {
		c.centerValue(frame, fmt.Sprintf(format, *v))
	}
	return changed
}

// DragFloat changes *v by speed per pixel of horizontal mouse motion.
// Pass lo == hi for no limits. format defaults to "%.3f".
func (c *Context) DragFloat(label string, v *float32, speed, lo, hi float32, format string) bool {
	if v == nil {
		return false
	}
	if format == "" {
		format = "%.3f"
	}
	id := c.ItemID(label)
	frame, text, ok := c.sliderFrame(label)
	if !ok {
		return false
	}
	s := c.ButtonBehaviorEx(frame, id, 0)
	changed := false
	if s.Started {
		c.dragOffset.X = c.mouse().X
	}
	if c.activeID == id && s.Held && !s.Started {
		mx := c.mouse().X
		base := *v
		if !mathx.Finite32(base) {
			base = 0
			if mathx.Finite32(lo) {
				base = lo
			}
		}
		nv := base + (mx-c.dragOffset.X)*speed
		c.dragOffset.X = mx
		if lo != hi {
			nv = min(max(nv, lo), hi)
		}
		if mathx.Finite32(nv) && nv != *v {
			*v, changed = nv, true
		}
	}
	c.drawFrame(frame, c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive), c.focusID == id)
	c.centerValue(frame, fmt.Sprintf(format, *v))
	c.frameLabel(frame, text)
	return changed
}
