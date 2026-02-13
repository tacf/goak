package goak

import (
	"goak/internal/goak/components"
)

// App is the application API. Create with NewApp, call InitWindow, then Run(ui).
type App struct {
	win *Window
}

// NewApp returns a new App. Call InitWindow before Run.
func NewApp() *App {
	return &App{}
}

// InitWindow creates and configures the window with the given title and size.
// Must be called before Run.
func (a *App) InitWindow(title string, width, height int) {
	a.win = InitWindow(title, width, height)
}

// InitWindowWithConfig creates and configures the window with explicit options.
func (a *App) InitWindowWithConfig(cfg Config) {
	a.win = New(cfg)
}

// SetAutoDPI toggles automatic HiDPI scaling on the app window.
func (a *App) SetAutoDPI(enabled bool) {
	if a.win != nil {
		a.win.SetAutoDPI(enabled)
	}
}

// SetWindowScale sets the runtime window scale multiplier.
func (a *App) SetWindowScale(scale float64) {
	if a.win != nil {
		a.win.SetWindowScale(scale)
	}
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

// Run runs the execution loop with the given UI; the window event loop blocks
// until the window is closed.
func (a *App) Run(ui *components.UI) {
	if a.win == nil || ui == nil {
		return
	}
	a.win.attachUI(ui)
	a.win.Run()
}

// Destroy closes the window and frees resources.
func (a *App) Destroy() {
	if a.win != nil {
		a.win.Destroy()
	}
}

// Window returns the window handle.
func (a *App) Window() *Window {
	return a.win
}
