package draw

import (
	"math"
	"testing"

	"github.com/tacf/goak/colors"
)

var red = colors.RGB(255, 0, 0)

func checkIndices(t *testing.T, d *DrawList) {
	t.Helper()
	total := 0
	for _, c := range d.Cmds {
		total += c.IdxCount
	}
	if total != len(d.Indices) {
		t.Fatalf("cmd idx sum %d != %d", total, len(d.Indices))
	}
	if len(d.Indices)%3 != 0 {
		t.Fatalf("indices not multiple of 3")
	}
	for _, i := range d.Indices {
		if int(i) >= len(d.Vertices) {
			t.Fatalf("index %d >= vertex count %d", i, len(d.Vertices))
		}
	}
}

func TestPrimitiveCounts(t *testing.T) {
	v2 := func(x, y float32) Vec2 { return Vec2{x, y} }
	tests := []struct {
		name     string
		fn       func(d *DrawList)
		vtx, idx int
	}{
		{"rect", func(d *DrawList) { d.AddRectFilled(v2(0, 0), v2(10, 10), red, 0) }, 4, 6},
		{"line", func(d *DrawList) { d.AddLine(v2(0, 0), v2(10, 0), red, 2) }, 4, 6},
		{"rectStroke", func(d *DrawList) { d.AddRect(v2(0, 0), v2(10, 10), red, 0, 1) }, 16, 24},
		{"tri", func(d *DrawList) { d.AddTriangleFilled(v2(0, 0), v2(5, 0), v2(0, 5), red) }, 3, 3},
		{"image", func(d *DrawList) { d.AddImage(7, v2(0, 0), v2(4, 4), [2]float32{0, 0}, [2]float32{1, 1}, red) }, 4, 6},
		{"text", func(d *DrawList) { d.AddText(3, make([]GlyphQuad, 3), red) }, 12, 18},
	}
	for _, tc := range tests {
		d := NewDrawList()
		tc.fn(d)
		if len(d.Vertices) != tc.vtx || len(d.Indices) != tc.idx {
			t.Errorf("%s: got %d/%d want %d/%d", tc.name, len(d.Vertices), len(d.Indices), tc.vtx, tc.idx)
		}
		checkIndices(t, d)
	}
}

func TestCircleCounts(t *testing.T) {
	d := NewDrawList()
	d.AddCircleFilled(Vec2{50, 50}, 10, red)
	s := circleSegments(10, 1)
	if len(d.Vertices) != 1+2*s || len(d.Indices) != 9*s {
		t.Fatalf("filled: %d/%d", len(d.Vertices), len(d.Indices))
	}
	checkIndices(t, d)

	d = NewDrawList()
	d.AddCircle(Vec2{50, 50}, 10, 2, red)
	if len(d.Vertices) != 4*s || len(d.Indices) != 18*s {
		t.Fatalf("ring: %d/%d", len(d.Vertices), len(d.Indices))
	}
	checkIndices(t, d)

	// thickness >= radius falls back to filled
	d = NewDrawList()
	d.AddCircle(Vec2{0, 0}, 4, 10, red)
	if len(d.Vertices) != 1+2*circleSegments(4, 1) {
		t.Fatalf("fallback: %d", len(d.Vertices))
	}
}

func TestCircleEdgeAlpha(t *testing.T) {
	d := NewDrawList()
	d.AddCircleFilled(Vec2{}, 10, red)
	if d.Vertices[0].Color>>24 != 255 {
		t.Fatal("center must be opaque")
	}
	last := d.Vertices[len(d.Vertices)-1]
	if last.Color>>24 != 0 {
		t.Fatal("outer ring must be transparent")
	}
}

func TestRoundedRect(t *testing.T) {
	d := NewDrawList()
	d.AddRectFilled(Vec2{0, 0}, Vec2{40, 40}, red, 8)
	if len(d.Vertices) <= 4 {
		t.Fatal("rounded rect should have extra vertices")
	}
	checkIndices(t, d)
	d = NewDrawList()
	d.AddRect(Vec2{0, 0}, Vec2{40, 40}, red, 8, 2)
	checkIndices(t, d)
	d = NewDrawList()
	d.AddRectFilledEx(Vec2{0, 0}, Vec2{40, 40}, red, 8, CornerTopLeft)
	checkIndices(t, d)
	// rounding clamps to half size
	d = NewDrawList()
	d.AddRectFilled(Vec2{0, 0}, Vec2{10, 10}, red, 100)
	checkIndices(t, d)
}

