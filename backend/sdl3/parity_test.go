package sdl3

import (
	"os"
	"testing"

	"github.com/tacf/goak"
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/font"
)

func newSizedBackend(t *testing.T, w, h int) *Backend {
	t.Helper()
	if os.Getenv("SDL_VIDEODRIVER") == "" {
		t.Setenv("SDL_VIDEODRIVER", "dummy")
	}
	b, err := New(Config{Title: "test", Width: w, Height: h, Renderer: "software", NoHighDPI: true, NoVSync: true})
	if err != nil {
		t.Skipf("SDL unavailable: %v", err)
	}
	t.Cleanup(b.Close)
	return b
}

func gradientTex(n int) []byte {
	p := make([]byte, n*n*4)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			o := (y*n + x) * 4
			p[o], p[o+1], p[o+2], p[o+3] = byte(x*255/(n-1)), byte(y*255/(n-1)), 160, byte(128+x*127/(n-1))
		}
	}
	return p
}

func renderScene(t *testing.T, b *Backend, fast bool, dl *goak.DrawList) []byte {
	t.Helper()
	b.SetFastPath(fast)
	if err := b.Clear(colors.RGB(30, 40, 50)); err != nil {
		t.Fatal(err)
	}
	dd := &goak.DrawData{Lists: []*goak.DrawList{dl}, DisplaySize: goak.Vec2{X: 200, Y: 160}, Scale: goak.Vec2{X: 1, Y: 1}}
	if err := b.Render(dd); err != nil {
		t.Fatal(err)
	}
	_, _, rgba, err := b.ReadPixels()
	if err != nil {
		t.Fatal(err)
	}
	return rgba
}

// diff returns the max channel difference and the count of pixels above tol.
func diff(a, b []byte, tol int) (maxd, bad int) {
	for i := 0; i+3 < len(a); i += 4 {
		worst := 0
		for k := 0; k < 4; k++ {
			d := int(a[i+k]) - int(b[i+k])
			if d < 0 {
				d = -d
			}
			worst = max(worst, d)
		}
		maxd = max(maxd, worst)
		if worst > tol {
			bad++
		}
	}
	return
}

func v2(x, y float32) goak.Vec2 { return goak.Vec2{X: x, Y: y} }

