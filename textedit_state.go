package goak

import "github.com/tacf/goak/internal/textedit"

// textEdit is the editing state of one InputText.
type textEdit struct {
	textedit.Buffer
	scrollX    float32
	scrollY    float32
	noDrag     bool // ignore mouse drag selection until release
	blink      float32
	wasFocused bool
	spans      [][2]int

	// Caches keyed on rev, bumped by every text change.
	rev       int
	cacheRev  int
	cachePW   bool
	cacheMult bool
	cacheOK   bool
	lines     []string
	maxW      float32
	maxWOK    bool
}
