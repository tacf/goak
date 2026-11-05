package draw

import (
	"math"

	"github.com/tacf/goak/colors"
)

// Corner flags select which rect corners are rounded.
type Corner uint8

const (
	CornerTopLeft Corner = 1 << iota
	CornerTopRight
	CornerBottomRight
	CornerBottomLeft
	CornerAll = CornerTopLeft | CornerTopRight | CornerBottomRight | CornerBottomLeft
)

// Pack converts a color to the vertex format (R in the low byte).
func Pack(c colors.Color) uint32 {
	return uint32(c.R) | uint32(c.G)<<8 | uint32(c.B)<<16 | uint32(c.A)<<24
}

// Unpack converts a packed vertex color back to colors.Color.
func Unpack(p uint32) colors.Color {
	return colors.RGBA(uint8(p), uint8(p>>8), uint8(p>>16), uint8(p>>24))
}

// DrawList accumulates geometry and batches it into commands.
type DrawList struct {
	Vertices []Vertex
	Indices  []uint32
	Cmds     []Cmd

	// AAWidth is the antialias fringe width in logical units (1/scale).
	AAWidth float32
	// Scale is pixels per logical unit. When set, filled and stroked rects
	// snap their edges and rounding to whole pixels, so shapes made of
	// several pieces meet without seams. Zero disables snapping.
	Scale float32

	clipStack []Rect
	texStack  []TextureID
	path      []Vec2
	baseClip  Rect
}

var infiniteClip = Rect{
	Min: Vec2{-math.MaxFloat32 / 2, -math.MaxFloat32 / 2},
	Max: Vec2{math.MaxFloat32 / 2, math.MaxFloat32 / 2},
}

// NewDrawList returns an empty list with an unbounded clip rect.
func NewDrawList() *DrawList {
	return &DrawList{AAWidth: 1, baseClip: infiniteClip}
}

// Reset clears all geometry and stacks for a new frame.
func (d *DrawList) Reset() {
	d.Vertices = d.Vertices[:0]
	d.Indices = d.Indices[:0]
	d.Cmds = d.Cmds[:0]
	d.clipStack = d.clipStack[:0]
	d.texStack = d.texStack[:0]
	d.path = d.path[:0]
}

// SetBaseClip sets the clip used when the clip stack is empty.
func (d *DrawList) SetBaseClip(r Rect) { d.baseClip = r }

// CurrentClip returns the active clip rect.
func (d *DrawList) CurrentClip() Rect {
	if n := len(d.clipStack); n > 0 {
		return d.clipStack[n-1]
	}
	return d.baseClip
}

// CurrentTexture returns the active texture.
func (d *DrawList) CurrentTexture() TextureID {
	if n := len(d.texStack); n > 0 {
		return d.texStack[n-1]
	}
	return 0
}

// PushClipRect pushes r intersected with the current clip.
func (d *DrawList) PushClipRect(r Rect) {
	if !finite(r.Min.X, r.Min.Y, r.Max.X, r.Max.Y) {
		r = d.CurrentClip()
	}
	d.clipStack = append(d.clipStack, r.Intersect(d.CurrentClip()))
}

// PopClipRect restores the previous clip rect.
func (d *DrawList) PopClipRect() {
	if len(d.clipStack) > 0 {
		d.clipStack = d.clipStack[:len(d.clipStack)-1]
	}
}

// PushTexture sets the active texture.
func (d *DrawList) PushTexture(id TextureID) { d.texStack = append(d.texStack, id) }

// PopTexture restores the previous texture.
func (d *DrawList) PopTexture() {
	if len(d.texStack) > 0 {
		d.texStack = d.texStack[:len(d.texStack)-1]
	}
}