// TestFastPathParity checks that the quad fast path matches the geometry path.
func TestFastPathParity(t *testing.T) {
	b := newSizedBackend(t, 200, 160)
	img, err := b.NewTexture(16, 16, gradientTex(16))
	if err != nil {
		t.Fatal(err)
	}
	f, err := font.New(nil, 16, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dl := goak.NewDrawList()
	dl.AddRectFilled(v2(4, 4), v2(60, 40), colors.RGB(200, 60, 60), 0)
	dl.AddRectFilled(v2(30, 20), v2(90, 60), colors.RGBA(60, 200, 90, 128), 0)
	dl.AddRectFilled(v2(10.5, 70), v2(50.5, 90.5), colors.RGB(250, 250, 0), 0) // fractional
	dl.AddRectFilled(v2(100, 4), v2(190, 70), colors.RGB(70, 70, 90), 6)       // split rounded
	dl.AddRectFilled(v2(110, 10), v2(180, 30), colors.RGBA(255, 255, 255, 90), 4)
	dl.AddRect(v2(100, 4), v2(190, 70), colors.RGB(255, 255, 255), 6, 1)
	dl.AddRect(v2(6, 100), v2(70, 140), colors.RGBA(255, 200, 0, 160), 0, 2)
	dl.AddRect(v2(80, 100), v2(90, 110), colors.RGB(0, 255, 255), 0, 6) // thick falls back
	dl.AddLine(v2(100, 90), v2(180, 90), colors.RGB(255, 0, 255), 1)
	dl.AddLine(v2(100, 95), v2(100, 150), colors.RGB(255, 0, 255), 3)
	dl.AddLine(v2(110, 95), v2(190, 150), colors.RGB(0, 255, 0), 1.5)
	dl.AddTriangleFilled(v2(150, 100), v2(190, 100), v2(190, 140), colors.RGB(0, 120, 255))
	dl.AddCircleFilled(v2(40, 120), 12, colors.RGBA(255, 128, 0, 200))
	dl.AddCircle(v2(80, 125), 10, 2, colors.RGB(255, 255, 255))
	// images: 1:1 tinted, translucent, scaled and sub-rect
	dl.AddImage(img, v2(120, 100), v2(136, 116), [2]float32{0, 0}, [2]float32{1, 1}, colors.RGB(255, 255, 255))
	dl.AddImage(img, v2(140, 100), v2(156, 116), [2]float32{0, 0}, [2]float32{1, 1}, colors.RGBA(255, 128, 64, 180))
	dl.AddImage(img, v2(120, 120), v2(152, 152), [2]float32{0, 0}, [2]float32{1, 1}, colors.RGB(255, 255, 255))
	dl.AddImage(img, v2(160, 120), v2(168, 128), [2]float32{0.25, 0.25}, [2]float32{0.75, 0.75}, colors.RGB(255, 255, 255))
	// nested clips with text and fills crossing the edges
	dl.PushClipRect(goak.Rect{Min: v2(8, 5), Max: v2(95, 55)})
	dl.PushClipRect(goak.Rect{Min: v2(0, 0), Max: v2(70.5, 40)})
	if err := f.DrawText(dl, b, 2, 4, "Clipped text sample", colors.RGB(255, 255, 255)); err != nil {
		t.Fatal(err)
	}
	dl.AddRectFilled(v2(0, 30), v2(120, 45), colors.RGBA(0, 0, 0, 100), 0)
	dl.PopClipRect()
	if err := f.DrawText(dl, b, 2, 40, "Outer clip text", colors.RGBA(255, 220, 0, 220)); err != nil {
		t.Fatal(err)
	}
	dl.PopClipRect()
	if err := f.DrawText(dl, b, 8, 72, "Plain text 123", colors.RGB(240, 240, 240)); err != nil {
		t.Fatal(err)
	}
	f.Finish(dl)

	nQuad := 0
	for _, c := range dl.Cmds {
		if c.Quads {
			nQuad++
		}
	}
	if nQuad == 0 {
		t.Fatal("scene has no quad cmds")
	}
	slow := renderScene(t, b, false, dl)
	if b.nFill+b.nCopy != 0 {
		t.Fatal("fast path used while disabled")
	}
	fast := renderScene(t, b, true, dl)
	t.Logf("fills %d copies %d geometry quads %d", b.nFill, b.nCopy, b.nSlow)
	if b.nFill == 0 || b.nCopy < 20 || b.nSlow == 0 {
		t.Fatalf("scene must hit every path: fills %d copies %d slow %d", b.nFill, b.nCopy, b.nSlow)
	}
	maxd, bad := diff(slow, fast, 2)
	t.Logf("quad cmds %d of %d, max diff %d, pixels above tol %d", nQuad, len(dl.Cmds), maxd, bad)
	if bad != 0 {
		t.Fatalf("%d pixels differ by more than 2 (max %d)", bad, maxd)
	}
}

// TestRoundedAndStrokeMatchLegacyPath compares the split rounded fill and the
// four-fill stroke against the single fan and the path stroker.
func TestRoundedAndStrokeMatchLegacyPath(t *testing.T) {
	b := newSizedBackend(t, 200, 160)
	white := colors.RGB(230, 230, 240)

	cur := goak.NewDrawList()
	cur.AddRectFilled(v2(10, 10), v2(110, 90), white, 8)
	cur.AddRectFilledEx(v2(120, 10), v2(190, 90), colors.RGB(90, 160, 220), 10, goak.CornerTopLeft|goak.CornerBottomRight)
	cur.AddRect(v2(10, 100), v2(100, 150), white, 0, 2)
	cur.AddRect(v2(110, 100), v2(190, 150), white, 0, 1)

	old := goak.NewDrawList()
	old.PathRect(v2(10, 10), v2(110, 90), 8, goak.CornerAll)
	old.PathFillConvex(goak.Pack(white))
	old.PathRect(v2(120, 10), v2(190, 90), 10, goak.CornerTopLeft|goak.CornerBottomRight)
	old.PathFillConvex(goak.Pack(colors.RGB(90, 160, 220)))
	old.PathRect(v2(10, 100), v2(100, 150), 0, goak.CornerAll)
	old.PathStroke(goak.Pack(white), true, 2)
	old.PathRect(v2(110, 100), v2(190, 150), 0, goak.CornerAll)
	old.PathStroke(goak.Pack(white), true, 1)

	a := renderScene(t, b, true, cur)
	c := renderScene(t, b, false, old)
	maxd, bad := diff(a, c, 2)
	t.Logf("max diff %d, pixels above tol %d", maxd, bad)
	if bad != 0 {
		t.Fatalf("%d pixels differ from legacy output (max %d)", bad, maxd)
	}
}

// TestFractionalSplitFillHasNoSeams renders large rounded fills at fractional
// positions. Their strips share edges and must leave no unfilled row.
func TestFractionalSplitFillHasNoSeams(t *testing.T) {
	b := newSizedBackend(t, 300, 120)
	for i := 0; i < 20; i++ {
		y0 := 10 + float32(i)*0.05
		y1 := y0 + 24.6
		dl := goak.NewDrawList()
		dl.AddRectFilled(v2(20, y0), v2(220, y1), colors.RGB(200, 50, 50), 3)
		if err := b.Clear(colors.RGB(0, 0, 0)); err != nil {
			t.Fatal(err)
		}
		dd := &goak.DrawData{Lists: []*goak.DrawList{dl}, DisplaySize: v2(300, 120), Scale: v2(1, 1)}
		if err := b.Render(dd); err != nil {
			t.Fatal(err)
		}
		w, _, px, err := b.ReadPixels()
		if err != nil {
			t.Fatal(err)
		}
		for y := int(y0) + 2; y < int(y1)-1; y++ {
			if px[(y*w+100)*4] < 100 {
				t.Fatalf("y0=%.2f: seam at row %d", y0, y)
			}
		}
	}
}
