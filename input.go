// Package input holds the neutral per-frame input state of the immediate-mode
// UI. It is pure Go and must not import any SDL package.
package goak

// Key identifies a keyboard key independent of the windowing library.
type Key uint8

// Keys known to the UI.
const (
	KeyNone Key = iota
	KeyTab
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
	KeyInsert
	KeyDelete
	KeyBackspace
	KeySpace
	KeyEnter
	KeyEscape
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	keyCount
)

// KeyCount is the number of Key values.
const KeyCount = int(keyCount)

// MouseButton identifies a mouse button.
type MouseButton uint8

// Mouse buttons.
const (
	MouseLeft MouseButton = iota
	MouseMiddle
	MouseRight
	mouseCount
)

// MouseButtonCount is the number of tracked mouse buttons.
const MouseButtonCount = int(mouseCount)

// Mod is a bit set of keyboard modifiers.
type Mod uint8

// Modifier bits.
const (
	ModCtrl Mod = 1 << iota
	ModShift
	ModAlt
	ModSuper
)

// Input is the input state for one frame. Positions are in logical units.
type Input struct {
	MousePos   [2]float32
	MouseDown  [MouseButtonCount]bool
	MousePress [MouseButtonCount]bool // went down this frame
	MouseRel   [MouseButtonCount]bool // went up this frame
	// MouseClicks holds the last click count per button (2 is a double click).
	MouseClicks [MouseButtonCount]uint8
	Wheel       [2]float32 // accumulated this frame
	KeyDown     [KeyCount]bool
	KeyPressed  [KeyCount]bool // down this frame, includes key repeat
	Text        string         // text typed this frame
	Mods        Mod
	DT          float32 // seconds since the previous frame
	Quit        bool    // a quit or close request was received
}

// BeginFrame clears the one-frame fields and keeps the held state.
func (in *Input) BeginFrame() {
	in.MousePress = [MouseButtonCount]bool{}
	in.MouseRel = [MouseButtonCount]bool{}
	in.MouseClicks = [MouseButtonCount]uint8{}
	in.Wheel = [2]float32{}
	in.KeyPressed = [KeyCount]bool{}
	in.Text = ""
}

// SetKey records a key transition. Repeats count as presses only.
func (in *Input) SetKey(k Key, down bool) {
	if k == KeyNone || int(k) >= KeyCount {
		return
	}
	if down {
		in.KeyPressed[k] = true
	}
	in.KeyDown[k] = down
}

// SetMouseButton records a button transition.
func (in *Input) SetMouseButton(b MouseButton, down bool, clicks uint8) {
	if int(b) >= MouseButtonCount {
		return
	}
	if down && !in.MouseDown[b] {
		in.MousePress[b] = true
		in.MouseClicks[b] = clicks
	}
	if !down && in.MouseDown[b] {
		in.MouseRel[b] = true
	}
	in.MouseDown[b] = down
}

// Has reports whether all bits of m are set.
func (in *Input) Has(m Mod) bool { return in.Mods&m == m }
