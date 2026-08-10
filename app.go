package goak

import (
	"context"
	"errors"
	"image"
	"sync"

	"github.com/tacf/goak/components"
)

// ErrWindowNotInitialized is returned by window operations called before
// InitWindow or InitWindowWithConfig.
var ErrWindowNotInitialized = errors.New("goak: window is not initialized")

// ErrRendererAlreadyInitialized is returned when changing the renderer after
// the application window and its renderer have already been created.
var ErrRendererAlreadyInitialized = errors.New("goak: renderer must be selected before window initialization")

// ErrNilUI is returned when a retained-only run has no UI.
var ErrNilUI = errors.New("goak: retained UI is nil")

// App is the application API. Create with NewApp, call InitWindow, then Run(ui).
type App struct {
	win      *Window
	renderer RendererDriver

	dispatchMu     sync.Mutex
	dispatchQueue  []dispatchEntry
	latestDispatch map[string]int
	ctx            context.Context
	cancel         context.CancelFunc
	stopped        bool
}

// NewApp returns a new App. Call InitWindow before Run.
func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		renderer:       RendererSoftware,
		latestDispatch: make(map[string]int),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// SetRenderer selects the SDL renderer used by the next window. It must be
// called before InitWindow or InitWindowWithConfig.
func (a *App) SetRenderer(renderer RendererDriver) error {
	if a.win != nil {
		return ErrRendererAlreadyInitialized
	}
	a.renderer = normalizeRendererDriver(renderer)
	return nil
}

// RendererName returns the active SDL renderer name. Before window
// initialization it returns the configured renderer.
func (a *App) RendererName() string {
	if a.win != nil {
		return a.win.RendererName()
	}
	return string(normalizeRendererDriver(a.renderer))
}

// InitWindow creates and configures the window with the given title and size.
// Must be called before Run.
func (a *App) InitWindow(title string, width, height int) {
	a.InitWindowWithConfig(Config{
		Title:       title,
		Width:       width,
		Height:      height,
		WindowScale: 1,
		Renderer:    a.renderer,
	})
}

// InitWindowWithConfig creates and configures the window with explicit options.
func (a *App) InitWindowWithConfig(cfg Config) {
	if err := a.TryInitWindowWithConfig(cfg); err != nil {
		panic(err)
	}
}

// TryInitWindowWithConfig is the error-returning form of InitWindowWithConfig.
// It is intended for applications that surface startup failures to users
// instead of treating them as programmer errors.
func (a *App) TryInitWindowWithConfig(cfg Config) error {
	if a.win != nil {
		a.win.Destroy()
		a.win = nil
	}
	if cfg.Renderer == "" {
		cfg.Renderer = a.renderer
	}
	a.renderer = normalizeRendererDriver(cfg.Renderer)
	win, err := createWindow(cfg)
	if err != nil {
		return err
	}
	a.win = win
	return nil
}

// SetAutoDPI toggles automatic HiDPI scaling on the app window.
func (a *App) SetAutoDPI(enabled bool) {
	if a.win != nil {
		a.win.SetAutoDPI(enabled)
	}
}

// SetWindowScale sets the runtime window scale multiplier.
func (a *App) SetWindowScale(scale float64) error {
	if a.win == nil {
		return ErrWindowNotInitialized
	}
	return a.win.SetWindowScale(scale)
}

// WindowScale returns the current runtime window scale multiplier.
func (a *App) WindowScale() float64 {
	if a.win == nil {
		return 1
	}
	return a.win.WindowScale()
}

// SetScaleHotkeysEnabled toggles built-in Ctrl +/- scale shortcuts.
func (a *App) SetScaleHotkeysEnabled(enabled bool) {
	if a.win != nil {
		a.win.SetScaleHotkeysEnabled(enabled)
	}
}

// SetWindowIcon sets the native window icon from a standard Go image.
func (a *App) SetWindowIcon(icon image.Image) error {
	if a.win == nil {
		return ErrWindowNotInitialized
	}
	return a.win.SetIcon(icon)
}

// Run runs the execution loop with the given UI; the window event loop blocks
// until the window is closed.
func (a *App) Run(ui *components.UI) error {
	if a.win == nil {
		return ErrWindowNotInitialized
	}
	if ui == nil {
		return ErrNilUI
	}
	a.win.attachUI(ui)
	a.win.setBeforeFrame(a.drainDispatch)
	err := a.win.Run()
	a.stopDispatch()
	return err
}

// RunScene runs a custom interface through the same SDL lifecycle as retained
// widget UIs. A scene can use the normalized event API and only drop down to
// the SDL renderer for application-specific drawing.
func (a *App) RunScene(scene Scene) error {
	return a.runScene(scene, nil)
}

// RunSceneWithUI runs a custom scene with a retained UI drawn above it. The UI
// receives input first using the scene context's UIInputPolicy.
func (a *App) RunSceneWithUI(scene Scene, ui *components.UI) error {
	return a.runScene(scene, ui)
}

func (a *App) runScene(scene Scene, ui *components.UI) error {
	if a.win == nil {
		return ErrWindowNotInitialized
	}
	if scene == nil {
		return ErrNilScene
	}
	ctx := &SceneContext{window: a.win}
	a.win.attachScene(scene, ctx)
	ctx.SetUI(ui)
	if initializer, ok := scene.(SceneInitializer); ok {
		if err := initializer.Init(ctx); err != nil {
			if closer, exists := scene.(SceneCloser); exists {
				closer.Close()
			}
			ctx.close()
			a.win.detachScene()
			return err
		}
	}
	a.win.setBeforeFrame(a.drainDispatch)
	runErr := a.win.Run()
	if closer, ok := scene.(SceneCloser); ok {
		closer.Close()
	}
	ctx.close()
	a.win.detachScene()
	a.stopDispatch()
	return runErr
}

// Destroy closes the window and frees resources.
func (a *App) Destroy() {
	a.stopDispatch()
	if a.win != nil {
		a.win.Destroy()
		a.win = nil
	}
}

// Window returns the window handle.
func (a *App) Window() *Window {
	return a.win
}
