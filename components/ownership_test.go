package components

import (
	"errors"
	"testing"

	"github.com/tacf/goak/layout"
)

func TestDetachedPanelSubtreeMountsRecursively(t *testing.T) {
	panel := NewPanel(layout.AutoSize(), layout.AutoSize())
	button := panel.CreateButton(layout.AutoSize(), layout.StaticPx(24), "Run")
	nested := panel.CreatePanel(layout.AutoSize(), layout.AutoSize())
	label := nested.CreateLabel(layout.AutoSize(), layout.StaticPx(20), "Status")
	menu := NewContextMenu(nil)
	if err := nested.AddContextMenu(menu); err != nil {
		t.Fatal(err)
	}

	ui := NewUI()
	if err := ui.Root().AddPanel(panel); err != nil {
		t.Fatal(err)
	}
	for _, component := range []Component{panel, button, nested, label} {
		if !ui.Contains(component) {
			t.Fatalf("mounted tree does not contain %T", component)
		}
	}
	if ui.ContextMenuCount() != 1 || ui.ContextMenu(0) != menu {
		t.Fatal("detached subtree context menu was not mounted recursively")
	}
}

func TestComponentOwnershipRejectsDuplicatesAndSupportsExplicitReparenting(t *testing.T) {
	ui := NewUI()
	first := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	second := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	button := NewButton(layout.AutoSize(), layout.AutoSize(), "Move")
	if err := first.Add(button); err != nil {
		t.Fatal(err)
	}
	if err := second.Add(button); !errors.Is(err, ErrComponentAttached) {
		t.Fatalf("duplicate attachment error = %v, want %v", err, ErrComponentAttached)
	}
	if !button.Remove() {
		t.Fatal("attached button did not remove")
	}
	if err := second.Add(button); err != nil {
		t.Fatalf("reattach after Remove: %v", err)
	}
	if button.Parent() != second || !ui.Contains(button) {
		t.Fatal("button did not acquire its new parent")
	}
}

func TestComponentTreeOwnsOrderVisibilityAndSnapshots(t *testing.T) {
	ui := NewUI()
	panel := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	high := panel.CreateButton(layout.AutoSize(), layout.AutoSize(), "High")
	low := panel.CreateLabel(layout.AutoSize(), layout.AutoSize(), "Low")
	high.SetZIndex(10)

	if got := ui.Component(1); got != low {
		t.Fatalf("component at draw index 1 = %T, want low-z label", got)
	}
	if got := ui.Component(2); got != high {
		t.Fatalf("component at draw index 2 = %T, want high-z button", got)
	}
	panel.SetVisible(false)
	if ui.ComponentVisible(high) {
		t.Fatal("child of hidden panel remained effectively visible")
	}
	panel.SetVisible(true)
	panel.SetEnabled(false)
	if ui.ComponentEnabled(high) {
		t.Fatal("child of disabled panel remained effectively enabled")
	}

	snapshot := ui.Buttons()
	snapshot[0] = nil
	if !ui.Contains(high) || ui.Buttons()[0] != high {
		t.Fatal("mutating a typed accessor changed UI ownership")
	}
}

func TestUILayoutRunsOnlyWhenDirty(t *testing.T) {
	ui := NewUI()
	panel := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	if !ui.Layout(100, 80) || ui.Layout(100, 80) {
		t.Fatal("layout cache did not distinguish dirty and unchanged passes")
	}
	panel.SetGap(4)
	if !ui.Layout(100, 80) {
		t.Fatal("layout setter did not invalidate layout")
	}
	panel.SetSize(layout.StaticPx(40), layout.StaticPx(30))
	if !ui.Layout(100, 80) {
		t.Fatal("component size did not invalidate layout")
	}
}

func TestNilComponentAttachmentReturnsError(t *testing.T) {
	ui := NewUI()
	var button *Button
	if err := ui.Root().Add(button); !errors.Is(err, ErrNilComponent) {
		t.Fatalf("typed nil attachment error = %v, want %v", err, ErrNilComponent)
	}
}
