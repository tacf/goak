package draw

import (
	"testing"
	"time"

	"github.com/tacf/goak/colors"
)

func hashData(x float32) *DrawData {
	dl := NewDrawList()
	dl.AddRectFilled(Vec2{X: x, Y: 1}, Vec2{X: x + 10, Y: 11}, colors.RGB(1, 2, 3), 0)
	return &DrawData{Lists: []*DrawList{dl}, DisplaySize: Vec2{X: 100, Y: 100}, Scale: Vec2{X: 1, Y: 1}}
}

func TestHashEqualAndDiffers(t *testing.T) {
	if hashData(0).Hash() != hashData(0).Hash() {
		t.Fatal("equal data hashed differently")
	}
	if hashData(0).Hash() == hashData(1).Hash() {
		t.Fatal("moved rect hashed equal")
	}
	a, b := hashData(0), hashData(0)
	b.Scale = Vec2{X: 2, Y: 2}
	if a.Hash() == b.Hash() {
		t.Fatal("scale ignored")
	}
}

func TestFrameSkipper(t *testing.T) {
	var s FrameSkipper
	if !s.Changed(hashData(0)) {
		t.Fatal("first frame must draw")
	}
	if s.Changed(hashData(0)) {
		t.Fatal("static frame must skip")
	}
	if !s.Changed(hashData(5)) {
		t.Fatal("changed frame must draw")
	}
	s.Interval = time.Millisecond
	time.Sleep(3 * time.Millisecond)
	if !s.Changed(hashData(5)) {
		t.Fatal("interval must force a redraw")
	}
}
