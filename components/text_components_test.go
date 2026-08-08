package components

import (
	"image"
	"math"
	"testing"
	"unicode/utf8"

	"github.com/tacf/goak/layout"

	"github.com/Zyko0/go-sdl3/sdl"
)

func TestTextInputEditsUnicodeAndEmitsTypedEvents(t *testing.T) {
	input := NewTextInput(layout.AutoSize(), layout.AutoSize(), "hé")
	var changes []TextInputChangedEvent
	var submitted TextInputSubmittedEvent
	input.SetOnChanged(func(event TextInputChangedEvent) { changes = append(changes, event) })
	input.SetOnSubmitted(func(event TextInputSubmittedEvent) { submitted = event })

	input.HandleKey(sdl.K_LEFT, 0)
	input.HandleTextInput("λ")
	input.HandleKey(sdl.K_BACKSPACE, 0)
	input.HandleKey(sdl.K_END, 0)
	input.HandleKey(sdl.K_RETURN, 0)

	if got, want := input.Text(), "hé"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	if input.Cursor() != utf8.RuneCountInString(input.Text()) {
		t.Fatalf("cursor = %d, want rune length", input.Cursor())
	}
	if len(changes) != 2 || changes[0].Previous != "hé" || changes[0].Text != "hλé" {
		t.Fatalf("unexpected change events: %#v", changes)
	}
	if submitted.TextInput != input || submitted.Text != "hé" {
		t.Fatalf("unexpected submit event: %#v", submitted)
	}
}

func TestTextInputNormalizesNewlines(t *testing.T) {
	input := NewTextInput(layout.AutoSize(), layout.AutoSize(), "first\r\nsecond")
	if got, want := input.Text(), "first second"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	input.HandleTextInput("\nthird")
	if got, want := input.Text(), "first second third"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}

func TestTextAreaEditingOptionsAndScrolling(t *testing.T) {
	area := NewTextArea(layout.AutoSize(), layout.AutoSize(), "one\ntwo")
	if !area.Wrap() || !area.LineNumbers() {
		t.Fatal("text area should wrap and show line numbers by default")
	}
	area.SetCursor(3)
	area.HandleKey(sdl.K_RETURN, 0)
	area.HandleTextInput("middle")
	if got, want := area.Text(), "one\nmiddle\ntwo"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}

	area.SetWrap(false)
	area.SetLineNumbers(false)
	area.SetScrollOffsets(12, 34)
	x, y := area.ScrollOffsets()
	if area.Wrap() || area.LineNumbers() || x != 12 || y != 34 {
		t.Fatalf("unexpected options/offsets: wrap=%v numbers=%v scroll=(%v,%v)", area.Wrap(), area.LineNumbers(), x, y)
	}
	area.SetWrap(true)
	x, _ = area.ScrollOffsets()
	if x != 0 {
		t.Fatalf("horizontal scroll = %v with wrapping enabled, want 0", x)
	}
}

func TestVisualLinesWrapAtWordsAndKeepLogicalNumbers(t *testing.T) {
	text := []rune("one two\nthree")
	lines := visualLinesMeasured(text, true, 5, func(value string) float64 {
		return float64(len([]rune(value)))
	})
	if len(lines) != 3 {
		t.Fatalf("visual line count = %d, want 3: %#v", len(lines), lines)
	}
	if got := string(text[lines[0].start:lines[0].end]); got != "one " {
		t.Fatalf("first visual line = %q, want %q", got, "one ")
	}
	if got := string(text[lines[1].start:lines[1].end]); got != "two" {
		t.Fatalf("second visual line = %q, want %q", got, "two")
	}
	if lines[0].logicalLine != 0 || lines[0].continuation ||
		lines[1].logicalLine != 0 || !lines[1].continuation ||
		lines[2].logicalLine != 1 || lines[2].continuation {
		t.Fatalf("unexpected logical line metadata: %#v", lines)
	}
}

func TestImageFitRects(t *testing.T) {
	bounds := layout.Rect{X: 10, Y: 20, W: 100, H: 100}
	source, destination := imageRects(bounds, 200, 100, ImageFitContain)
	if source != nil || destination.X != 10 || destination.Y != 45 ||
		destination.W != 100 || destination.H != 50 {
		t.Fatalf("unexpected contain rects: source=%#v destination=%#v", source, destination)
	}

	source, destination = imageRects(bounds, 200, 100, ImageFitCover)
	if source == nil || destination.X != 10 || destination.Y != 20 ||
		destination.W != 100 || destination.H != 100 ||
		math.Abs(float64(source.X)-50) > 0.001 || source.W != 100 || source.H != 100 {
		t.Fatalf("unexpected cover rects: source=%#v destination=%#v", source, destination)
	}
}

func TestNewComponentsRegisterInUIAndLayout(t *testing.T) {
	ui := NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(200), layout.StaticPx(200))
	picture := panel.CreateImage(layout.StaticPx(20), layout.StaticPx(30), image.NewRGBA(image.Rect(0, 0, 2, 2)))
	input := panel.CreateTextInput(layout.StaticPx(100), layout.StaticPx(25), "")
	area := panel.CreateTextArea(layout.StaticPx(100), layout.StaticPx(80), "")
	layout.Layout(ui.Root().Container(), 300, 300)

	if len(ui.Images()) != 1 || ui.Images()[0] != picture ||
		len(ui.TextInputs()) != 1 || ui.TextInputs()[0] != input ||
		len(ui.TextAreas()) != 1 || ui.TextAreas()[0] != area {
		t.Fatal("new components were not registered with UI")
	}
	if picture.Bounds().H != 30 || input.Bounds().Y != picture.Bounds().Y+picture.Bounds().H ||
		area.Bounds().Y != input.Bounds().Y+input.Bounds().H {
		t.Fatalf("unexpected component layout: image=%#v input=%#v area=%#v", picture.Bounds(), input.Bounds(), area.Bounds())
	}
}
