package goak

import (
	"fmt"
	"math"

	"github.com/tacf/goak/colors"
)

// PlotFlags tune BeginPlot.
type PlotFlags uint8

// Plot flags.
const (
	// PlotNoLegend hides the legend.
	PlotNoLegend PlotFlags = 1 << iota
	// PlotNoInteract disables panning, zooming and legend clicks.
	PlotNoInteract
	// PlotNoGrid hides the grid lines.
	PlotNoGrid
)

// PlotAxis selects an axis of a plot.
type PlotAxis uint8

// Plot axes.
const (
	AxisX PlotAxis = iota
	AxisY
)

type plotKind uint8

const (
	plotLine plotKind = iota
	plotScatter
	plotBars
)

type plotItem struct {
	label  string
	kind   plotKind
	xs, ys []float64
	width  float64 // bar width in x units
	col    colors.Color
	id     ID
}

type plotAxis struct {
	lo, hi float64
	fit    bool // follow the data
	label  string
}

// plotState is kept between frames.
type plotState struct {
	axes    [2]plotAxis
	hidden  map[ID]bool
	limited [2]bool // limits were set by PlotSetupAxisLimits at least once
	last    Vec2    // mouse at the previous frame while panning
	started bool
}

type plotBuild struct {
	title string
	id    ID
	frame Rect
	flags PlotFlags
	st    *plotState
	items []plotItem
	// Limits requested this frame.
	setLimits [2]bool
	limits    [2][2]float64
	limCond   [2]Cond
}

var plotPalette = [...]colors.Color{
	colors.RGB(76, 157, 255), colors.RGB(255, 127, 14), colors.RGB(44, 190, 90), colors.RGB(230, 60, 60),
	colors.RGB(160, 110, 210), colors.RGB(160, 100, 85), colors.RGB(235, 125, 200), colors.RGB(200, 200, 50),
}

// BeginPlot starts a plot. A size of 0 uses the available width and 250 of
// height. It returns false when the window is hidden, then EndPlot must not
// be called. Submit PlotLine, PlotScatter and PlotBars items, then EndPlot.
// Axes fit the data until the user pans or zooms. Double click fits again.
func (c *Context) BeginPlot(title string, size Vec2, flags PlotFlags) bool {
	w := c.cur
	if w == nil || w.skip {
		return false
	}
	if size.X <= 0 {
		size.X = max(c.ContentRegionAvail().X+size.X, 1)
	}
	if size.Y <= 0 {
		size.Y = 250
	}
	id := c.GetID(title)
	key := hashStr("#PLOT", id)
	st := StateGet[*plotState](c.States, key, nil)
	if st == nil {
		st = &plotState{hidden: map[ID]bool{}}
		st.axes[AxisX].fit, st.axes[AxisY].fit = true, true
	}
	StateSetKeep(c.States, key, st, LongAge)
	r := c.ItemAdd(size)
	c.plots = append(c.plots, &plotBuild{title: title, id: id, frame: r, flags: flags, st: st})
	return true
}

func (c *Context) curPlot() *plotBuild {
	if len(c.plots) == 0 {
		panic("goak: plot call outside BeginPlot")
	}
	return c.plots[len(c.plots)-1]
}

// PlotSetupAxes names the axes.
func (c *Context) PlotSetupAxes(xLabel, yLabel string) {
	p := c.curPlot()
	p.st.axes[AxisX].label, p.st.axes[AxisY].label = xLabel, yLabel
}

// PlotSetupAxisLimits fixes the range of an axis. With CondOnce it only sets
// the first range, the user can still pan and zoom. With CondAlways the range
// is held every frame.
func (c *Context) PlotSetupAxisLimits(axis PlotAxis, lo, hi float64, cond Cond) {
	p := c.curPlot()
	p.setLimits[axis], p.limits[axis], p.limCond[axis] = true, [2]float64{lo, hi}, cond
}

func (c *Context) plotAdd(label string, kind plotKind, xs, ys []float64, width float64) {
	p := c.curPlot()
	n := len(p.items)
	it := plotItem{label: label, kind: kind, xs: xs, ys: ys, width: width,
		col: plotPalette[n%len(plotPalette)], id: hashStr(label, p.id)}
	p.items = append(p.items, it)
}