// prim reserves space for a primitive and returns the first vertex index.
// It starts a new Cmd only when clip or texture differ from the last one.
// A Cmd also splits when the quad kind differs, see Cmd.Quads.
func (d *DrawList) prim(tex TextureID, nIdx int, quad bool) {
	clip := d.CurrentClip()
	if n := len(d.Cmds); n > 0 {
		last := &d.Cmds[n-1]
		if last.IdxCount == 0 {
			last.ClipRect, last.TextureID, last.IdxOffset, last.Quads = clip, tex, len(d.Indices), quad
		} else if last.ClipRect != clip || last.TextureID != tex || last.Quads != quad {
			d.Cmds = append(d.Cmds, Cmd{ClipRect: clip, TextureID: tex, IdxOffset: len(d.Indices), Quads: quad})
		}
	} else {
		d.Cmds = append(d.Cmds, Cmd{ClipRect: clip, TextureID: tex, IdxOffset: len(d.Indices), Quads: quad})
	}
	d.Cmds[len(d.Cmds)-1].IdxCount += nIdx
}

// outside reports whether the box lies fully outside the current clip.
func (d *DrawList) outside(x0, y0, x1, y1 float32) bool {
	c := d.CurrentClip()
	return c.Empty() || x1 <= c.Min.X || x0 >= c.Max.X || y1 <= c.Min.Y || y0 >= c.Max.Y
}

// quad appends one axis aligned quad as 4 vertices and 6 indices in the
// canonical layout described by Cmd.Quads.
func (d *DrawList) quad(min, max Vec2, uvMin, uvMax [2]float32, col uint32) {
	a := d.vtx(min, uvMin, col)
	d.vtx(Vec2{max.X, min.Y}, [2]float32{uvMax[0], uvMin[1]}, col)
	d.vtx(max, uvMax, col)
	d.vtx(Vec2{min.X, max.Y}, [2]float32{uvMin[0], uvMax[1]}, col)
	d.tri(a, a+1, a+2)
	d.tri(a, a+2, a+3)
}

func (d *DrawList) vtx(p Vec2, uv [2]float32, col uint32) uint32 {
	d.Vertices = append(d.Vertices, Vertex{Pos: p, UV: uv, Color: col})
	return uint32(len(d.Vertices) - 1)
}

func (d *DrawList) tri(a, b, c uint32) { d.Indices = append(d.Indices, a, b, c) }

// addMesh appends vertices and indices (relative) as one primitive.
func (d *DrawList) addMesh(tex TextureID, vs []Vertex, idx []uint32) {
	if len(idx) == 0 {
		return
	}
	for _, i := range idx {
		if int(i) >= len(vs) {
			return
		}
	}
	d.prim(tex, len(idx), false)
	base := uint32(len(d.Vertices))
	d.Vertices = append(d.Vertices, vs...)
	for _, i := range idx {
		d.Indices = append(d.Indices, base+i)
	}
}

// Path helpers

// PathClear drops the current path.
func (d *DrawList) PathClear() { d.path = d.path[:0] }

// PathLineTo appends a point.
func (d *DrawList) PathLineTo(p Vec2) {
	if finite(p.X, p.Y) {
		d.path = append(d.path, p)
	}
}

// PathArcTo appends an arc from angle a0 to a1 (radians) with n segments.
func (d *DrawList) PathArcTo(c Vec2, r, a0, a1 float32, n int) {
	if !finite(c.X, c.Y, r, a0, a1) {
		return
	}
	n = min(max(n, 1), maxArcSegments)
	for i := 0; i <= n; i++ {
		a := float64(a0) + float64(a1-a0)*float64(i)/float64(n)
		d.path = append(d.path, Vec2{c.X + float32(math.Cos(a))*r, c.Y + float32(math.Sin(a))*r})
	}
}

