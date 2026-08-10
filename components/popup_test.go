package components

import (
	"math"
	"testing"

	"github.com/tacf/goak/layout"
)

func TestDropdownClampsScrollsAndNavigatesInViewport(t *testing.T) {
	dropdown := NewDropdown(layout.StaticPx(80), layout.StaticPx(20), "Pick", []DropdownOption{
		{Label: "One"}, {Label: "Two"}, {Label: "Three"}, {Label: "Four"},
	})
	dropdown.Container().Bounds = layout.Rect{X: 70, Y: 35, W: 80, H: 20}
	if err := dropdown.SetItemHeight(20); err != nil {
		t.Fatal(err)
	}
	dropdown.Place(layout.Rect{W: 100, H: 60})
	dropdown.Open()

	bounds := dropdown.ListBounds()
	if bounds.X != 20 || bounds.Y != 0 || bounds.W != 80 || bounds.H != 35 {
		t.Fatalf("dropdown bounds = %+v, want x=20 y=0 w=80 h=35", bounds)
	}
	if !dropdown.HandleKey("end") || dropdown.HoveredIndex() != 3 {
		t.Fatalf("keyboard hover = %d, want 3", dropdown.HoveredIndex())
	}
	if dropdown.scrollY <= 0 {
		t.Fatal("keyboard navigation did not reveal the last option")
	}
	if got := dropdown.HitTestList(bounds.X+2, bounds.Y+2); got < 0 {
		t.Fatal("scrolled dropdown did not map visible coordinates to an option")
	}
	if !dropdown.HandleKey("return") || dropdown.SelectedIndex() != 3 || dropdown.IsOpen() {
		t.Fatal("Return did not select and close the dropdown")
	}
}

func TestPopupGeometryRejectsInvalidLengths(t *testing.T) {
	dropdown := NewDropdown(layout.AutoSize(), layout.AutoSize(), "", nil)
	if err := dropdown.SetItemHeight(0); err == nil {
		t.Fatal("dropdown accepted zero item height")
	}
	menu := NewContextMenu(nil)
	if err := menu.SetItemHeight(-1); err == nil {
		t.Fatal("context menu accepted a negative item height")
	}
	if err := menu.SetPadding(0, -1); err == nil {
		t.Fatal("context menu accepted negative padding")
	}
}

func TestMenuBarUsesFontMeasurementsAndConstrainsSubmenu(t *testing.T) {
	menu := NewMenuBar(layout.StaticPx(24), MenuBarWidthAuto)
	menu.AddItem("File", nil).AddSubItem("Very wide action", nil)
	menu.Container().Bounds = layout.Rect{W: 132, H: 24}
	menu.syncWidth(nil, func(text string) float64 {
		if text == "Very wide action" {
			return 180
		}
		return 100
	})

	if got := menu.Container().Width.Value; got != 132 {
		t.Fatalf("measured menu width = %v, want 132", got)
	}
	menu.Place(layout.Rect{W: 160, H: 70})
	if !menu.OnMouseDown(10, 10) {
		t.Fatal("top-level menu did not open")
	}
	if bounds := menu.OpenSubMenuBounds(); bounds.W != 160 || bounds.X != 0 {
		t.Fatalf("submenu bounds = %+v, want viewport-clamped width 160", bounds)
	}
	if !menu.HandleKey("end") || menu.HoverSubIndex() != 0 {
		t.Fatalf("submenu keyboard selection = %d, want 0", menu.HoverSubIndex())
	}
}

func TestMenuItemBuilderRemainsAttachedAfterMoreItemsAreAdded(t *testing.T) {
	menu := NewMenuBar(layout.StaticPx(24), MenuBarWidthAuto)
	file := menu.AddItem("File", nil)
	for index := 0; index < 32; index++ {
		menu.AddItem("Other", nil)
	}
	file.AddSubItem("Open", nil)

	items := menu.Items()
	if got := len(items[0].SubItems()); got != 1 {
		t.Fatalf("first item's submenu length = %d, want 1", got)
	}

	replacement := NewMenuItem("Edit", nil)
	replacement.AddSubItem("Undo", nil)
	menu.SetItems([]MenuItem{replacement})
	if got := menu.Items(); len(got) != 1 || got[0].Label() != "Edit" || len(got[0].SubItems()) != 1 {
		t.Fatalf("SetItems snapshot = %#v", got)
	}
}

func TestMenuBarClampsAndScrollsTallSubmenu(t *testing.T) {
	menu := NewMenuBar(layout.StaticPx(20), MenuBarWidthAuto)
	file := menu.AddItem("File", nil)
	for _, label := range []string{"One", "Two", "Three", "Four", "Five"} {
		file.AddSubItem(label, nil)
	}
	menu.Container().Bounds = layout.Rect{X: 0, Y: 30, W: 80, H: 20}
	menu.SyncWidth(nil)
	menu.Place(layout.Rect{W: 100, H: 60})
	if !menu.OnMouseDown(10, 35) {
		t.Fatal("top-level menu did not open")
	}
	if bounds := menu.OpenSubMenuBounds(); bounds.Y != 0 || bounds.H != 30 {
		t.Fatalf("submenu bounds = %+v, want upward-clamped height 30", bounds)
	}
	if !menu.HandleKey("end") || menu.HoverSubIndex() != 4 || menu.scrollY <= 0 {
		t.Fatalf("submenu end navigation = hover %d scroll %v", menu.HoverSubIndex(), menu.scrollY)
	}
}

func TestDropdownArrowScalesWithControlAndStaysInsideBounds(t *testing.T) {
	smallBounds := layout.Rect{X: 10, Y: 20, W: 100, H: 20}
	largeBounds := layout.Rect{X: 10, Y: 20, W: 200, H: 40}
	small := dropdownArrowGeometry(smallBounds, false)
	large := dropdownArrowGeometry(largeBounds, false)

	smallWidth := small[1].x - small[0].x
	largeWidth := large[1].x - large[0].x
	if math.Abs(largeWidth/smallWidth-2) > 1e-9 {
		t.Fatalf("arrow widths = %v and %v, want proportional scaling", smallWidth, largeWidth)
	}
	for _, bounds := range []layout.Rect{{W: 4, H: 20}, smallBounds, largeBounds} {
		for _, open := range []bool{false, true} {
			for _, point := range dropdownArrowGeometry(bounds, open) {
				if point.x < bounds.X || point.x > bounds.X+bounds.W || point.y < bounds.Y || point.y > bounds.Y+bounds.H {
					t.Fatalf("arrow point %+v escaped bounds %+v", point, bounds)
				}
			}
		}
	}
	open := dropdownArrowGeometry(smallBounds, true)
	if small[2].y <= small[0].y || open[2].y >= open[0].y {
		t.Fatal("open and closed arrows do not point in opposite directions")
	}
}