func TestSkipInvisible(t *testing.T) {
	d := NewDrawList()
	d.AddRectFilled(Vec2{0, 0}, Vec2{10, 10}, colors.RGBA(1, 2, 3, 0), 0)
	d.AddRectFilled(Vec2{5, 5}, Vec2{5, 9}, red, 0)
	if len(d.Cmds) != 0 || len(d.Vertices) != 0 {
		t.Fatal("expected nothing drawn")
	}
}

func TestMergeSameState(t *testing.T) {
	d := NewDrawList()
	for i := 0; i < 5; i++ {
		d.AddRectFilled(Vec2{0, 0}, Vec2{1, 1}, red, 0)
		d.AddLine(Vec2{0, 0}, Vec2{5, 0}, red, 1)
	}
	if len(d.Cmds) != 1 || d.Cmds[0].IdxCount != len(d.Indices) {
		t.Fatalf("cmds=%d", len(d.Cmds))
	}
	checkIndices(t, d)
}

func TestClipIntersectAndRestore(t *testing.T) {
	d := NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{100, 100}})
	d.PushClipRect(Rect{Vec2{50, 50}, Vec2{200, 200}})
	if got := d.CurrentClip(); got != (Rect{Vec2{50, 50}, Vec2{100, 100}}) {
		t.Fatalf("intersect: %+v", got)
	}
	d.PopClipRect()
	if got := d.CurrentClip(); got != (Rect{Vec2{0, 0}, Vec2{100, 100}}) {
		t.Fatalf("restore: %+v", got)
	}
	d.PopClipRect()
	if d.CurrentClip() != infiniteClip {
		t.Fatal("base clip not restored")
	}
	d.PopClipRect() // extra pop is safe
}

func TestDisjointClipEmpty(t *testing.T) {
	d := NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{10, 10}})
	d.PushClipRect(Rect{Vec2{20, 20}, Vec2{30, 30}})
	if !d.CurrentClip().Empty() {
		t.Fatal("disjoint clips must give empty clip")
	}
}

func TestSplitOnClipChange(t *testing.T) {
	d := NewDrawList()
	r := func() { d.AddRectFilled(Vec2{0, 0}, Vec2{1, 1}, red, 0) }
	r()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{5, 5}})
	r()
	r()
	d.PopClipRect()
	r()
	if len(d.Cmds) != 3 {
		t.Fatalf("cmds=%d", len(d.Cmds))
	}
	if d.Cmds[1].IdxCount != 12 || d.Cmds[1].IdxOffset != 6 || d.Cmds[2].IdxOffset != 18 {
		t.Fatalf("bad cmds %+v", d.Cmds)
	}
	checkIndices(t, d)
}

func TestSplitOnTextureChange(t *testing.T) {
	d := NewDrawList()
	d.AddRectFilled(Vec2{}, Vec2{1, 1}, red, 0)
	d.PushTexture(9)
	d.AddRectFilled(Vec2{}, Vec2{1, 1}, red, 0)
	d.PopTexture()
	d.AddRectFilled(Vec2{}, Vec2{1, 1}, red, 0)
	if len(d.Cmds) != 3 || d.Cmds[1].TextureID != 9 || d.Cmds[2].TextureID != 0 {
		t.Fatalf("cmds %+v", d.Cmds)
	}
	d.AddImage(9, Vec2{}, Vec2{1, 1}, [2]float32{}, [2]float32{1, 1}, red)
	if len(d.Cmds) != 4 {
		t.Fatal("image with other texture must split")
	}
	checkIndices(t, d)
}

func TestNoEmptyCmds(t *testing.T) {
	d := NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{5, 5}})
	d.PopClipRect()
	d.AddRectFilled(Vec2{}, Vec2{1, 1}, red, 0)
	for _, c := range d.Cmds {
		if c.IdxCount == 0 {
			t.Fatal("empty cmd")
		}
	}
}

func TestResetAndPack(t *testing.T) {
	d := NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{5, 5}})
	d.AddRectFilled(Vec2{}, Vec2{1, 1}, red, 0)
	d.Reset()
	if len(d.Vertices)+len(d.Indices)+len(d.Cmds) != 0 || d.CurrentClip() != infiniteClip {
		t.Fatal("reset incomplete")
	}
	c := colors.RGBA(1, 2, 3, 4)
	if Unpack(Pack(c)) != c {
		t.Fatal("pack roundtrip")
	}
}

func TestStrokeClosedPolylineValid(t *testing.T) {
	d := NewDrawList()
	d.PathLineTo(Vec2{0, 0})
	d.PathLineTo(Vec2{0, 0}) // degenerate segment
	d.PathLineTo(Vec2{10, 0})
	d.PathStroke(Pack(red), true, 2)
	checkIndices(t, d)
}

