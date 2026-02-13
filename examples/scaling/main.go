package main

import (
	"fmt"

	"goak/internal/goak"
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
)

func main() {
	app := goak.NewApp()
	defer app.Destroy()

	app.InitWindowWithConfig(goak.Config{
		Title:       "Scaling Example (Ctrl +/-)",
		Width:       960,
		Height:      640,
		AutoDPI:     true,
		WindowScale: 1.0,
	})

	app.SetScaleHotkeysEnabled(true)

	if win := app.Window(); win != nil {
		win.SetOnWindowScaleChanged(func(s float64) {
			title := fmt.Sprintf("Scaling Example (Ctrl +/-) - scale: %.1fx", s)
			win.SetTitle(title)
			fmt.Printf("window scale changed: %.1fx\n", s)
		})
		// Trigger initial title state.
		win.SetWindowScale(win.WindowScale())
	}

	ui := buildUI()
	app.Run(ui)
}

func buildUI() *components.UI {
	ui := components.NewUI()
	root := ui.Root()

	menu := root.CreateMenuBar(layout.StaticPx(30), components.MenuBarWidthFull)
	menu.
		AddItem("View", nil).
		AddSubItem("Zoom In  (Ctrl +)", nil).
		AddSubItem("Zoom Out (Ctrl -)", nil).
		AddSeparator().
		AddSubItem("Reset (set scale to 1.0 in code)", nil)

	content := root.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	content.SetBackground(colors.DarkBlue)
	content.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	info := content.CreateButton(layout.StaticPx(500), layout.StaticPx(42), "Press Ctrl + / Ctrl - to scale the whole app")
	info.OnClick = func() { fmt.Println("hint: use keyboard shortcuts") }

	return ui
}
