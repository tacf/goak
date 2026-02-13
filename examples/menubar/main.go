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

	app.InitWindow("MenuBar Example", 960, 640)
	app.SetAutoDPI(true)
	ui := buildUI()
	app.Run(ui)
}

func buildUI() *components.UI {
	ui := components.NewUI()
	root := ui.Root()

	// Full-width top menu bar
	mainMenu := root.CreateMenuBar(layout.StaticPx(30), components.MenuBarWidthFull)
	mainMenu.
		AddItem("File", nil).
		AddSubItem("New", func() { fmt.Println("File > New") }).
		AddSubItem("Open", func() { fmt.Println("File > Open") }).
		AddSeparator().
		AddSubItem("Save", func() { fmt.Println("File > Save") }).
		AddSubItem("Exit", func() { fmt.Println("File > Exit") })
	mainMenu.
		AddItem("Edit", nil).
		AddSubItem("Undo", func() { fmt.Println("Edit > Undo") }).
		AddSubItem("Redo", func() { fmt.Println("Edit > Redo") }).
		AddSeparator().
		AddSubItem("Cut", func() { fmt.Println("Edit > Cut") }).
		AddSubItem("Copy", func() { fmt.Println("Edit > Copy") }).
		AddSubItem("Paste", func() { fmt.Println("Edit > Paste") })
	mainMenu.
		AddItem("View", nil).
		AddSubItem("Zoom In", func() { fmt.Println("View > Zoom In") }).
		AddSubItem("Zoom Out", func() { fmt.Println("View > Zoom Out") }).
		AddSeparator().
		AddSubItem("Reset Zoom", func() { fmt.Println("View > Reset Zoom") })
	mainMenu.AddItem("Help", func() { fmt.Println("Help clicked") })

	// Content area
	content := root.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	content.SetBackground(colors.DarkBlue)

	// Auto-width contextual toolbar menu
	tools := content.CreateMenuBar(layout.StaticPx(26), components.MenuBarWidthAuto)
	tools.
		AddItem("Tools", nil).
		AddSubItem("Format", func() { fmt.Println("Tools > Format") }).
		AddSubItem("Sort", func() { fmt.Println("Tools > Sort") }).
		AddSeparator().
		AddSubItem("Preferences", func() { fmt.Println("Tools > Preferences") })
	tools.
		AddItem("Window", nil).
		AddSubItem("Split", func() { fmt.Println("Window > Split") }).
		AddSubItem("Close", func() { fmt.Println("Window > Close") })

	center := content.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	center.SetBackground(colors.DarkGray)
	center.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	info := center.CreateButton(layout.StaticPx(320), layout.StaticPx(36), "Use menus above (top and auto-width)")
	info.OnClick = func() { fmt.Println("info button clicked") }

	return ui
}
