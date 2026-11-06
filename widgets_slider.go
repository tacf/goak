package goak

import (
	"fmt"
	"math"
	"strconv"

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

// labelAfter draws a frame-aligned label as its own item. Used by composite
// widgets that lay out several frames before the label.
func (c *Context) labelAfter(text string) {
	if text == "" {
		return
	}
	fp := c.Style.FramePadding
	ts := c.TextSize(text)
	r := c.ItemAdd(Vec2{X: ts.X, Y: ts.Y + 2*fp.Y})
	if c.cur.skip {
		return
	}
	c.DrawText(Vec2{X: r.Min.X, Y: r.Min.Y + fp.Y}, text, c.col(ColText))
}

// multi lays out n frames side by side that share CalcItemWidth, then the
// label. each draws component i with the width already set.
func (c *Context) multi(label string, n int, each func(i int) bool) bool {
	text := DisplayLabel(label)
	gap := c.Style.ItemInnerSpacing.X
	total := c.CalcItemWidth()
	w := max((total-gap*float32(n-1))/float32(n), 1)
	c.PushID(label)
	c.BeginGroup()
	changed := false
	for i := 0; i < n; i++ {
		if i > 0 {
			c.SameLine(0, gap)
		}
		c.SetNextItemWidth(w)
		if each(i) {
			changed = true
		}
	}
	if text != "" {
		c.SameLine(0, c.Style.ItemInnerSpacing.X)
		c.labelAfter(text)
	}
	c.EndGroup()
	c.PopID()
	return changed
}

func compID(i int) string { return "##" + strconv.Itoa(i) }

// SliderFloat2 edits two floats in one row.
func (c *Context) SliderFloat2(label string, v *[2]float32, lo, hi float32, format string) bool {
	return c.multi(label, 2, func(i int) bool { return c.SliderFloat(compID(i), &v[i], lo, hi, format) })
}

// SliderFloat3 edits three floats in one row.
func (c *Context) SliderFloat3(label string, v *[3]float32, lo, hi float32, format string) bool {
	return c.multi(label, 3, func(i int) bool { return c.SliderFloat(compID(i), &v[i], lo, hi, format) })
}

// SliderFloat4 edits four floats in one row.
func (c *Context) SliderFloat4(label string, v *[4]float32, lo, hi float32, format string) bool {
	return c.multi(label, 4, func(i int) bool { return c.SliderFloat(compID(i), &v[i], lo, hi, format) })
}

// SliderInt2 edits two ints in one row.
func (c *Context) SliderInt2(label string, v *[2]int, lo, hi int, format string) bool {
	return c.multi(label, 2, func(i int) bool { return c.SliderInt(compID(i), &v[i], lo, hi, format) })
}

// SliderInt3 edits three ints in one row.
func (c *Context) SliderInt3(label string, v *[3]int, lo, hi int, format string) bool {
	return c.multi(label, 3, func(i int) bool { return c.SliderInt(compID(i), &v[i], lo, hi, format) })
}

// SliderInt4 edits four ints in one row.
func (c *Context) SliderInt4(label string, v *[4]int, lo, hi int, format string) bool {
	return c.multi(label, 4, func(i int) bool { return c.SliderInt(compID(i), &v[i], lo, hi, format) })
}

// DragFloat2 edits two floats in one row.
func (c *Context) DragFloat2(label string, v *[2]float32, speed, lo, hi float32, format string) bool {
	return c.multi(label, 2, func(i int) bool { return c.DragFloat(compID(i), &v[i], speed, lo, hi, format) })
}

// DragFloat3 edits three floats in one row.
func (c *Context) DragFloat3(label string, v *[3]float32, speed, lo, hi float32, format string) bool {
	return c.multi(label, 3, func(i int) bool { return c.DragFloat(compID(i), &v[i], speed, lo, hi, format) })
}

// DragFloat4 edits four floats in one row.
func (c *Context) DragFloat4(label string, v *[4]float32, speed, lo, hi float32, format string) bool {
	return c.multi(label, 4, func(i int) bool { return c.DragFloat(compID(i), &v[i], speed, lo, hi, format) })
}

// DragFloatRange2 edits a min and a max that never cross.
func (c *Context) DragFloatRange2(label string, vmin, vmax *float32, speed, lo, hi float32, format string) bool {
	return c.multi(label, 2, func(i int) bool {
		if i == 0 {
			ch := c.DragFloat(compID(0), vmin, speed, lo, hi, format)
			if *vmin > *vmax {
				*vmin = *vmax
			}
			return ch
		}
		ch := c.DragFloat(compID(1), vmax, speed, lo, hi, format)
		if *vmax < *vmin {
			*vmax = *vmin
		}
		return ch
	})
}

// DragInt changes *v by speed per pixel of horizontal mouse motion. Pass
// lo == hi for no limits. format defaults to "%d".
func (c *Context) DragInt(label string, v *int, speed float32, lo, hi int, format string) bool {
	if v == nil {
		return false
	}
	if format == "" {
		format = "%d"
	}
	id := c.ItemID(label)
	frame, text, ok := c.sliderFrame(label)
	if !ok {
		return false
	}
	s := c.ButtonBehaviorEx(frame, id, 0)
	accKey := hashStr("#DRAGACC", id)
	acc := StateGet(c.States, accKey, float64(*v))
	changed := false
	if s.Started {
		acc = float64(*v)
		c.dragOffset.X = c.mouse().X
	}
	if c.activeID == id && s.Held && !s.Started {
		mx := c.mouse().X
		acc += float64(mx-c.dragOffset.X) * float64(speed)
		c.dragOffset.X = mx
		if lo != hi {
			acc = math.Min(math.Max(acc, float64(lo)), float64(hi))
		}
		nv := int(math.Round(acc))
		if nv != *v {
			*v, changed = nv, true
		}
	}
	StateSet(c.States, accKey, acc)
	c.drawFrame(frame, c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive), c.focusID == id)
	c.centerValue(frame, fmt.Sprintf(format, *v))
	c.frameLabel(frame, text)
	return changed
}

