package sdl3

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/font"

	"github.com/tacf/goak"
)

func TestNewFrameQuit(t *testing.T) {
	b := newTestBackend(t)
	var in goak.Input
	if !b.NewFrame(&in) || in.Quit {
		t.Fatal("fresh backend must keep running")
	}
	if in.DT < 0 || in.DT > maxDT {
		t.Fatalf("dt %v", in.DT)
	}
	if err := sdl.PushEvent(&sdl.Event{Type: sdl.EVENT_QUIT}); err != nil {
		t.Skipf("push event: %v", err)
	}
	if b.NewFrame(&in) || !in.Quit {
		t.Fatal("quit event must end the loop")
	}
}

func TestScaleChanged(t *testing.T) {
	b := newTestBackend(t)
	if _, ok := b.ScaleChanged(); ok {
		t.Fatal("no change expected")
	}
	b.seenScale = 3
	s, ok := b.ScaleChanged()
	if !ok || s != b.Scale() {
		t.Fatalf("got %v %v", s, ok)
	}
	if _, ok := b.ScaleChanged(); ok {
		t.Fatal("change must be reported once")
	}
}

func widgetFrame(ctx *goak.Context, in *goak.Input, b *Backend) {
	ctx.NewFrame(in, b.DisplaySize())
	ctx.SetNextWindowPos(goak.Vec2{X: 4, Y: 4}, goak.CondAlways)
	if ctx.Begin("Win", nil, 0) {
		ctx.Button("Go")
		ctx.Text("hello")
	}
	ctx.End()
}

func countNonBG(t *testing.T, b *Backend, bg colors.Color) int {
	t.Helper()
	w, h, rgba, err := b.ReadPixels()
	if err != nil {
		t.Skipf("readback unavailable: %v", err)
	}
	n := 0
	for i := 0; i < w*h; i++ {
		if rgba[i*4] != bg.R || rgba[i*4+1] != bg.G || rgba[i*4+2] != bg.B {
			n++
		}
	}
	return n
}

func TestRenderDrawDataWidgets(t *testing.T) {
	b := newTestBackend(t)
	fnt, err := font.New(nil, 14, float64(b.Scale()))
	if err != nil {
		t.Fatal(err)
	}
	ctx := goak.New(fnt, b)
	var in goak.Input
	bg := colors.RGB(24, 24, 28)
	for i := 0; i < 3; i++ {
		b.NewFrame(&in)
		widgetFrame(ctx, &in, b)
		dd := ctx.Render()
		_ = b.Clear(bg)
		if err := b.RenderDrawData(dd); err != nil {
			t.Fatal(err)
		}
		_ = b.Present()
	}
	if n := countNonBG(t, b, bg); n < 100 {
		t.Fatalf("only %d non background pixels", n)
	}
}

func TestSetFontRebuildRendersWithoutStaleTexture(t *testing.T) {
	b := newTestBackend(t)
	fnt, _ := font.New(nil, 14, 1)
	ctx := goak.New(fnt, b)
	var in goak.Input
	bg := colors.RGB(24, 24, 28)
	frame := func() {
		b.NewFrame(&in)
		widgetFrame(ctx, &in, b)
		dd := ctx.Render()
		_ = b.Clear(bg)
		if err := b.RenderDrawData(dd); err != nil {
			t.Fatal(err)
		}
	}
	frame()
	oldTex := fnt.Atlas().Texture()
	before := len(b.textures)
	if err := ctx.RebuildFont(nil, 14, 1); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.textures[oldTex]; ok {
		t.Fatal("old atlas texture still alive")
	}
	frame()
	if len(b.textures) != before {
		t.Fatalf("textures %d, want %d", len(b.textures), before)
	}
	if ctx.TextErr != nil {
		t.Fatal(ctx.TextErr)
	}
	if n := countNonBG(t, b, bg); n < 100 {
		t.Fatalf("only %d non background pixels after rebuild", n)
	}
}