// PathRect appends a rect outline, rounding the corners selected by flags.
func (d *DrawList) PathRect(min, max Vec2, rounding float32, flags Corner) {
	if !finite(min.X, min.Y, max.X, max.Y, rounding) {
		return
	}
	w, h := max.X-min.X, max.Y-min.Y
	rounding = clampf(rounding, 0, min32(w, h)/2)
	if rounding <= 0 || flags == 0 {
		d.path = append(d.path, min, Vec2{max.X, min.Y}, max, Vec2{min.X, max.Y})
		return
	}
	segs := arcSegments(rounding, 1/d.aa())
	corner := func(f Corner, c Vec2, a0 float32, sx, sy float32) {
		if flags&f != 0 {
			d.PathArcTo(c, rounding, a0*math.Pi/2, (a0+1)*math.Pi/2, segs)
		} else {
			d.path = append(d.path, Vec2{c.X + sx*rounding, c.Y + sy*rounding})
		}
	}
	corner(CornerTopLeft, Vec2{min.X + rounding, min.Y + rounding}, 2, -1, -1)
	corner(CornerTopRight, Vec2{max.X - rounding, min.Y + rounding}, 3, 1, -1)
	corner(CornerBottomRight, Vec2{max.X - rounding, max.Y - rounding}, 0, 1, 1)
	corner(CornerBottomLeft, Vec2{min.X + rounding, max.Y - rounding}, 1, -1, 1)
}

// PathFillConvex fills the current path as a triangle fan and clears it.
func (d *DrawList) PathFillConvex(col uint32) {
	p := d.path
	if len(p) >= 3 && col>>24 != 0 && !d.pathOutside(p, 0) {
		d.prim(d.CurrentTexture(), (len(p)-2)*3, false)
		first := d.vtx(p[0], [2]float32{}, col)
		for _, q := range p[1:] {
			d.vtx(q, [2]float32{}, col)
		}
		flip := polyArea(p) < 0
		for i := 1; i < len(p)-1; i++ {
			if flip {
				d.tri(first, first+uint32(i)+1, first+uint32(i))
			} else {
				d.tri(first, first+uint32(i), first+uint32(i)+1)
			}
		}
	}
	d.path = d.path[:0]
}

// PathStroke strokes the current path with miter joins and clears it.
func (d *DrawList) PathStroke(col uint32, closed bool, thickness float32) {
	p := dedupe(d.path, closed)
	n := len(p)
	if n < 3 {
		closed = false
	}
	if n >= 2 && col>>24 != 0 && thickness > 0 && finite(thickness) && !d.pathOutside(p, thickness) {
		half := thickness / 2
		segs := n - 1
		if closed {
			segs = n
		}
		// per-segment unit normals
		nor := make([]Vec2, segs)
		for i := 0; i < segs; i++ {
			a, b := p[i], p[(i+1)%n]
			dx, dy := b.X-a.X, b.Y-a.Y
			l := float32(math.Hypot(float64(dx), float64(dy)))
			if l > 0 {
				nor[i] = Vec2{-dy / l, dx / l}
			}
		}
		d.prim(d.CurrentTexture(), segs*6, false)
		base := uint32(len(d.Vertices))
		for i := 0; i < n; i++ {
			var m Vec2
			switch {
			case closed:
				m = miter(nor[(i+n-1)%n], nor[i])
			case i == 0:
				m = nor[0]
			case i == n-1:
				m = nor[segs-1]
			default:
				m = miter(nor[i-1], nor[i])
			}
			d.vtx(Vec2{p[i].X + m.X*half, p[i].Y + m.Y*half}, [2]float32{}, col)
			d.vtx(Vec2{p[i].X - m.X*half, p[i].Y - m.Y*half}, [2]float32{}, col)
		}
		for i := 0; i < segs; i++ {
			a := base + uint32(i)*2
			b := base + uint32((i+1)%n)*2
			d.tri(a, a+1, b)
			d.tri(b, a+1, b+1)
		}
	}
	d.path = d.path[:0]
}

