package components_test

import (
	"testing"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

func TestPublicContextMenuConfiguration(t *testing.T) {
	menu := components.NewContextMenu([]components.ContextMenuItem{
		components.NewContextMenuActionWithHint("Open", "Ctrl+O", nil),
		components.NewContextMenuSeparator(),
		components.NewDisabledContextMenuActionWithHint("Delete", "Del"),
	})
	menu.SetItems(menu.Items())
	menu.SetPadding(8, 4)
	menu.SetSeparatorHeight(6)
	menu.SetMinWidth(100)
	menu.OpenAt(95, 95, layout.Rect{W: 100, H: 100})

	if !menu.IsOpen() || menu.Bounds().W != 100 {
		t.Fatalf("configured menu bounds = %+v", menu.Bounds())
	}
}
