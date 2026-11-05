package goak

import "github.com/tacf/goak/internal/draw"

// Renderer draws a frame and owns the textures it samples. Texture ids are
// opaque to goak and 0 means no texture.
type Renderer = draw.Renderer

// Backend is a complete platform: window, input, clipboard and renderer.
// Implement it to run goak on something other than SDL3. See backend/sdl3.
type Backend interface {
	Renderer

	// NewFrame polls events into in and reports false once the application
	// should quit. It calls in.BeginFrame and fills in.DT.
	NewFrame(in *Input) bool
	// DisplaySize is the drawable area in logical units.
	DisplaySize() Vec2
	// Scale is pixels per logical unit.
	Scale() float32
	// ScaleChanged reports a new Scale once after it changes.
	ScaleChanged() (float32, bool)
	// Clipboard is the system text clipboard.
	Clipboard() Clipboard
	// Clear fills the target with c before drawing.
	Clear(c Color) error
	// Present shows the finished frame.
	Present() error
	// Close releases the window and all backend resources.
	Close()
}
