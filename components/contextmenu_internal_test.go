package components

import (
	"testing"

	"github.com/tacf/goak/layout"
)

func TestContextMenuSetItemsCopiesInput(t *testing.T) {
	items := []ContextMenuItem{NewContextMenuAction("Open", nil)}
	menu := NewContextMenu(items)
	items[0].Label = "Changed"
	if got := menu.Items()[0].Label; got != "Open" {
		t.Fatalf("menu label = %q, want Open", got)
	}

	menu.SetItems([]ContextMenuItem{
		NewContextMenuActionWithHint("Save", "Ctrl+S", nil),
		NewContextMenuSeparator(),
		NewDisabledContextMenuActionWithHint("Delete", "Del"),
	})
	got := menu.Items()
	if len(got) != 3 || got[0].Hint != "Ctrl+S" || got[2].Hint != "Del" {
		t.Fatalf("unexpected replacement items: %#v", got)
	}
}

func TestContextMenuAutomaticOpeningIsConfigurable(t *testing.T) {
	menu := NewContextMenu(nil)
	if !menu.AutoOpen() {
		t.Fatal("context menu should auto-open by default")
	}
	menu.SetAutoOpen(false)
	if menu.AutoOpen() {
		t.Fatal("context menu ignored SetAutoOpen(false)")
	}
}

func TestContextMenuConfigurationAndViewportClamping(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", nil),
		NewContextMenuSeparator(),
		NewContextMenuAction("Second", nil),
	})
	menu.SetItemHeight(30)
	menu.SetSeparatorHeight(10)
	menu.SetPadding(12, 5)
	menu.SetMinWidth(120)
	menu.OpenAt(190, 95, layout.Rect{X: 10, Y: 5, W: 200, H: 100})

	bounds := menu.Bounds()
	want := layout.Rect{X: 90, Y: 25, W: 120, H: 80}
	if bounds != want {
		t.Fatalf("bounds = %+v, want %+v", bounds, want)
	}
	if got := menu.HitTest(bounds.X+1, bounds.Y+6); got != 0 {
		t.Fatalf("first row hit = %d, want 0", got)
	}
	if got := menu.HitTest(bounds.X+1, bounds.Y+40); got != -1 {
		t.Fatalf("separator hit = %d, want -1", got)
	}
	if got := menu.HitTest(bounds.X+1, bounds.Y+50); got != 1 {
		t.Fatalf("second row hit = %d, want 1", got)
	}
}

func TestContextMenuMeasuredWidthIncludesHints(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuActionWithHint("Open", "Ctrl+O", nil),
	})
	menu.SetMinWidth(0)
	menu.SetPadding(10, 0)

	width := menu.measureWidth(func(text string) float64 {
		return float64(len(text) * 10)
	})
	if width != 144 {
		t.Fatalf("measured width = %v, want 144", width)
	}

	menu.OpenAt(190, 20, layout.Rect{X: 10, Y: 0, W: 200, H: 100})
	menu.measuredW = width
	menu.place()
	if bounds := menu.Bounds(); bounds.X != 66 || bounds.W != 144 {
		t.Fatalf("measured bounds = %+v, want x=66 width=144", bounds)
	}
}

func TestContextMenuCapsHeightAndScrollsHitTesting(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", nil),
		NewContextMenuAction("Second", nil),
		NewContextMenuAction("Third", nil),
		NewContextMenuAction("Fourth", nil),
	})
	menu.SetItemHeight(20)
	menu.SetPadding(4, 5)
	menu.OpenAt(10, 90, layout.Rect{X: 0, Y: 0, W: 100, H: 50})

	bounds := menu.Bounds()
	if bounds.H != 50 || bounds.Y != 0 {
		t.Fatalf("bounds = %+v, want height 50 contained at y=0", bounds)
	}
	if got := menu.HitTest(bounds.X+5, bounds.Y+10); got != 0 {
		t.Fatalf("initial hit = %d, want 0", got)
	}

	menu.ScrollWheel(-1)
	if menu.scrollY != 20 {
		t.Fatalf("scroll offset = %v, want 20", menu.scrollY)
	}
	if got := menu.HitTest(bounds.X+5, bounds.Y+10); got != 1 {
		t.Fatalf("scrolled hit = %d, want 1", got)
	}

	menu.ScrollWheel(-100)
	if menu.scrollY != 40 {
		t.Fatalf("clamped scroll offset = %v, want 40", menu.scrollY)
	}
	if got := menu.HitTest(bounds.X+5, bounds.Y+30); got != 3 {
		t.Fatalf("last visible hit = %d, want 3", got)
	}

	menu.ScrollWheel(100)
	if menu.scrollY != 0 {
		t.Fatalf("reverse-clamped scroll offset = %v, want 0", menu.scrollY)
	}
}

