package components_test

import (
	"math"
	"testing"

	"github.com/tacf/goak/components"
)

func TestUIFontDataIsDefensivelyCopiedAndVersioned(t *testing.T) {
	ui := components.NewUI()
	initialRevision := ui.FontRevision()
	if data := ui.FontData(); data != nil {
		t.Fatalf("default font data = %v, want nil", data)
	}

	source := []byte{1, 2, 3}
	ui.SetFontData(source)
	if got := ui.FontRevision(); got != initialRevision+1 {
		t.Fatalf("font revision after custom data = %d, want %d", got, initialRevision+1)
	}
	source[0] = 9
	got := ui.FontData()
	if len(got) != 3 || got[0] != 1 {
		t.Fatalf("stored font changed with source slice: %v", got)
	}

	got[1] = 9
	if next := ui.FontData(); next[1] != 2 {
		t.Fatalf("stored font changed with returned slice: %v", next)
	}

	unchangedRevision := ui.FontRevision()
	ui.SetFontData([]byte{1, 2, 3})
	if got := ui.FontRevision(); got != unchangedRevision {
		t.Fatalf("equal font data advanced revision to %d", got)
	}

	ui.SetFontData([]byte{1, 2, 4})
	if got := ui.FontRevision(); got != unchangedRevision+1 {
		t.Fatalf("changed font data revision = %d, want %d", got, unchangedRevision+1)
	}

	ui.SetFontData([]byte{})
	if data := ui.FontData(); data != nil {
		t.Fatalf("empty font data did not reset default: %v", data)
	}
	resetRevision := ui.FontRevision()
	ui.SetFontData(nil)
	if got := ui.FontRevision(); got != resetRevision {
		t.Fatalf("repeated default reset advanced revision to %d", got)
	}
}

func TestUIFontRevisionTracksEffectiveConfiguration(t *testing.T) {
	ui := components.NewUI()
	revision := ui.FontRevision()

	ui.SetFontSize(18)
	if got := ui.FontRevision(); got != revision+1 {
		t.Fatalf("font-size revision = %d, want %d", got, revision+1)
	}
	revision = ui.FontRevision()

	ui.SetFontSize(18)
	ui.SetFontSize(0)
	ui.SetFontSize(math.NaN())
	if got := ui.FontRevision(); got != revision {
		t.Fatalf("unchanged or invalid font sizes advanced revision to %d", got)
	}

	ui.SetFontData([]byte{4, 5, 6})
	if got := ui.FontRevision(); got != revision+1 {
		t.Fatalf("font-data revision = %d, want %d", got, revision+1)
	}
}

func TestNilUIFontConfigurationIsSafe(t *testing.T) {
	var ui *components.UI
	ui.SetFontData([]byte{1})
	if ui.FontData() != nil || ui.FontRevision() != 0 {
		t.Fatal("nil UI reported font configuration")
	}
}
