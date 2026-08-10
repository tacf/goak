package goak

import (
	"errors"
	"math"
	"testing"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

func TestSceneAndUIMouseCaptureAreReconciled(t *testing.T) {
	var calls []bool
	win := &Window{
		handle: new(sdl.Window),
		captureMouse: func(capture bool) error {
			calls = append(calls, capture)
			return nil
		},
	}
	ctx := &SceneContext{window: win}
	if err := ctx.CaptureMouse(true); err != nil {
		t.Fatal(err)
	}
	if err := win.setUIMouseCapture(true); err != nil {
		t.Fatal(err)
	}
	win.releaseUIMouseCapture()
	if !win.mouseCaptureActive {
		t.Fatal("releasing UI capture canceled the Scene capture")
	}
	if err := ctx.CaptureMouse(false); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || !calls[0] || calls[1] {
		t.Fatalf("effective capture calls = %v, want [true false]", calls)
	}
}

func TestInvalidFontRevisionIsAttemptedOnceAndExposed(t *testing.T) {
	want := errors.New("invalid font")
	attempts := 0
	win := &Window{
		renderer: new(sdl.Renderer),
		openFont: func(*sdl.Renderer, []byte, float32, float32) (*rendering.Font, error) {
			attempts++
			return nil, want
		},
	}
	ui := components.NewUI()
	ui.SetFontData([]byte("bad"))

	for range 2 {
		if err := win.ensureFont(1, ui); !errors.Is(err, want) {
			t.Fatalf("ensureFont error = %v, want %v", err, want)
		}
	}
	if attempts != 1 {
		t.Fatalf("unchanged invalid font attempts = %d, want 1", attempts)
	}
	if !errors.Is(win.UIFontError(), want) {
		t.Fatalf("UIFontError = %v, want %v", win.UIFontError(), want)
	}

	ui.SetFontData([]byte("different"))
	_ = win.ensureFont(1, ui)
	if attempts != 2 {
		t.Fatalf("new font revision attempts = %d, want 2", attempts)
	}
}

func TestWindowScaleRejectsInvalidValuesWithoutChangingState(t *testing.T) {
	win := &Window{windowScale: 2}
	for _, scale := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if err := win.SetWindowScale(scale); !errors.Is(err, ErrInvalidWindowScale) {
			t.Fatalf("SetWindowScale(%v) error = %v", scale, err)
		}
		if win.WindowScale() != 2 {
			t.Fatalf("invalid scale changed window scale to %v", win.WindowScale())
		}
	}
}

func TestRetainedRunReportsPreconditionErrors(t *testing.T) {
	app := NewApp()
	if err := app.Run(components.NewUI()); !errors.Is(err, ErrWindowNotInitialized) {
		t.Fatalf("Run before initialization error = %v", err)
	}
	app.win = new(Window)
	if err := app.Run(nil); !errors.Is(err, ErrNilUI) {
		t.Fatalf("Run nil UI error = %v", err)
	}
}

func TestPopupConsumesEveryPointerButtonBeforeScene(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(100), layout.StaticPx(100))
	dropdown := panel.CreateDropdown(layout.StaticPx(80), layout.StaticPx(20), "Pick", []components.DropdownOption{{Label: "One"}})
	dropdown.Container().Bounds.W = 80
	dropdown.Container().Bounds.H = 20
	dropdown.Open()
	win := &Window{ui: ui}

	if !win.mouseDown(MouseMiddle, 150, 150) {
		t.Fatal("middle click leaked through an open popup")
	}
	if dropdown.IsOpen() {
		t.Fatal("middle click did not dismiss popup")
	}
	dropdown.Open()
	if !win.mouseDown(MouseRight, 150, 150) {
		t.Fatal("right click leaked through an open popup")
	}
}

func TestRendererStateIsResetBetweenSceneAndRetainedLayers(t *testing.T) {
	state := new(recordingRendererState)
	if err := resetRendererStateOn(state, 1.5); err != nil {
		t.Fatal(err)
	}
	if state.target != nil {
		t.Fatalf("render target = %p, want window target", state.target)
	}
	if state.logicalWidth != 0 || state.logicalHeight != 0 || state.logicalMode != sdl.LOGICAL_PRESENTATION_DISABLED {
		t.Fatalf("logical presentation = %v×%v %v", state.logicalWidth, state.logicalHeight, state.logicalMode)
	}
	if state.viewport != nil || state.clip != nil {
		t.Fatalf("viewport/clip = %v/%v, want nil/nil", state.viewport, state.clip)
	}
	if state.scaleX != 1.5 || state.scaleY != 1.5 || state.colorScale != 1 {
		t.Fatalf("scale/color scale = %v,%v/%v", state.scaleX, state.scaleY, state.colorScale)
	}
	if state.blendMode != sdl.BLENDMODE_BLEND {
		t.Fatalf("renderer blend mode = %v", state.blendMode)
	}
}

type recordingRendererState struct {
	target                      *sdl.Texture
	logicalWidth, logicalHeight int32
	logicalMode                 sdl.RendererLogicalPresentation
	viewport, clip              *sdl.Rect
	scaleX, scaleY, colorScale  float32
	blendMode                   sdl.BlendMode
}

func (state *recordingRendererState) SetRenderTarget(target *sdl.Texture) error {
	state.target = target
	return nil
}

func (state *recordingRendererState) SetLogicalPresentation(width, height int32, mode sdl.RendererLogicalPresentation) error {
	state.logicalWidth, state.logicalHeight, state.logicalMode = width, height, mode
	return nil
}

func (state *recordingRendererState) SetViewport(viewport *sdl.Rect) error {
	state.viewport = viewport
	return nil
}

func (state *recordingRendererState) SetClipRect(clip *sdl.Rect) error {
	state.clip = clip
	return nil
}

func (state *recordingRendererState) SetScale(x, y float32) error {
	state.scaleX, state.scaleY = x, y
	return nil
}

func (state *recordingRendererState) SetColorScale(scale float32) error {
	state.colorScale = scale
	return nil
}

func (state *recordingRendererState) SetDrawBlendMode(mode sdl.BlendMode) error {
	state.blendMode = mode
	return nil
}