// SliderAngle edits an angle stored in radians and shown in degrees.
func (c *Context) SliderAngle(label string, rad *float32, degMin, degMax float32, format string) bool {
	if rad == nil {
		return false
	}
	if format == "" {
		format = "%.0f deg"
	}
	deg := *rad * 180 / math.Pi
	if !c.SliderFloat(label, &deg, degMin, degMax, format) {
		return false
	}
	*rad = deg * math.Pi / 180
	return true
}

// vslider is the shared vertical slider behavior. t is the normalized value.
func (c *Context) vslider(label string, size Vec2, t float32, value string) (nt float32, moved bool) {
	id := c.ItemID(label)
	r := c.ItemAdd(size)
	if c.cur.skip {
		return t, false
	}
	s := c.ButtonBehaviorEx(r, id, 0)
	nt = t
	if c.activeID == id && s.Held {
		g := c.Style.GrabMinSize
		span := r.Max.Y - r.Min.Y - g
		if span > 0 {
			nt = 1 - min(max((c.mouse().Y-r.Min.Y-g/2)/span, 0), 1)
		}
	}
	c.drawFrame(r, c.frameColor(s, id, ColFrameBg, ColFrameBgHovered, ColFrameBgActive), c.focusID == id)
	g := c.Style.GrabMinSize
	gy := r.Min.Y + (r.Max.Y-r.Min.Y-g)*(1-nt)
	gc := ColSliderGrab
	if c.activeID == id {
		gc = ColSliderGrabActive
	}
	c.cur.DL.AddRectFilled(Vec2{X: r.Min.X + 2, Y: gy + 2}, Vec2{X: r.Max.X - 2, Y: gy + g - 2}, c.col(gc), c.Style.GrabRounding)
	ts := c.TextSize(value)
	c.DrawText(Vec2{X: r.Min.X + (r.Max.X-r.Min.X-ts.X)/2, Y: r.Max.Y - ts.Y - 2}, value, c.col(ColText))
	return nt, nt != t
}

// VSliderFloat is a vertical slider. The label only scopes the ID.
func (c *Context) VSliderFloat(label string, size Vec2, v *float32, lo, hi float32, format string) bool {
	if v == nil {
		return false
	}
	if format == "" {
		format = "%.3f"
	}
	t := float32(0)
	if hi != lo && mathx.Finite32(*v) && mathx.Finite32(lo) && mathx.Finite32(hi) {
		t = min(max((*v-lo)/(hi-lo), 0), 1)
	}
	nt, moved := c.vslider(label, size, t, fmt.Sprintf(format, *v))
	if !moved || hi == lo {
		return false
	}
	nv := lo + nt*(hi-lo)
	if !mathx.Finite32(nv) || nv == *v {
		return false
	}
	*v = nv
	return true
}

// VSliderInt is a vertical int slider. The label only scopes the ID.
func (c *Context) VSliderInt(label string, size Vec2, v *int, lo, hi int, format string) bool {
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
	nt, moved := c.vslider(label, size, t, fmt.Sprintf(format, *v))
	if !moved || hi == lo {
		return false
	}
	nv := lo + int(math.Round(float64(nt)*float64(hi-lo)))
	if nv == *v {
		return false
	}
	*v = nv
	return true
}
