package layout

import "testing"

func TestPaddingInsetsChildLayout(t *testing.T) {
	child := NewContainer(PercentOf(50), StaticPx(20))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 10

	Layout(root, 100, 80)

	want := Rect{X: 10, Y: 10, W: 40, H: 20}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}

func TestPaddingConstrainsCenteredContent(t *testing.T) {
	child := NewContainer(StaticPx(20), StaticPx(10))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 10
	root.HorizontalAlign = AlignCenter
	root.VerticalAlign = AlignCenter

	Layout(root, 100, 60)

	want := Rect{X: 40, Y: 25, W: 20, H: 10}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}

func TestPaddingIsClampedToContainer(t *testing.T) {
	child := NewContainer(PercentOf(100), PercentOf(100))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 100

	Layout(root, 40, 20)

	want := Rect{X: 20, Y: 10, W: 0, H: 0}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}
