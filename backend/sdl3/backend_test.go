package sdl3

import (
	"os"
	"testing"

	"github.com/tacf/goak"
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/font"
)

func px(rgba []byte, w, x, y int) [4]byte {
	o := (y*w + x) * 4
	return [4]byte{rgba[o], rgba[o+1], rgba[o+2], rgba[o+3]}
}

func newTestBackend(t *testing.T) *Backend {
	t.Helper()
	if os.Getenv("SDL_VIDEODRIVER") == "" {
		t.Setenv("SDL_VIDEODRIVER", "dummy")
	}
	b, err := New(Config{Title: "test", Width: 64, Height: 64, Renderer: "software", NoHighDPI: true, NoVSync: true})
	if err != nil {
		t.Skipf("SDL unavailable: %v", err)
	}
	t.Cleanup(b.Close)
	return b
}

func TestRenderRectAndTexturedQuad(t *testing.T) {
	b := newTestBackend(t)
	if b.RendererName() == "" {
		t.Fatal("empty renderer name")
	}
	// 2x2 solid blue texture.
	tex := make([]byte, 2*2*4)
	for i := 0; i < 4; i++ {
		copy(tex[i*4:], []byte{0, 0, 255, 255})
	}
	id, err := b.NewTexture(2, 2, tex)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.UpdateTexture(id, 2, 2, tex); err != nil {
		t.Fatal(err)
	}
	dl := goak.NewDrawList()
	dl.AddRectFilled(goak.Vec2{X: 4, Y: 4}, goak.Vec2{X: 24, Y: 24}, colors.RGB(255, 0, 0), 0)
	dl.AddImage(id, goak.Vec2{X: 36, Y: 36}, goak.Vec2{X: 56, Y: 56},
		[2]float32{0, 0}, [2]float32{1, 1}, colors.RGB(255, 255, 255))
	// Clipped rect must not draw outside its clip.
	dl.PushClipRect(goak.Rect{Min: goak.Vec2{X: 40, Y: 4}, Max: goak.Vec2{X: 50, Y: 14}})
	dl.AddRectFilled(goak.Vec2{X: 30, Y: 0}, goak.Vec2{X: 60, Y: 20}, colors.RGB(0, 255, 0), 0)
	dl.PopClipRect()
	// Empty clip skips the command.
	dl.PushClipRect(goak.Rect{Min: goak.Vec2{X: 10, Y: 40}, Max: goak.Vec2{X: 10, Y: 50}})
	dl.AddRectFilled(goak.Vec2{X: 0, Y: 40}, goak.Vec2{X: 30, Y: 60}, colors.RGB(255, 255, 0), 0)
	dl.PopClipRect()

	if err := b.Clear(colors.RGB(0, 0, 0)); err != nil {
		t.Fatal(err)
	}
	dd := &goak.DrawData{Lists: []*goak.DrawList{dl}, DisplaySize: goak.Vec2{X: 64, Y: 64}, Scale: goak.Vec2{X: 1, Y: 1}}
	if err := b.Render(dd); err != nil {
		t.Fatal(err)
	}
	w, _, rgba, err := b.ReadPixels()
	if err != nil {
		t.Fatal(err)
	}
	check := func(name string, x, y int, want [4]byte) {
		t.Helper()
		if got := px(rgba, w, x, y); got != want {
			t.Errorf("%s (%d,%d) = %v, want %v", name, x, y, got, want)
		}
	}
	check("rect", 14, 14, [4]byte{255, 0, 0, 255})
	check("quad", 46, 46, [4]byte{0, 0, 255, 255})
	check("clipped inside", 45, 8, [4]byte{0, 255, 0, 255})
	check("clipped outside", 35, 8, [4]byte{0, 0, 0, 255})
	check("empty clip", 5, 50, [4]byte{0, 0, 0, 255})
	check("background", 30, 30, [4]byte{0, 0, 0, 255})

	b.DestroyTexture(id)
	if err := b.Present(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderUnknownTexture(t *testing.T) {
	b := newTestBackend(t)
	dl := goak.NewDrawList()
	dl.AddImage(99, goak.Vec2{}, goak.Vec2{X: 8, Y: 8}, [2]float32{}, [2]float32{1, 1}, colors.RGB(255, 255, 255))
	dd := &goak.DrawData{Lists: []*goak.DrawList{dl}, Scale: goak.Vec2{X: 1, Y: 1}}
	if err := b.Render(dd); err == nil {
		t.Fatal("want error for unknown texture")
	}
}

func TestPixelClipFloatError(t *testing.T) {
	// 0.1*3 style error must not widen the rect by a pixel.
	r, ok := pixelClip(goak.Rect{Min: goak.Vec2{X: 9.999999, Y: 2.0000002}, Max: goak.Vec2{X: 20.000002, Y: 30.999998}}, 1, 1, 100, 80)
	if !ok || r.X != 10 || r.Y != 2 || r.W != 10 || r.H != 29 {
		t.Fatalf("got %+v", r)
	}
}

func TestTextInputToggle(t *testing.T) {
	b := newTestBackend(t)
	if err := b.StopTextInput(); err != nil {
		t.Fatal(err)
	}
	if err := b.StartTextInput(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderText(t *testing.T) {
	b := newTestBackend(t)
	f, err := font.New(nil, 20, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dl := goak.NewDrawList()
	if err := f.DrawText(dl, b, 4, 4, "Hello", colors.RGB(255, 255, 255)); err != nil {
		t.Fatal(err)
	}
	if err := b.Clear(colors.RGB(0, 0, 0)); err != nil {
		t.Fatal(err)
	}
	dd := &goak.DrawData{Lists: []*goak.DrawList{dl}, DisplaySize: goak.Vec2{X: 64, Y: 64}, Scale: goak.Vec2{X: 1, Y: 1}}
	if err := b.Render(dd); err != nil {
		t.Fatal(err)
	}
	w, h, rgba, err := b.ReadPixels()
	if err != nil {
		t.Fatal(err)
	}
	w0, h0 := f.Measure("Hello")
	lit, outside := 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := px(rgba, w, x, y)
			if p[0] == 0 && p[1] == 0 && p[2] == 0 {
				continue
			}
			if float64(x) >= 3 && float64(x) <= 5+w0 && float64(y) >= 3 && float64(y) <= 5+h0 {
				lit++
			} else {
				outside++
			}
		}
	}
	if lit < 30 {
		t.Fatalf("only %d text pixels in expected region", lit)
	}
	if outside != 0 {
		t.Fatalf("%d lit pixels outside text region", outside)
	}
}

func TestPixelClip(t *testing.T) {
	inf := goak.Rect{Min: goak.Vec2{X: -1e38, Y: -1e38}, Max: goak.Vec2{X: 1e38, Y: 1e38}}
	r, ok := pixelClip(inf, 2, 2, 100, 80)
	if !ok || r.X != 0 || r.Y != 0 || r.W != 100 || r.H != 80 {
		t.Fatalf("infinite clip: %+v %v", r, ok)
	}
	r, ok = pixelClip(goak.Rect{Min: goak.Vec2{X: 1.2, Y: 1}, Max: goak.Vec2{X: 3.1, Y: 2}}, 2, 2, 100, 80)
	if !ok || r.X != 2 || r.W != 5 || r.Y != 2 || r.H != 2 {
		t.Fatalf("scaled clip: %+v", r)
	}
	if _, ok = pixelClip(goak.Rect{Min: goak.Vec2{X: 5, Y: 5}, Max: goak.Vec2{X: 5, Y: 9}}, 1, 1, 100, 80); ok {
		t.Fatal("empty clip must report false")
	}
	if _, ok = pixelClip(goak.Rect{Min: goak.Vec2{X: 200, Y: 0}, Max: goak.Vec2{X: 300, Y: 9}}, 1, 1, 100, 80); ok {
		t.Fatal("offscreen clip must report false")
	}
}