// PlotLine draws a line through the points (xs[i], ys[i]).
func (c *Context) PlotLine(label string, xs, ys []float64) {
	c.plotAdd(label, plotLine, xs, ys, 0)
}

// PlotLineY draws ys against their index.
func (c *Context) PlotLineY(label string, ys []float64) { c.plotAdd(label, plotLine, nil, ys, 0) }

// PlotScatter draws a marker at each point.
func (c *Context) PlotScatter(label string, xs, ys []float64) {
	c.plotAdd(label, plotScatter, xs, ys, 0)
}

// PlotBars draws a bar of the given width in x units from zero to ys[i] at
// xs[i]. A nil xs places the bars at their index.
func (c *Context) PlotBars(label string, xs, ys []float64, width float64) {
	if width <= 0 {
		width = 0.67
	}
	c.plotAdd(label, plotBars, xs, ys, width)
}

func (it *plotItem) point(i int) (x, y float64) {
	x = float64(i)
	if it.xs != nil {
		x = it.xs[i]
	}
	return x, it.ys[i]
}

func (it *plotItem) count() int {
	if it.xs != nil {
		return min(len(it.xs), len(it.ys))
	}
	return len(it.ys)
}

func finite64(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// fitAxes sets the ranges of axes that follow the data.
func (p *plotBuild) fitAxes() {
	var lo, hi [2]float64
	has := [2]bool{}
	grow := func(a PlotAxis, v float64) {
		if !finite64(v) {
			return
		}
		if !has[a] {
			lo[a], hi[a], has[a] = v, v, true
			return
		}
		lo[a], hi[a] = math.Min(lo[a], v), math.Max(hi[a], v)
	}
	for i := range p.items {
		it := &p.items[i]
		if p.st.hidden[it.id] {
			continue
		}
		for k := 0; k < it.count(); k++ {
			x, y := it.point(k)
			grow(AxisX, x)
			grow(AxisY, y)
			if it.kind == plotBars {
				grow(AxisX, x-it.width/2)
				grow(AxisX, x+it.width/2)
				grow(AxisY, 0)
			}
		}
	}
	for a := AxisX; a <= AxisY; a++ {
		ax := &p.st.axes[a]
		if !ax.fit {
			continue
		}
		if !has[a] {
			ax.lo, ax.hi = 0, 1
			continue
		}
		if hi[a] <= lo[a] {
			lo[a]--
			hi[a]++
		}
		pad := (hi[a] - lo[a]) * 0.05
		ax.lo, ax.hi = lo[a]-pad, hi[a]+pad
	}
}

// applyLimits applies PlotSetupAxisLimits requests.
func (p *plotBuild) applyLimits() {
	for a := AxisX; a <= AxisY; a++ {
		if !p.setLimits[a] {
			continue
		}
		if p.limCond[a] == CondOnce && p.st.limited[a] {
			continue
		}
		p.st.limited[a] = true
		ax := &p.st.axes[a]
		ax.lo, ax.hi, ax.fit = p.limits[a][0], p.limits[a][1], false
		if ax.hi <= ax.lo {
			ax.hi = ax.lo + 1
		}
	}
}

// niceTicks returns round tick positions inside [lo, hi] and the decimals
// needed to print them.
func niceTicks(lo, hi float64, maxTicks int) (ticks []float64, decimals int) {
	span := hi - lo
	if span <= 0 || maxTicks < 2 || !finite64(span) {
		return nil, 0
	}
	raw := span / float64(maxTicks)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	step := mag
	switch norm := raw / mag; {
	case norm > 5:
		step = 10 * mag
	case norm > 2:
		step = 5 * mag
	case norm > 1:
		step = 2 * mag
	}
	for v := math.Ceil(lo/step) * step; v <= hi+step*1e-9; v += step {
		ticks = append(ticks, math.Round(v/step)*step+0) // +0 turns -0 into 0
	}
	decimals = max(0, min(8, int(math.Ceil(-math.Log10(step)-1e-9))))
	return ticks, decimals
}

// EndPlot draws the plot and handles its interaction.
func (c *Context) EndPlot() {
	p := c.curPlot()
	c.plots = c.plots[:len(c.plots)-1]
	st := p.st
	w := c.cur
	dl := w.DL
	lh := c.lineHeight()
	pad := c.Style.FramePadding

	p.applyLimits()
	p.fitAxes()

	// Layout: title on top, tick labels left and below, axis labels outside.
	frame := p.frame
	dl.AddRectFilled(frame.Min, frame.Max, c.col(ColFrameBg), c.Style.FrameRounding)
	top := frame.Min.Y + pad.Y
	if p.title != "" && p.title[0] != '#' {
		ts := c.TextSize(DisplayLabel(p.title))
		c.DrawText(Vec2{X: (frame.Min.X + frame.Max.X - ts.X) / 2, Y: top}, DisplayLabel(p.title), c.col(ColText))
		top += lh + pad.Y
	}
	left := frame.Min.X + pad.X
	if st.axes[AxisY].label != "" {
		left += lh // room for the label drawn above the axis instead
	}
	bottom := frame.Max.Y - pad.Y - lh
	if st.axes[AxisX].label != "" {
		bottom -= lh
	}
	left += c.plotLabelWidth(st, top, bottom)
	inner := Rect{Min: Vec2{X: left, Y: top}, Max: Vec2{X: frame.Max.X - pad.X, Y: bottom}}
	if inner.Max.X-inner.Min.X < 20 || inner.Max.Y-inner.Min.Y < 20 {
		return
	}

	// Interaction. The legend sits on top of the plot area so it goes first.
	var legendHover []bool
	if p.flags&PlotNoLegend == 0 {
		legendHover = c.legendInteract(p, inner)
	}
	if p.flags&PlotNoInteract == 0 {
		c.plotInteract(p, inner)
	}
	xa, ya := st.axes[AxisX], st.axes[AxisY]
	sx := float64(inner.Max.X-inner.Min.X) / (xa.hi - xa.lo)
	sy := float64(inner.Max.Y-inner.Min.Y) / (ya.hi - ya.lo)
	px := func(x float64) float32 { return inner.Min.X + float32((x-xa.lo)*sx) }
	py := func(y float64) float32 { return inner.Max.Y - float32((y-ya.lo)*sy) }

	// Grid and ticks.
	xt, xd := niceTicks(xa.lo, xa.hi, int((inner.Max.X-inner.Min.X)/70))
	yt, yd := niceTicks(ya.lo, ya.hi, int((inner.Max.Y-inner.Min.Y)/40))
	grid := c.col(ColTableBorder)
	grid.A = grid.A / 2
	for _, v := range xt {
		x := px(v)
		if p.flags&PlotNoGrid == 0 {
			dl.AddLine(Vec2{X: x, Y: inner.Min.Y}, Vec2{X: x, Y: inner.Max.Y}, grid, 1)
		}
		s := fmt.Sprintf("%.*f", xd, v)
		c.DrawText(Vec2{X: x - c.TextSize(s).X/2, Y: inner.Max.Y + pad.Y/2}, s, c.col(ColText))
	}
	for _, v := range yt {
		y := py(v)
		if p.flags&PlotNoGrid == 0 {
			dl.AddLine(Vec2{X: inner.Min.X, Y: y}, Vec2{X: inner.Max.X, Y: y}, grid, 1)
		}
		s := fmt.Sprintf("%.*f", yd, v)
		c.DrawText(Vec2{X: inner.Min.X - c.TextSize(s).X - 4, Y: y - lh/2}, s, c.col(ColText))
	}
	if l := xa.label; l != "" {
		c.DrawText(Vec2{X: (inner.Min.X+inner.Max.X)/2 - c.TextSize(l).X/2, Y: frame.Max.Y - pad.Y - lh}, l, c.col(ColText))
	}
	if l := ya.label; l != "" {
		c.DrawText(Vec2{X: frame.Min.X + pad.X, Y: top - lh*0.2}, l, c.col(ColText))
	}

	// Data.
	dl.PushClipRect(inner.Intersect(w.clip))
	width := int(inner.Max.X - inner.Min.X)
	for i := range p.items {
		it := &p.items[i]
		if st.hidden[it.id] {
			continue
		}
		c.drawPlotItem(it, px, py, width, inner)
	}
	dl.PopClipRect()
	dl.AddRect(inner.Min, inner.Max, c.col(ColBorder), 0, 1)

	if p.flags&PlotNoLegend == 0 {
		c.legendDraw(p, inner, legendHover)
	}
	if c.mouseOver(inner) && p.flags&PlotNoInteract == 0 {
		m := c.mouse()
		s := fmt.Sprintf("%.*f, %.*f", max(xd, 2), xa.lo+float64(m.X-inner.Min.X)/sx, max(yd, 2), ya.hi-float64(m.Y-inner.Min.Y)/sy)
		ts := c.TextSize(s)
		c.DrawText(Vec2{X: inner.Max.X - ts.X - 4, Y: inner.Max.Y - ts.Y - 2}, s, c.col(ColTextDisabled))
	}
}

// plotLabelWidth measures the widest y tick label for the current limits.
func (c *Context) plotLabelWidth(st *plotState, top, bottom float32) float32 {
	ya := st.axes[AxisY]
	ticks, dec := niceTicks(ya.lo, ya.hi, int((bottom-top)/40))
	var wmax float32
	for _, v := range ticks {
		wmax = max(wmax, c.TextSize(fmt.Sprintf("%.*f", dec, v)).X)
	}
	return wmax + 8
}

func (c *Context) drawPlotItem(it *plotItem, px, py func(float64) float32, width int, inner Rect) {
	dl := c.cur.DL
	col := c.fade(it.col)
	n := it.count()
	switch it.kind {
	case plotLine:
		stride := max(1, n/max(2*width, 1))
		var prev Vec2
		have := false
		for i := 0; i < n; i += stride {
			x, y := it.point(i)
			if !finite64(x) || !finite64(y) {
				have = false
				continue
			}
			p := Vec2{X: px(x), Y: py(y)}
			if have {
				dl.AddLine(prev, p, col, 1.5)
			}
			prev, have = p, true
		}
	case plotScatter:
		for i := 0; i < n; i++ {
			x, y := it.point(i)
			if !finite64(x) || !finite64(y) {
				continue
			}
			p := Vec2{X: px(x), Y: py(y)}
			if p.X < inner.Min.X-4 || p.X > inner.Max.X+4 || p.Y < inner.Min.Y-4 || p.Y > inner.Max.Y+4 {
				continue
			}
			dl.AddRectFilled(Vec2{X: p.X - 2.5, Y: p.Y - 2.5}, Vec2{X: p.X + 2.5, Y: p.Y + 2.5}, col, 0)
		}
	case plotBars:
		base := py(0)
		for i := 0; i < n; i++ {
			x, y := it.point(i)
			if !finite64(x) || !finite64(y) {
				continue
			}
			x0, x1 := px(x-it.width/2), px(x+it.width/2)
			y0 := py(y)
			dl.AddRectFilled(Vec2{X: x0, Y: min(y0, base)}, Vec2{X: max(x1, x0+1), Y: max(y0, base) + 1}, col, 0)
		}
	}
}

// plotInteract handles panning, zooming and fitting.
func (c *Context) plotInteract(p *plotBuild, inner Rect) {
	st := p.st
	in := c.In
	id := hashUint(0x91, 0, p.id)
	s := c.ButtonBehaviorEx(inner, id, BtnNoFocus)
	xa, ya := &st.axes[AxisX], &st.axes[AxisY]
	if s.Started && in.MouseClicks[MouseLeft] >= 2 {
		xa.fit, ya.fit = true, true
		for a := range st.limited {
			st.limited[a] = false
		}
	}
	m := c.mouse()
	if s.Started {
		st.last, st.started = m, true
	}
	if c.activeID == id && s.Held && st.started {
		dx, dy := float64(m.X-st.last.X), float64(m.Y-st.last.Y)
		st.last = m
		if dx != 0 || dy != 0 {
			w := float64(inner.Max.X - inner.Min.X)
			h := float64(inner.Max.Y - inner.Min.Y)
			shiftX, shiftY := -dx/w*(xa.hi-xa.lo), dy/h*(ya.hi-ya.lo)
			xa.lo, xa.hi, xa.fit = xa.lo+shiftX, xa.hi+shiftX, false
			ya.lo, ya.hi, ya.fit = ya.lo+shiftY, ya.hi+shiftY, false
		}
	}
	if s.Hovered && !c.wheelUsed && in.Wheel[1] != 0 {
		c.wheelUsed = true
		f := math.Pow(0.9, float64(in.Wheel[1]))
		fx := float64(m.X-inner.Min.X) / float64(inner.Max.X-inner.Min.X)
		fy := float64(inner.Max.Y-m.Y) / float64(inner.Max.Y-inner.Min.Y)
		zoom := func(a *plotAxis, frac float64) {
			span := a.hi - a.lo
			at := a.lo + frac*span
			a.lo, a.hi, a.fit = at-frac*span*f, at+(1-frac)*span*f, false
		}
		zoom(xa, fx)
		zoom(ya, fy)
	}
}

// legendGeometry returns the legend box and one row rect per item.
func (c *Context) legendGeometry(p *plotBuild, inner Rect) (Rect, []Rect) {
	if len(p.items) == 0 {
		return Rect{}, nil
	}
	rowH := c.lineHeight() + 2
	var wmax float32
	for _, it := range p.items {
		wmax = max(wmax, c.TextSize(DisplayLabel(it.label)).X)
	}
	x1 := inner.Max.X - 6
	y := inner.Min.Y + 6
	box := Rect{Min: Vec2{X: x1 - wmax - 28, Y: y}, Max: Vec2{X: x1, Y: y + rowH*float32(len(p.items)) + 4}}
	rows := make([]Rect, len(p.items))
	for i := range rows {
		rows[i] = Rect{Min: Vec2{X: box.Min.X, Y: box.Min.Y + 2 + rowH*float32(i)}, Max: Vec2{X: box.Max.X, Y: box.Min.Y + 2 + rowH*float32(i+1)}}
	}
	return box, rows
}

// legendInteract toggles items clicked in the legend and reports which rows
// are hovered. It runs before the plot area takes the mouse.
func (c *Context) legendInteract(p *plotBuild, inner Rect) []bool {
	_, rows := c.legendGeometry(p, inner)
	hov := make([]bool, len(rows))
	if p.flags&PlotNoInteract != 0 {
		return hov
	}
	for i, row := range rows {
		s := c.ButtonBehaviorEx(row, hashUint(0x92, uint64(i), p.id), BtnNoFocus)
		if s.Pressed {
			it := &p.items[i]
			p.st.hidden[it.id] = !p.st.hidden[it.id]
		}
		hov[i] = s.Hovered
	}
	return hov
}

// legendDraw draws the legend in the top right corner.
func (c *Context) legendDraw(p *plotBuild, inner Rect, hov []bool) {
	box, rows := c.legendGeometry(p, inner)
	if len(rows) == 0 {
		return
	}
	dl := c.cur.DL
	bg := c.col(ColPopupBg)
	bg.A = 200
	dl.AddRectFilled(box.Min, box.Max, bg, 3)
	for i := range p.items {
		it, row := &p.items[i], rows[i]
		if hov[i] {
			dl.AddRectFilled(row.Min, row.Max, c.col(ColHeaderHovered), 2)
		}
		col, tcol := it.col, c.col(ColText)
		if p.st.hidden[it.id] {
			col, tcol = c.col(ColTextDisabled), c.col(ColTextDisabled)
		}
		cy := (row.Min.Y + row.Max.Y) / 2
		dl.AddRectFilled(Vec2{X: row.Min.X + 6, Y: cy - 3}, Vec2{X: row.Min.X + 18, Y: cy + 3}, col, 1)
		c.DrawText(Vec2{X: row.Min.X + 22, Y: row.Min.Y + 1}, DisplayLabel(it.label), tcol)
	}
}
