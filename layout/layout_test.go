package layout_test

import (
	"testing"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

func TestPublicLayoutConfiguration(t *testing.T) {
	ui := components.NewUI()
	root := ui.Root()
	root.SetDirection(layout.Row)
	root.SetGap(8)
	root.SetInsets(layout.Insets{Top: 2, Right: 4, Bottom: 4, Left: 3})

	first := root.CreatePanel(layout.StaticPx(20), layout.AutoSize())
	second := root.CreatePanel(layout.AutoSize().WithMin(30), layout.AutoSize())
	layout.Layout(root.Container(), 100, 40)

	if got, want := first.Bounds(), (layout.Rect{X: 3, Y: 2, W: 20, H: 34}); got != want {
		t.Fatalf("first bounds = %+v, want %+v", got, want)
	}
	if got, want := second.Bounds(), (layout.Rect{X: 31, Y: 2, W: 65, H: 34}); got != want {
		t.Fatalf("second bounds = %+v, want %+v", got, want)
	}
}

func TestComponentLayoutSettersNormalizeValues(t *testing.T) {
	ui := components.NewUI()
	root := ui.Root()
	root.SetDirection(layout.Direction(99))
	root.SetGap(-2)
	root.SetInsets(layout.Insets{Top: -1, Right: 2, Bottom: 3, Left: -4})

	container := root.Container()
	if container.Direction != layout.Column {
		t.Fatalf("direction = %v, want Column", container.Direction)
	}
	if container.Gap != 0 {
		t.Fatalf("gap = %v, want 0", container.Gap)
	}
	if got, want := container.Insets, (layout.Insets{Right: 2, Bottom: 3}); got != want {
		t.Fatalf("insets = %+v, want %+v", got, want)
	}

	panel := root.CreatePanel(layout.AutoSize(), layout.AutoSize())
	panel.SetPadding(7)
	if panel.Container().Padding != 7 || panel.Container().Insets != (layout.Insets{}) {
		t.Fatalf("SetPadding did not select uniform padding: %+v", panel.Container())
	}
	panel.SetInsets(layout.UniformInsets(5))
	if panel.Container().Padding != 0 || panel.Container().Insets != layout.UniformInsets(5) {
		t.Fatalf("SetInsets did not replace uniform padding: %+v", panel.Container())
	}
}