// pathOutside reports whether the bounds of p grown by pad are outside the clip.
func (d *DrawList) pathOutside(p []Vec2, pad float32) bool {
	x0, y0, x1, y1 := p[0].X, p[0].Y, p[0].X, p[0].Y
	for _, q := range p[1:] {
		x0, y0, x1, y1 = min32(x0, q.X), min32(y0, q.Y), max32(x1, q.X), max32(y1, q.Y)
	}
	return d.outside(x0-pad, y0-pad, x1+pad, y1+pad)
}

// miterLimit caps the join offset length (in half-thickness units).
const miterLimit = 2

// maxArcSegments caps arc subdivision.
const maxArcSegments = 1024

// miter returns the join offset vector for two unit normals. Sharp turns
// are beveled by limiting the offset length instead of pinching.
func miter(a, b Vec2) Vec2 {
	m := Vec2{a.X + b.X, a.Y + b.Y}
	l2 := m.X*m.X + m.Y*m.Y
	if l2 < 1e-6 {
		return a
	}
	l := float32(math.Sqrt(float64(l2)))
	k := min32(2/l2, miterLimit/l)
	return Vec2{m.X * k, m.Y * k}
}

// dedupe drops consecutive duplicate points and, when closed, a final
// point equal to the first.
func dedupe(p []Vec2, closed bool) []Vec2 {
	out := make([]Vec2, 0, len(p))
	for _, q := range p {
		if len(out) == 0 || out[len(out)-1] != q {
			out = append(out, q)
		}
	}
	if closed {
		for len(out) > 1 && out[len(out)-1] == out[0] {
			out = out[:len(out)-1]
		}
	}
	return out
}

// polyArea returns the signed area (positive is clockwise on y-down).
func polyArea(p []Vec2) float32 {
	var a float32
	for i := range p {
		q := p[(i+1)%len(p)]
		a += p[i].X*q.Y - q.X*p[i].Y
	}
	return a / 2
}

func finite(fs ...float32) bool {
	for _, f := range fs {
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return false
		}
	}
	return true
}

// Primitives

// AddRectFilled fills a rect with optional rounding on all corners.
func (d *DrawList) AddRectFilled(min, max Vec2, c colors.Color, rounding float32) {
	d.AddRectFilledEx(min, max, c, rounding, CornerAll)
}

// AddRectFilledEx fills a rect rounding only the corners in flags.
func (d *DrawList) AddRectFilledEx(min, max Vec2, c colors.Color, rounding float32, flags Corner) {
	min, max, rounding = d.snapRect(min, max, rounding)
	col := Pack(c)
	if col>>24 == 0 || !finite(min.X, min.Y, max.X, max.Y, rounding) || max.X <= min.X || max.Y <= min.Y {
		return
	}
	if d.outside(min.X, min.Y, max.X, max.Y) {
		return
	}
	if rounding <= 0 || flags == 0 {
		d.fillRect(min, max, col)
		return
	}
	if (max.X-min.X)*(max.Y-min.Y) >= roundedSplitArea {
		d.fillRounded(min, max, col, rounding, flags)
		return
	}
	d.PathClear()
	d.PathRect(min, max, rounding, flags)
	d.PathFillConvex(col)
}

// roundedSplitArea is the rect area from which rounded fills are split into
// solid quads plus corner pies so backends can fill the body quickly.
const roundedSplitArea = 1024

