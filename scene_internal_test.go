package goak

import (
	"errors"
	"runtime"
	"testing"

	"github.com/tacf/goak/components"

	"github.com/Zyko0/go-sdl3/sdl"
)

type lifecycleScene struct {
	initErr     error
	initialized bool
	closed      bool
	ui          *components.UI
}

func (scene *lifecycleScene) Init(ctx *SceneContext) error {
	scene.initialized = ctx != nil
	scene.ui = ctx.UI()
	return scene.initErr
}

func TestRunSceneWithUIAttachesBeforeInitializationAndCleansUp(t *testing.T) {
	app := NewApp()
	app.win = &Window{}
	ui := components.NewUI()
	scene := &lifecycleScene{}
	if err := app.RunSceneWithUI(scene, ui); !errors.Is(err, ErrWindowNotInitialized) {
		t.Fatalf("RunSceneWithUI error = %v, want %v", err, ErrWindowNotInitialized)
	}
	if scene.ui != ui {
		t.Fatal("scene initializer did not receive the configured retained UI")
	}
	if app.win.ui != nil || app.win.scene != nil || app.win.sceneCtx != nil {
		t.Fatal("scene UI or lifecycle references remained attached after RunSceneWithUI")
	}
}

func (*lifecycleScene) Draw(*SceneContext) {}

func (scene *lifecycleScene) Close() { scene.closed = true }

func TestRunSceneCallsOptionalLifecycle(t *testing.T) {
	app := NewApp()
	app.win = &Window{}
	scene := &lifecycleScene{}
	if err := app.RunScene(scene); !errors.Is(err, ErrWindowNotInitialized) {
		t.Fatalf("RunScene error = %v, want %v", err, ErrWindowNotInitialized)
	}
	if !scene.initialized || !scene.closed {
		t.Fatalf("scene lifecycle = initialized %v, closed %v", scene.initialized, scene.closed)
	}
}

func TestRunSceneClosesAfterInitializationError(t *testing.T) {
	want := errors.New("initialize")
	app := NewApp()
	app.win = &Window{}
	scene := &lifecycleScene{initErr: want}
	if err := app.RunScene(scene); !errors.Is(err, want) {
		t.Fatalf("RunScene error = %v, want %v", err, want)
	}
	if !scene.closed {
		t.Fatal("scene was not closed after initialization failure")
	}
}

func TestKeyChordUsesStableApplicationBindings(t *testing.T) {
	tests := []struct {
		key  sdl.Keycode
		mod  sdl.Keymod
		want string
	}{
		{sdl.K_P, sdl.KMOD_CTRL | sdl.KMOD_SHIFT, "ctrl+shift+p"},
		{sdl.K_LEFT, 0, "left"},
		{sdl.K_F10, 0, "f10"},
		{sdl.K_SLASH, sdl.KMOD_CTRL, "ctrl+slash"},
	}
	for _, test := range tests {
		if got := keyChord(test.key, test.mod); got != test.want {
			t.Errorf("keyChord(%v, %v) = %q, want %q", test.key, test.mod, got, test.want)
		}
	}
}

func TestModifiersNormalizePrimaryShortcut(t *testing.T) {
	if got := modifiers(sdl.KMOD_CTRL); !got.Primary || !got.Control {
		t.Fatalf("Control modifiers = %+v", got)
	}
	got := modifiers(sdl.KMOD_GUI)
	if got.Primary != (runtime.GOOS == "darwin") || !got.Super {
		t.Fatalf("GUI modifiers = %+v on %s", got, runtime.GOOS)
	}
}

func TestRendererPointAppliesPixelDensity(t *testing.T) {
	x, y := rendererPoint(100, 50, 1.5)
	if x != 150 || y != 75 {
		t.Fatalf("renderer point = (%v, %v), want (150, 75)", x, y)
	}
}