var _ Renderer = (*fakeBackend)(nil)

type fakeBackend struct{ got *DrawData }

func (f *fakeBackend) NewTexture(w, h int, rgba []byte) (TextureID, error) { return 1, nil }
func (f *fakeBackend) UpdateTexture(TextureID, int, int, []byte) error     { return nil }
func (f *fakeBackend) DestroyTexture(TextureID)                            {}
func (f *fakeBackend) Render(dd *DrawData) error                           { f.got = dd; return nil }

func TestBackendInterface(t *testing.T) {
	fb := &fakeBackend{}
	dd := &DrawData{Lists: []*DrawList{NewDrawList()}, DisplaySize: Vec2{800, 600}, Scale: Vec2{1, 1}}
	if err := fb.Render(dd); err != nil || fb.got != dd {
		t.Fatal("render")
	}
}

func cross(d *DrawList, i int) float32 {
	a, b, c := d.Vertices[d.Indices[i]].Pos, d.Vertices[d.Indices[i+1]].Pos, d.Vertices[d.Indices[i+2]].Pos
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func allClockwise(t *testing.T, name string, d *DrawList) {
	t.Helper()
	if len(d.Indices) == 0 {
		t.Fatalf("%s: empty", name)
	}
	for i := 0; i < len(d.Indices); i += 3 {
		if cross(d, i) < -1e-4 {
			t.Fatalf("%s: tri %d counter-clockwise", name, i/3)
		}
	}
}

func TestWinding(t *testing.T) {
	mk := func(f func(d *DrawList)) *DrawList { d := NewDrawList(); f(d); return d }
	allClockwise(t, "rect", mk(func(d *DrawList) { d.AddRectFilled(Vec2{}, Vec2{10, 10}, red, 0) }))
	allClockwise(t, "rounded", mk(func(d *DrawList) { d.AddRectFilled(Vec2{}, Vec2{20, 20}, red, 5) }))
	allClockwise(t, "stroke", mk(func(d *DrawList) { d.AddRect(Vec2{}, Vec2{20, 20}, red, 4, 2) }))
	allClockwise(t, "line", mk(func(d *DrawList) { d.AddLine(Vec2{}, Vec2{5, 7}, red, 2) }))
	allClockwise(t, "line2", mk(func(d *DrawList) { d.AddLine(Vec2{5, 7}, Vec2{}, red, 2) }))
	allClockwise(t, "circle", mk(func(d *DrawList) { d.AddCircleFilled(Vec2{}, 9, red) }))
	allClockwise(t, "ring", mk(func(d *DrawList) { d.AddCircle(Vec2{}, 9, 2, red) }))
	allClockwise(t, "triCW", mk(func(d *DrawList) { d.AddTriangleFilled(Vec2{0, 0}, Vec2{5, 0}, Vec2{0, 5}, red) }))
	allClockwise(t, "triCCW", mk(func(d *DrawList) { d.AddTriangleFilled(Vec2{0, 0}, Vec2{0, 5}, Vec2{5, 0}, red) }))
	// reversed path still fills clockwise
	d := NewDrawList()
	d.PathLineTo(Vec2{0, 0})
	d.PathLineTo(Vec2{0, 10})
	d.PathLineTo(Vec2{10, 10})
	d.PathLineTo(Vec2{10, 0})
	d.PathFillConvex(Pack(red))
	allClockwise(t, "revpath", d)
}

func TestCircleSegmentsScale(t *testing.T) {
	if circleSegments(20, 2) <= circleSegments(20, 1) {
		t.Fatal("higher scale needs more segments")
	}
	d1, d2 := NewDrawList(), NewDrawList()
	d2.AAWidth = 0.5
	d1.AddCircleFilled(Vec2{}, 50, red)
	d2.AddCircleFilled(Vec2{}, 50, red)
	if len(d2.Vertices) <= len(d1.Vertices) {
		t.Fatal("AAWidth 0.5 should add segments")
	}
}

func TestStrokeDedupe(t *testing.T) {
	d := NewDrawList()
	for _, p := range []Vec2{{0, 0}, {0, 0}, {10, 0}, {10, 0}, {10, 10}, {0, 0}} {
		d.PathLineTo(p)
	}
	d.PathStroke(Pack(red), true, 2)
	// triangle path: 3 unique points, 3 segments
	if len(d.Vertices) != 6 || len(d.Indices) != 18 {
		t.Fatalf("got %d/%d", len(d.Vertices), len(d.Indices))
	}
	checkIndices(t, d)
	// zero length line draws nothing
	d = NewDrawList()
	d.AddLine(Vec2{1, 1}, Vec2{1, 1}, red, 2)
	if len(d.Indices) != 0 {
		t.Fatal("zero-length line drew")
	}
}

func TestMiterBevel(t *testing.T) {
	// near 180 degree turn
	a := Vec2{0, 1}
	b := Vec2{0.001, -1}
	m := miter(a, Vec2{b.X, b.Y})
	if got := m.X*m.X + m.Y*m.Y; got > miterLimit*miterLimit+1e-3 {
		t.Fatalf("offset too long %v", got)
	}
	// right angle keeps full miter
	m = miter(Vec2{0, 1}, Vec2{1, 0})
	if got := m.X*m.X + m.Y*m.Y; got < 1.9 || got > 2.1 {
		t.Fatalf("right angle miter len2=%v", got)
	}
	// sharp spike stays bounded in the draw list
	d := NewDrawList()
	d.PathLineTo(Vec2{0, 0})
	d.PathLineTo(Vec2{100, 0})
	d.PathLineTo(Vec2{0, 1})
	d.PathStroke(Pack(red), false, 2)
	for _, v := range d.Vertices {
		if v.Pos.X > 100+2*miterLimit+1e-3 {
			t.Fatalf("spike vertex too far %v", v.Pos)
		}
	}
}

func TestNonFiniteRejected(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	d := NewDrawList()
	d.AddCircleFilled(Vec2{}, nan, red)
	d.AddCircle(Vec2{}, 5, inf, red)
	d.AddCircle(Vec2{nan, 0}, 5, 1, red)
	d.AddRectFilled(Vec2{}, Vec2{inf, 5}, red, 0)
	d.AddRectFilled(Vec2{}, Vec2{5, 5}, red, nan)
	d.AddRect(Vec2{nan, 0}, Vec2{5, 5}, red, 0, 1)
	d.AddLine(Vec2{}, Vec2{nan, 1}, red, 1)
	d.AddLine(Vec2{}, Vec2{1, 1}, red, inf)
	d.AddTriangleFilled(Vec2{}, Vec2{nan, 1}, Vec2{1, 0}, red)
	d.AddImage(1, Vec2{}, Vec2{inf, 1}, [2]float32{}, [2]float32{1, 1}, red)
	d.AddText(1, []GlyphQuad{{Max: Vec2{nan, 1}}}, red)
	if len(d.Vertices) != 0 || len(d.Cmds) != 0 {
		t.Fatalf("drew %d vertices", len(d.Vertices))
	}
	d.PushClipRect(Rect{Vec2{nan, 0}, Vec2{1, 1}})
	if d.CurrentClip() != infiniteClip {
		t.Fatal("bad clip should keep current")
	}
	d.PopClipRect()
	d.PathLineTo(Vec2{nan, 0})
	d.PathArcTo(Vec2{}, inf, 0, 1, 4)
	if len(d.path) != 0 {
		t.Fatal("path accepted non-finite")
	}
}

func TestArcClampAndMeshValidation(t *testing.T) {
	d := NewDrawList()
	d.PathArcTo(Vec2{}, 5, 0, 1, 1<<30)
	if len(d.path) != maxArcSegments+1 {
		t.Fatalf("arc points %d", len(d.path))
	}
	d.addMesh(0, make([]Vertex, 3), []uint32{0, 1, 3})
	if len(d.Vertices) != 0 || len(d.Cmds) != 0 || len(d.Indices) != 0 {
		t.Fatal("invalid mesh accepted")
	}
	d.addMesh(0, make([]Vertex, 3), []uint32{0, 1, 2})
	if len(d.Vertices) != 3 || len(d.Indices) != 3 {
		t.Fatal("valid mesh rejected")
	}
	checkIndices(t, d)
}

func TestQuadFlag(t *testing.T) {
	d := NewDrawList()
	d.AddRectFilled(Vec2{0, 0}, Vec2{10, 10}, red, 0)
	d.AddRect(Vec2{0, 0}, Vec2{10, 10}, red, 0, 1)
	d.AddLine(Vec2{0, 0}, Vec2{10, 0}, red, 1)
	d.AddImage(5, Vec2{0, 0}, Vec2{4, 4}, [2]float32{}, [2]float32{1, 1}, red)
	d.AddText(5, make([]GlyphQuad, 2), red)
	for i, c := range d.Cmds {
		if !c.Quads {
			t.Errorf("cmd %d should be quads", i)
		}
	}
	n := len(d.Cmds)
	d.AddRectFilled(Vec2{0, 0}, Vec2{10, 10}, red, 3)
	if len(d.Cmds) != n+1 || d.Cmds[n].Quads {
		t.Fatal("rounded rect must start a general cmd")
	}
	d.AddCircleFilled(Vec2{5, 5}, 4, red)
	d.AddRectFilled(Vec2{0, 0}, Vec2{10, 10}, red, 0)
	if len(d.Cmds) != n+2 || !d.Cmds[n+1].Quads || d.Cmds[n+1].IdxCount != 6 {
		t.Fatalf("quad after general must split: %+v", d.Cmds[n:])
	}
}

func TestQuadLayout(t *testing.T) {
	d := NewDrawList()
	d.AddImage(5, Vec2{1, 2}, Vec2{5, 8}, [2]float32{.1, .2}, [2]float32{.3, .4}, red)
	v := d.Vertices
	if v[0].Pos != (Vec2{1, 2}) || v[2].Pos != (Vec2{5, 8}) || v[2].UV != [2]float32{.3, .4} {
		t.Fatalf("bad corners %+v", v)
	}
	want := []uint32{0, 1, 2, 0, 2, 3}
	for i, w := range want {
		if d.Indices[i] != w {
			t.Fatalf("indices %v", d.Indices)
		}
	}
}

func TestStrokeSquareIsFourFills(t *testing.T) {
	d := NewDrawList()
	d.AddRect(Vec2{10, 10}, Vec2{30, 20}, red, 0, 2)
	if len(d.Cmds) != 1 || !d.Cmds[0].Quads || len(d.Vertices) != 16 {
		t.Fatalf("cmds=%+v verts=%d", d.Cmds, len(d.Vertices))
	}
	// thick stroke falls back to the path stroker
	d = NewDrawList()
	d.AddRect(Vec2{10, 10}, Vec2{14, 14}, red, 0, 8)
	if d.Cmds[0].Quads {
		t.Fatal("overlapping stroke must not use quads")
	}
	// rounded stroke stays general
	d = NewDrawList()
	d.AddRect(Vec2{10, 10}, Vec2{40, 40}, red, 6, 1)
	if d.Cmds[0].Quads {
		t.Fatal("rounded stroke must not use quads")
	}
}

func TestCullOutsideClip(t *testing.T) {
	d := NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{100, 100}})
	d.AddRectFilled(Vec2{200, 0}, Vec2{210, 10}, red, 0)
	d.AddRect(Vec2{0, 200}, Vec2{10, 210}, red, 0, 1)
	d.AddRectFilled(Vec2{200, 0}, Vec2{210, 10}, red, 4)
	d.AddLine(Vec2{-50, 5}, Vec2{-10, 5}, red, 1)
	d.AddLine(Vec2{-50, -5}, Vec2{-10, 500}, red, 1)
	d.AddTriangleFilled(Vec2{300, 0}, Vec2{310, 0}, Vec2{300, 10}, red)
	d.AddCircleFilled(Vec2{300, 300}, 5, red)
	d.AddCircle(Vec2{300, 300}, 5, 1, red)
	d.AddImage(5, Vec2{150, 0}, Vec2{160, 10}, [2]float32{}, [2]float32{1, 1}, red)
	if len(d.Vertices) != 0 || len(d.Cmds) != 0 {
		t.Fatalf("expected nothing, got %d verts %d cmds", len(d.Vertices), len(d.Cmds))
	}
	// text keeps only quads that touch the clip
	qs := []GlyphQuad{{Min: Vec2{10, 10}, Max: Vec2{20, 20}}, {Min: Vec2{150, 10}, Max: Vec2{160, 20}}, {Min: Vec2{90, 10}, Max: Vec2{110, 20}}}
	d.AddText(5, qs, red)
	if len(d.Vertices) != 8 || d.Cmds[0].IdxCount != 12 {
		t.Fatalf("verts=%d cmds=%+v", len(d.Vertices), d.Cmds)
	}
	// touching edge and empty clip
	d = NewDrawList()
	d.PushClipRect(Rect{Vec2{0, 0}, Vec2{100, 100}})
	d.AddRectFilled(Vec2{100, 0}, Vec2{110, 10}, red, 0)
	d.PushClipRect(Rect{Vec2{50, 50}, Vec2{50, 80}})
	d.AddRectFilled(Vec2{0, 0}, Vec2{200, 200}, red, 0)
	if len(d.Vertices) != 0 {
		t.Fatal("edge touching and empty clip must cull")
	}
}