// fillRounded fills a rounded rect as solid quads for the body and one pie
// per rounded corner. Quads come first so they share one Cmd.
func (d *DrawList) fillRounded(min, max Vec2, col uint32, rounding float32, flags Corner) {
	r := clampf(rounding, 0, min32(max.X-min.X, max.Y-min.Y)/2)
	inset := func(f Corner) float32 {
		if flags&f != 0 {
			return r
		}
		return 0
	}
	tl, tr, br, bl := inset(CornerTopLeft), inset(CornerTopRight), inset(CornerBottomRight), inset(CornerBottomLeft)
	if x0, x1 := min.X+tl, max.X-tr; x1 > x0 {
		d.fillRect(Vec2{x0, min.Y}, Vec2{x1, min.Y + r}, col)
	}
	if max.Y-r > min.Y+r {
		d.fillRect(Vec2{min.X, min.Y + r}, Vec2{max.X, max.Y - r}, col)
	}
	if x0, x1 := min.X+bl, max.X-br; x1 > x0 {
		d.fillRect(Vec2{x0, max.Y - r}, Vec2{x1, max.Y}, col)
	}
	segs := arcSegments(r, 1/d.aa())
	pie := func(f Corner, c Vec2, a0 float32) {
		if flags&f == 0 {
			return
		}
		d.PathClear()
		d.path = append(d.path, c)
		d.PathArcTo(c, r, a0*math.Pi/2, (a0+1)*math.Pi/2, segs)
		d.fillFan(col)
	}
	pie(CornerTopLeft, Vec2{min.X + r, min.Y + r}, 2)
	pie(CornerTopRight, Vec2{max.X - r, min.Y + r}, 3)
	pie(CornerBottomRight, Vec2{max.X - r, max.Y - r}, 0)
	pie(CornerBottomLeft, Vec2{min.X + r, max.Y - r}, 1)
	d.path = d.path[:0]
}

// fillFan fills path[0] and the following points as a triangle fan.
func (d *DrawList) fillFan(col uint32) {
	p := d.path
	d.prim(d.CurrentTexture(), (len(p)-2)*3, false)
	first := d.vtx(p[0], [2]float32{}, col)
	for _, q := range p[1:] {
		d.vtx(q, [2]float32{}, col)
	}
	for i := 1; i < len(p)-1; i++ {
		d.tri(first, first+uint32(i), first+uint32(i)+1)
	}
}

// fillRect emits one solid quad.
func (d *DrawList) fillRect(min, max Vec2, col uint32) {
	d.prim(d.CurrentTexture(), 6, true)
	d.quad(min, max, [2]float32{}, [2]float32{}, col)
}

// fillRectClipped emits a solid quad unless it is outside the clip.
func (d *DrawList) fillRectClipped(min, max Vec2, col uint32) {
	if !d.outside(min.X, min.Y, max.X, max.Y) {
		d.fillRect(min, max, col)
	}
}

// AddRect strokes a rect outline centered on the edge.
func (d *DrawList) AddRect(min, max Vec2, c colors.Color, rounding, thickness float32) {
	if !finite(min.X, min.Y, max.X, max.Y, rounding, thickness) {
		return
	}
	min, max, rounding = d.snapRect(min, max, rounding)
	col := Pack(c)
	if col>>24 == 0 || thickness <= 0 {
		return
	}
	h := thickness / 2
	if d.outside(min.X-h, min.Y-h, max.X+h, max.Y+h) {
		return
	}
	w, ht := max.X-min.X, max.Y-min.Y
	if rounding <= 0 && w > 0 && ht > 0 && thickness <= min32(w, ht) {
		// Square corners are four non-overlapping fills.
		d.fillRect(Vec2{min.X - h, min.Y - h}, Vec2{max.X + h, min.Y + h}, col)
		d.fillRect(Vec2{min.X - h, max.Y - h}, Vec2{max.X + h, max.Y + h}, col)
		if ht > thickness {
			d.fillRect(Vec2{min.X - h, min.Y + h}, Vec2{min.X + h, max.Y - h}, col)
			d.fillRect(Vec2{max.X - h, min.Y + h}, Vec2{max.X + h, max.Y - h}, col)
		}
		return
	}
	d.PathClear()
	d.PathRect(min, max, rounding, CornerAll)
	d.PathStroke(col, true, thickness)
}

