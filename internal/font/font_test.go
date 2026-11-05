package font

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"testing"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

type fakeBackend struct {
	news, updates int
	w, h          int
	px            []byte
}

func (b *fakeBackend) NewTexture(w, h int, rgba []byte) (draw.TextureID, error) {
	b.news++
	b.w, b.h, b.px = w, h, append([]byte(nil), rgba...)
	return 7, nil
}
func (b *fakeBackend) UpdateTexture(id draw.TextureID, w, h int, rgba []byte) error {
	if id != 7 {
		return fmt.Errorf("bad id %d", id)
	}
	b.updates++
	b.w, b.h, b.px = w, h, append([]byte(nil), rgba...)
	return nil
}
func (b *fakeBackend) DestroyTexture(draw.TextureID) {}
func (b *fakeBackend) Render(*draw.DrawData) error   { return nil }

func TestShelfPackNoOverlap(t *testing.T) {
	a := NewAtlas(128, 128, 128)
	type r struct{ x, y, w, h int }
	var got []r
	sizes := [][2]int{{10, 12}, {8, 20}, {30, 9}, {12, 12}, {5, 5}, {40, 14}, {9, 11}}
	for i := 0; i < 60; i++ {
		s := sizes[i%len(sizes)]
		x, y, ok := a.Alloc(s[0], s[1])
		if !ok {
			continue
		}
		if x+s[0] > 128 || y+s[1] > 128 {
			t.Fatalf("out of bounds %d,%d", x, y)
		}
		got = append(got, r{x, y, s[0], s[1]})
	}
	if len(got) < 10 {
		t.Fatalf("too few placed: %d", len(got))
	}
	for i := range got {
		for j := i + 1; j < len(got); j++ {
			p, q := got[i], got[j]
			// Padding of 1px means regions must not even touch.
			if p.x < q.x+q.w+1 && q.x < p.x+p.w+1 && p.y < q.y+q.h+1 && q.y < p.y+p.h+1 {
				t.Fatalf("overlap %+v %+v", p, q)
			}
		}
	}
}

