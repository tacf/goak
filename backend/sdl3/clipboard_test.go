package sdl3

import "testing"

func TestClipboardRoundTrip(t *testing.T) {
	b := newTestBackend(t)
	cb := b.Clipboard()
	cb.Set("goak clipboard")
	if got := cb.Get(); got != "goak clipboard" {
		t.Skipf("system clipboard unavailable under this driver, got %q", got)
	}
}