// AddLine draws a line segment.
func (d *DrawList) AddLine(a, b Vec2, c colors.Color, thickness float32) {
	if !finite(a.X, a.Y, b.X, b.Y, thickness) {
		return
	}
	col := Pack(c)
	if col>>24 == 0 || thickness <= 0 || a == b {
		return
	}
	h := thickness / 2
	switch {
	case a.Y == b.Y:
		d.fillRectClipped(Vec2{min32(a.X, b.X), a.Y - h}, Vec2{max32(a.X, b.X), a.Y + h}, col)
		return
	case a.X == b.X:
		d.fillRectClipped(Vec2{a.X - h, min32(a.Y, b.Y)}, Vec2{a.X + h, max32(a.Y, b.Y)}, col)
		return
	}
	d.PathClear()
	d.PathLineTo(a)
	d.PathLineTo(b)
	d.PathStroke(col, false, thickness)
}

// AddTriangleFilled fills a triangle.
func (d *DrawList) AddTriangleFilled(a, b, c Vec2, col colors.Color) {
	p := Pack(col)
	if p>>24 == 0 || !finite(a.X, a.Y, b.X, b.Y, c.X, c.Y) {
		return
	}
	cross := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	if cross == 0 {
		return
	}
	if d.outside(min32(a.X, min32(b.X, c.X)), min32(a.Y, min32(b.Y, c.Y)),
		max32(a.X, max32(b.X, c.X)), max32(a.Y, max32(b.Y, c.Y))) {
		return
	}
	if cross < 0 {
		b, c = c, b
	}
	d.prim(d.CurrentTexture(), 3, false)
	i := d.vtx(a, [2]float32{}, p)
	d.vtx(b, [2]float32{}, p)
	d.vtx(c, [2]float32{}, p)
	d.tri(i, i+1, i+2)
}

// AddImage draws a textured quad tinted by tint.
func (d *DrawList) AddImage(tex TextureID, min, max Vec2, uvMin, uvMax [2]float32, tint colors.Color) {
	col := Pack(tint)
	if col>>24 == 0 || !finite(min.X, min.Y, max.X, max.Y) {
		return
	}
	if d.outside(min32(min.X, max.X), min32(min.Y, max.Y), max32(min.X, max.X), max32(min.Y, max.Y)) {
		return
	}
	d.prim(tex, 6, true)
	d.quad(min, max, uvMin, uvMax, col)
}

// AddText emits pre-laid-out glyph quads using the atlas texture tex.
func (d *DrawList) AddText(tex TextureID, quads []GlyphQuad, c colors.Color) {
	col := Pack(c)
	if col>>24 == 0 || len(quads) == 0 {
		return
	}
	for _, q := range quads {
		if !finite(q.Min.X, q.Min.Y, q.Max.X, q.Max.Y) {
			return
		}
	}
	vis := 0
	for _, q := range quads {
		if !d.outside(q.Min.X, q.Min.Y, q.Max.X, q.Max.Y) {
			vis++
		}
	}
	if vis == 0 {
		return
	}
	d.prim(tex, vis*6, true)
	for _, q := range quads {
		if !d.outside(q.Min.X, q.Min.Y, q.Max.X, q.Max.Y) {
			d.quad(q.Min, q.Max, q.UVMin, q.UVMax, col)
		}
	}
}

// AddCircleFilled draws an antialiased filled circle.
func (d *DrawList) AddCircleFilled(center Vec2, radius float32, c colors.Color) {
	col := Pack(c)
	if !finite(center.X, center.Y, radius) || radius <= 0 || col>>24 == 0 {
		return
	}
	aa := d.aa()
	if r := radius + aa; d.outside(center.X-r, center.Y-r, center.X+r, center.Y+r) {
		return
	}
	inner := max32(0, radius-aa*0.5)
	outer := radius + aa*0.5
	segs := circleSegments(radius, 1/aa)
	transparent := col &^ 0xFF000000

	d.prim(d.CurrentTexture(), segs*9, false)
	base := d.vtx(center, [2]float32{}, col)
	d.appendRing(center, inner, segs, col)
	d.appendRing(center, outer, segs, transparent)
	s := uint32(segs)
	for i := uint32(0); i < s; i++ {
		d.tri(base, base+1+i, base+1+(i+1)%s)
	}
	d.appendRingTriangles(base+1, base+1+s, segs)
}

