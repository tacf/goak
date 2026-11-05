package sdl3

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak"
)

func TestTranslateKey(t *testing.T) {
	cases := map[sdl.Keycode]goak.Key{
		sdl.K_A: goak.KeyA, sdl.K_Z: goak.KeyZ, sdl.K_5: goak.Key5,
		sdl.K_F1: goak.KeyF1, sdl.K_F12: goak.KeyF12, sdl.K_RETURN: goak.KeyEnter,
		sdl.K_LEFT: goak.KeyLeft, sdl.K_ESCAPE: goak.KeyEscape,
		sdl.K_HOME: goak.KeyHome, sdl.K_END: goak.KeyEnd, sdl.K_DELETE: goak.KeyDelete,
		sdl.K_BACKSPACE: goak.KeyBackspace, sdl.K_C: goak.KeyC, sdl.K_V: goak.KeyV, sdl.K_X: goak.KeyX,
	}
	for in, want := range cases {
		if got := translateKey(in); got != want {
			t.Errorf("key %v = %v, want %v", in, got, want)
		}
	}
	if translateKey(sdl.Keycode(0x7fffffff)) != goak.KeyNone {
		t.Error("unknown key must map to KeyNone")
	}
}

func TestTranslateMods(t *testing.T) {
	m := translateMods(sdl.KMOD_LCTRL | sdl.KMOD_RSHIFT)
	if m != goak.ModCtrl|goak.ModShift {
		t.Fatalf("got %v", m)
	}
}
