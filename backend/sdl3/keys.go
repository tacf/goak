package sdl3

import (
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak"
)

var namedKeys = map[sdl.Keycode]goak.Key{
	sdl.K_TAB:       goak.KeyTab,
	sdl.K_LEFT:      goak.KeyLeft,
	sdl.K_RIGHT:     goak.KeyRight,
	sdl.K_UP:        goak.KeyUp,
	sdl.K_DOWN:      goak.KeyDown,
	sdl.K_PAGEUP:    goak.KeyPageUp,
	sdl.K_PAGEDOWN:  goak.KeyPageDown,
	sdl.K_HOME:      goak.KeyHome,
	sdl.K_END:       goak.KeyEnd,
	sdl.K_INSERT:    goak.KeyInsert,
	sdl.K_DELETE:    goak.KeyDelete,
	sdl.K_BACKSPACE: goak.KeyBackspace,
	sdl.K_SPACE:     goak.KeySpace,
	sdl.K_RETURN:    goak.KeyEnter,
	sdl.K_KP_ENTER:  goak.KeyEnter,
	sdl.K_ESCAPE:    goak.KeyEscape,
}

// translateKey maps an SDL keycode to the neutral key enum.
func translateKey(k sdl.Keycode) goak.Key {
	switch {
	case k >= sdl.K_A && k <= sdl.K_Z:
		return goak.KeyA + goak.Key(k-sdl.K_A)
	case k >= sdl.K_0 && k <= sdl.K_9:
		return goak.Key0 + goak.Key(k-sdl.K_0)
	case k >= sdl.K_F1 && k <= sdl.K_F12:
		return goak.KeyF1 + goak.Key(k-sdl.K_F1)
	}
	return namedKeys[k]
}

// translateMods maps SDL modifier flags to neutral modifier bits.
func translateMods(m sdl.Keymod) goak.Mod {
	var out goak.Mod
	if m&sdl.KMOD_CTRL != 0 {
		out |= goak.ModCtrl
	}
	if m&sdl.KMOD_SHIFT != 0 {
		out |= goak.ModShift
	}
	if m&sdl.KMOD_ALT != 0 {
		out |= goak.ModAlt
	}
	if m&sdl.KMOD_GUI != 0 {
		out |= goak.ModSuper
	}
	return out
}

// translateButton maps an SDL button index to the neutral enum.
func translateButton(b uint8) (goak.MouseButton, bool) {
	switch b {
	case uint8(sdl.BUTTON_LEFT):
		return goak.MouseLeft, true
	case uint8(sdl.BUTTON_MIDDLE):
		return goak.MouseMiddle, true
	case uint8(sdl.BUTTON_RIGHT):
		return goak.MouseRight, true
	}
	return 0, false
}