func TestGrowthKeepsGlyphs(t *testing.T) {
	f, err := New(nil, 24, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.atlas = NewAtlas(32, 32, 1024)
	var sb strings.Builder
	for r := rune('A'); r <= 'z'; r++ {
		sb.WriteRune(r)
	}
	text := sb.String()
	q := f.Layout(text, 0, 0)
	if f.atlas.Revision() == 0 {
		t.Fatal("expected growth")
	}
	w, h := f.atlas.Size()
	if w < 32 || h <= 32 {
		t.Fatalf("size %dx%d", w, h)
	}
	for r, g := range f.glyphs {
		if g.w == 0 {
			continue
		}
		// Every glyph region must hold some coverage after repack.
		var sum int
		for y := g.y; y < g.y+g.h; y++ {
			for x := g.x; x < g.x+g.w; x++ {
				sum += int(f.atlas.pix[(y*w+x)*4+3])
			}
		}
		if sum == 0 {
			t.Fatalf("glyph %q empty after growth", r)
		}
	}
	// Quads are valid: UVs inside [0,1].
	for _, qq := range q {
		for _, u := range append(qq.UVMin[:], qq.UVMax[:]...) {
			if u < 0 || u > 1 {
				t.Fatalf("uv out of range %v", qq)
			}
		}
	}
	// Layout again must give identical quads.
	q2 := f.Layout(text, 0, 0)
	if len(q) != len(q2) || math.Abs(float64(q[10].UVMin[1]-q2[10].UVMin[1])) > 1e-5 || q[10].Min != q2[10].Min {
		t.Fatal("layout not stable after growth")
	}
}

func TestAtlasFullDropsGlyph(t *testing.T) {
	f, _ := New(nil, 40, 1)
	f.atlas = NewAtlas(16, 16, 16)
	q := f.Layout("WWW", 0, 0)
	if len(q) != 0 {
		t.Fatalf("want no quads in a full atlas, got %d", len(q))
	}
}

func TestMeasureParity(t *testing.T) {
	const size, scale = 16.0, 2.0
	f, err := New(nil, size, scale)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := opentype.Parse(goregular.TTF)
	face, _ := opentype.NewFace(parsed, &opentype.FaceOptions{Size: size * scale, DPI: 72, Hinting: font.HintingNone})
	s := "Hello, World 123"
	var want float64
	for _, r := range s {
		adv, _ := face.GlyphAdvance(r)
		want += float64(adv) / 64 / scale
	}
	w, h := f.Measure(s)
	if math.Abs(w-want) > 1e-3 || h != f.LineHeight() || h <= 0 {
		t.Fatalf("measure %v,%v want %v", w, h, want)
	}
	// Tab advances to the next multiple of four spaces.
	sp, _ := f.Measure(" ")
	tw, _ := f.Measure("\t")
	if math.Abs(tw-4*sp) > 1e-3 {
		t.Fatalf("tab %v want %v", tw, 4*sp)
	}
	tw2, _ := f.Measure(" \t")
	if math.Abs(tw2-4*sp) > 1e-3 {
		t.Fatalf("space+tab %v want %v", tw2, 4*sp)
	}
	if w0, _ := (*Font)(nil).Measure("x"); w0 != 0 {
		t.Fatal("nil font measure")
	}
}

func TestLayoutMonotonic(t *testing.T) {
	f, _ := New(nil, 18, 1.5)
	q := f.Layout("Monotonic text\tgoes right", 10, 20)
	if len(q) < 15 {
		t.Fatalf("quads %d", len(q))
	}
	prev := float32(-1)
	for i, g := range q {
		if g.Max.X <= g.Min.X || g.Max.Y <= g.Min.Y {
			t.Fatalf("empty quad %d", i)
		}
		if g.Min.X < prev-1 {
			t.Fatalf("quad %d x %v not monotonic after %v", i, g.Min.X, prev)
		}
		prev = g.Min.X
		if g.Min.X < 10-1 {
			t.Fatalf("quad left of origin")
		}
	}
	if q[0].Min.Y < 20 {
		t.Fatalf("glyph above line top: %v", q[0].Min.Y)
	}
	// Spaces produce no quad.
	if n := len(f.Layout("   ", 0, 0)); n != 0 {
		t.Fatalf("spaces gave %d quads", n)
	}
}

func TestUnknownRuneFallback(t *testing.T) {
	f, _ := New(nil, 16, 1)
	// A private use rune has no glyph in Go Regular.
	const odd = '\U0010FFFD'
	wq, _ := f.Measure(string(odd))
	wm, _ := f.Measure("?")
	if wq != wm {
		t.Fatalf("fallback advance %v want %v", wq, wm)
	}
	a := f.Layout(string(odd), 0, 0)
	b := f.Layout("?", 0, 0)
	if len(a) != 1 || len(b) != 1 || a[0].UVMin != b[0].UVMin {
		t.Fatalf("fallback quad mismatch %v %v", a, b)
	}
}

func TestUVExactEdges(t *testing.T) {
	f, _ := New(nil, 16, 1)
	q := f.Layout("Hi", 0, 0)
	w, h := f.atlas.Size()
	for _, g := range q {
		for i, u := range []float32{g.UVMin[0] * float32(w), g.UVMax[0] * float32(w), g.UVMin[1] * float32(h), g.UVMax[1] * float32(h)} {
			if math.Abs(float64(u-float32(math.Round(float64(u))))) > 1e-3 {
				t.Fatalf("uv %d not on texel edge: %v", i, u)
			}
		}
	}
}

func TestNewlineZeroAdvance(t *testing.T) {
	f, _ := New(nil, 16, 1)
	w, _ := f.Measure("ab")
	w2, _ := f.Measure("a\r\nb")
	if w != w2 {
		t.Fatalf("measure %v want %v", w2, w)
	}
	if n := len(f.Layout("\n\r", 0, 0)); n != 0 {
		t.Fatalf("newline quads %d", n)
	}
	a, b := f.Layout("ab", 0, 0), f.Layout("a\nb", 0, 0)
	if len(a) != 2 || len(b) != 2 || a[1].Min != b[1].Min {
		t.Fatal("newline changed layout")
	}
}

// pixelsOf returns the coverage of the atlas region the UV rect addresses.
func pixelsOf(a *Atlas, uvMin, uvMax [2]float32) []byte {
	w, h := a.Size()
	x0, x1 := int(math.Round(float64(uvMin[0]*float32(w)))), int(math.Round(float64(uvMax[0]*float32(w))))
	y0, y1 := int(math.Round(float64(uvMin[1]*float32(h)))), int(math.Round(float64(uvMax[1]*float32(h))))
	var out []byte
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			out = append(out, a.pix[(y*w+x)*4+3])
		}
	}
	return out
}