// AddCircle draws an antialiased circle outline.
func (d *DrawList) AddCircle(center Vec2, radius, thickness float32, c colors.Color) {
	col := Pack(c)
	if !finite(center.X, center.Y, radius, thickness) || radius <= 0 || thickness <= 0 || col>>24 == 0 {
		return
	}
	if thickness >= radius {
		d.AddCircleFilled(center, radius, c)
		return
	}
	aa := d.aa()
	if r := radius + aa; d.outside(center.X-r, center.Y-r, center.X+r, center.Y+r) {
		return
	}
	in := radius - thickness
	radii := [4]float32{max32(0, in-aa*0.5), in + aa*0.5, radius - aa*0.5, radius + aa*0.5}
	alpha := float32(1)
	if radii[1] > radii[2] {
		radii[1] = (in + radius) * 0.5
		radii[2] = radii[1]
		alpha = min32(1, thickness/aa)
	}
	opaque := col
	if alpha < 1 {
		opaque = col&^0xFF000000 | uint32(float32(col>>24)*alpha)<<24
	}
	transparent := col &^ 0xFF000000
	cols := [4]uint32{transparent, opaque, opaque, transparent}
	segs := circleSegments(radius, 1/aa)

	d.prim(d.CurrentTexture(), segs*18, false)
	base := uint32(len(d.Vertices))
	for i, r := range radii {
		d.appendRing(center, r, segs, cols[i])
	}
	s := uint32(segs)
	for i := uint32(0); i < 3; i++ {
		d.appendRingTriangles(base+i*s, base+(i+1)*s, segs)
	}
}

// snapRect rounds the edges and the rounding radius to whole pixels.
func (d *DrawList) snapRect(min, max Vec2, rounding float32) (Vec2, Vec2, float32) {
	sc := d.Scale
	if sc <= 0 || !finite(min.X, min.Y, max.X, max.Y, rounding) {
		return min, max, rounding
	}
	snap := func(v float32) float32 { return float32(math.Ceil(float64(v*sc)-0.5)) / sc }
	min = Vec2{snap(min.X), snap(min.Y)}
	max = Vec2{snap(max.X), snap(max.Y)}
	return min, max, snap(rounding)
}

func (d *DrawList) aa() float32 {
	if d.AAWidth <= 0 {
		return 1
	}
	return d.AAWidth
}

// appendRing adds segs vertices on a circle (port of appendCircleRing).
func (d *DrawList) appendRing(c Vec2, radius float32, segs int, col uint32) {
	for i := 0; i < segs; i++ {
		a := float64(i) * 2 * math.Pi / float64(segs)
		d.vtx(Vec2{c.X + float32(math.Cos(a))*radius, c.Y + float32(math.Sin(a))*radius}, [2]float32{}, col)
	}
}

// appendRingTriangles stitches two rings (port of appendRingTriangles).
func (d *DrawList) appendRingTriangles(innerStart, outerStart uint32, segs int) {
	s := uint32(segs)
	for i := uint32(0); i < s; i++ {
		n := (i + 1) % s
		d.tri(innerStart+i, outerStart+i, outerStart+n)
		d.tri(innerStart+i, outerStart+n, innerStart+n)
	}
}

// circleSegments uses the physical radius (radius*scale).
func circleSegments(radius, scale float32) int {
	phys := float64(max32(1, radius*scale))
	return min(256, max(24, int(math.Ceil(2*math.Pi*math.Sqrt(phys)))))
}

func arcSegments(r, scale float32) int {
	return max(2, circleSegments(r, scale)/4)
}

func clampf(v, lo, hi float32) float32 { return min32(max32(v, lo), hi) }
func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
