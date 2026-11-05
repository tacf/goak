package sdl3

import (
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak"
)

// Clipboard is the SDL system clipboard. It satisfies goak.Clipboard.
type Clipboard struct{}

// Clipboard returns the system clipboard.
func (b *Backend) Clipboard() goak.Clipboard { return Clipboard{} }

// Get returns the clipboard text, or "" when empty or unavailable.
func (Clipboard) Get() string {
	s, err := sdl.GetClipboardText()
	if err != nil {
		return ""
	}
	return s
}

// Set replaces the clipboard text. Errors are ignored.
func (Clipboard) Set(s string) { _ = sdl.SetClipboardText(s) }