func TestFinishFixesUVsAfterGrowth(t *testing.T) {
	f, _ := New(nil, 20, 1)
	f.atlas = NewAtlas(32, 32, 1024)
	be := &fakeBackend{}
	dl := draw.NewDrawList()
	col := colors.RGB(255, 255, 255)
	if err := f.DrawText(dl, be, 0, 0, "AB", col); err != nil {
		t.Fatal(err)
	}
	r0 := f.atlas.Revision()
	if err := f.DrawText(dl, be, 0, 30, "The quick brown fox jumps over the lazy dog", col); err != nil {
		t.Fatal(err)
	}
	if f.atlas.Revision() == r0 {
		t.Fatal("expected growth between calls")
	}
	f.Finish(dl)
	// Each quad's UV region must hold exactly that glyph's reference pixels.
	text := []rune("ABThequickbrownfoxjumpsoverthelazydog")
	if len(dl.Vertices) != len(text)*4 {
		t.Fatalf("verts %d want %d", len(dl.Vertices), len(text)*4)
	}
	for i, r := range text {
		g := f.glyphs[r]
		v0, v2 := dl.Vertices[i*4], dl.Vertices[i*4+2]
		got := pixelsOf(f.atlas, v0.UV, v2.UV)
		want := pixelsOf(f.atlas, f.atlas.uvOf(g), f.atlas.uvOfMax(g))
		if len(got) != g.w*g.h || string(got) != string(want) {
			t.Fatalf("glyph %q uv maps to wrong pixels", r)
		}
		var sum int
		for _, p := range got {
			sum += int(p)
		}
		if sum == 0 {
			t.Fatalf("glyph %q blank", r)
		}
	}
	// A second Finish is a no-op.
	before := append([]draw.Vertex(nil), dl.Vertices...)
	f.Finish(dl)
	for i := range before {
		if before[i] != dl.Vertices[i] {
			t.Fatal("second Finish changed vertices")
		}
	}
}

func TestUploadFlow(t *testing.T) {
	f, _ := New(nil, 16, 1)
	b := &fakeBackend{}
	f.Layout("abc", 0, 0)
	if !f.atlas.Dirty() {
		t.Fatal("atlas should be dirty")
	}
	id, err := f.atlas.Upload(b)
	if err != nil || id != 7 || b.news != 1 || f.atlas.Dirty() {
		t.Fatalf("first upload id=%v err=%v news=%d", id, err, b.news)
	}
	if _, err := f.atlas.Upload(b); err != nil || b.updates != 0 || b.news != 1 {
		t.Fatal("clean atlas must not upload")
	}
	f.Layout("abcxyz", 0, 0)
	if _, err := f.atlas.Upload(b); err != nil || b.updates != 1 {
		t.Fatalf("expected one update, got %d", b.updates)
	}
	// Growth changes the size and the same id is kept.
	f.atlas = NewAtlas(16, 16, 256)
	f.glyphs = map[rune]*glyph{}
	b2 := &fakeBackend{}
	f.Layout("a", 0, 0)
	id1, _ := f.atlas.Upload(b2)
	f.Layout("WWWWWWWWWW", 0, 0)
	id2, _ := f.atlas.Upload(b2)
	w, _ := f.atlas.Size()
	if id1 != id2 || b2.updates != 1 || b2.w != w {
		t.Fatalf("growth upload ids %v %v updates %d w %d", id1, id2, b2.updates, b2.w)
	}
}

func TestDrawText(t *testing.T) {
	f, _ := New(nil, 16, 1)
	b := &fakeBackend{}
	dl := draw.NewDrawList()
	if err := f.DrawText(dl, b, 4, 4, "Hello", colors.RGB(255, 255, 255)); err != nil {
		t.Fatal(err)
	}
	if len(dl.Vertices) != 5*4 || len(dl.Cmds) != 1 || dl.Cmds[0].TextureID != 7 {
		t.Fatalf("verts %d cmds %d", len(dl.Vertices), len(dl.Cmds))
	}
}

func TestNewErrors(t *testing.T) {
	if _, err := New([]byte("not a font"), 12, 1); err == nil {
		t.Fatal("want parse error")
	}
	if _, err := New(nil, 0, 1); err == nil {
		t.Fatal("want size error")
	}
}

func TestNoSDLDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Skipf("go list: %v", err)
	}
	if strings.Contains(strings.ToLower(string(out)), "sdl") {
		t.Fatal("internal/font must not depend on SDL")
	}
}

func benchText() string { return strings.Repeat("The quick brown fox jumps. ", 100)[:2000] }

func BenchmarkLayoutDraw2000Glyphs(b *testing.B) {
	f, _ := New(nil, 16, 1)
	be := &fakeBackend{}
	dl := draw.NewDrawList()
	s := benchText()
	_ = f.DrawText(dl, be, 0, 0, s, colors.RGB(255, 255, 255))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dl.Reset()
		_ = f.DrawText(dl, be, 0, 0, s, colors.RGB(255, 255, 255))
	}
}

func (a *Atlas) uvOf(g *glyph) [2]float32 {
	m, _ := a.UV(g.x, g.y, g.w, g.h)
	return m
}

func (a *Atlas) uvOfMax(g *glyph) [2]float32 {
	_, m := a.UV(g.x, g.y, g.w, g.h)
	return m
}