func TestContextMenuScrollHitTestingAccountsForSeparators(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", nil),
		NewContextMenuSeparator(),
		NewContextMenuAction("Second", nil),
	})
	menu.SetItemHeight(20)
	menu.SetSeparatorHeight(10)
	menu.SetPadding(0, 0)
	menu.OpenAt(0, 0, layout.Rect{W: 100, H: 25})
	menu.ScrollWheel(-1)

	if got := menu.HitTest(5, 5); got != -1 {
		t.Fatalf("separator hit = %d, want -1", got)
	}
	if got := menu.HitTest(5, 15); got != 1 {
		t.Fatalf("scrolled second-row hit = %d, want 1", got)
	}
}

func TestContextMenuKeyboardNavigationRevealsHoveredItem(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", nil),
		NewContextMenuAction("Second", nil),
		NewContextMenuAction("Third", nil),
		NewContextMenuAction("Fourth", nil),
	})
	menu.SetItemHeight(20)
	menu.SetPadding(0, 0)
	menu.OpenAt(0, 0, layout.Rect{W: 100, H: 40})

	menu.HandleKey("end")
	if menu.HoveredIndex() != 3 || menu.scrollY != 40 {
		t.Fatalf("end: hover=%d scroll=%v, want hover=3 scroll=40", menu.HoveredIndex(), menu.scrollY)
	}
	if got := menu.HitTest(5, 30); got != 3 {
		t.Fatalf("revealed end hit = %d, want 3", got)
	}

	menu.HandleKey("home")
	if menu.HoveredIndex() != 0 || menu.scrollY != 0 {
		t.Fatalf("home: hover=%d scroll=%v, want hover=0 scroll=0", menu.HoveredIndex(), menu.scrollY)
	}

	menu.HandleKey("down")
	menu.HandleKey("down")
	if menu.HoveredIndex() != 2 || menu.scrollY != 20 {
		t.Fatalf("down navigation: hover=%d scroll=%v, want hover=2 scroll=20", menu.HoveredIndex(), menu.scrollY)
	}
}

func TestContextMenuScrollingDoesNotChangeFittingMenu(t *testing.T) {
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", nil),
		NewContextMenuAction("Second", nil),
	})
	menu.SetPadding(0, 0)
	menu.OpenAt(0, 0, layout.Rect{W: 200, H: 100})
	want := menu.Bounds()

	menu.ScrollWheel(-10)
	if menu.scrollY != 0 || menu.Bounds() != want {
		t.Fatalf("fitting menu changed: scroll=%v bounds=%+v, want %+v", menu.scrollY, menu.Bounds(), want)
	}
}

func TestContextMenuKeyboardNavigationSkipsUnavailableItems(t *testing.T) {
	var runs []string
	menu := NewContextMenu([]ContextMenuItem{
		NewDisabledContextMenuAction("Disabled"),
		NewContextMenuSeparator(),
		NewContextMenuAction("Open", func() { runs = append(runs, "open") }),
		NewContextMenuAction("Save", func() { runs = append(runs, "save") }),
	})
	var event ContextMenuActionEvent
	menu.SetOnAction(func(got ContextMenuActionEvent) { event = got })
	menu.Open(0, 0)

	menu.SetHovered(0)
	if got := menu.HoveredIndex(); got != -1 {
		t.Fatalf("disabled hover = %d, want -1", got)
	}
	if !menu.HandleKey("down") || menu.HoveredIndex() != 1 {
		t.Fatalf("first down hover = %d, want 1", menu.HoveredIndex())
	}
	menu.HandleKey("down")
	if menu.HoveredIndex() != 2 {
		t.Fatalf("second down hover = %d, want 2", menu.HoveredIndex())
	}
	menu.HandleKey("down")
	if menu.HoveredIndex() != 1 {
		t.Fatalf("wrapped hover = %d, want 1", menu.HoveredIndex())
	}
	menu.HandleKey("up")
	if menu.HoveredIndex() != 2 {
		t.Fatalf("up hover = %d, want 2", menu.HoveredIndex())
	}
	menu.HandleKey("home")
	if menu.HoveredIndex() != 1 {
		t.Fatalf("home hover = %d, want 1", menu.HoveredIndex())
	}
	menu.HandleKey("end")
	if menu.HoveredIndex() != 2 {
		t.Fatalf("end hover = %d, want 2", menu.HoveredIndex())
	}
	if menu.HandleKey("left") {
		t.Fatal("unrecognized key was consumed")
	}
	if !menu.HandleKey("return") {
		t.Fatal("return was not consumed")
	}
	if menu.IsOpen() || len(runs) != 1 || runs[0] != "save" || event.Index != 2 {
		t.Fatalf("activation: open=%v runs=%v event=%#v", menu.IsOpen(), runs, event)
	}
	if menu.HandleKey("down") {
		t.Fatal("closed menu consumed a key")
	}
}

func TestContextMenuInvalidClickDoesNotActivateFirstItem(t *testing.T) {
	var runs int
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("First", func() { runs++ }),
	})
	menu.Open(0, 0)
	menu.Click(10)
	if runs != 0 {
		t.Fatalf("invalid click ran %d actions", runs)
	}
}
