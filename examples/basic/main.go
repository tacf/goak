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

	app.InitWindow("Canvas", 800, 600)
	app.SetAutoDPI(true)
	ui := buildUI()

	app.Run(ui)
}

func buildUI() *components.UI {
	ui := components.NewUI()
	root := ui.Root()

	mainMenu := root.CreateMenuBar(layout.StaticPx(28), components.MenuBarWidthFull)
	mainMenu.
		AddItem("File", nil).
		AddSubItem("New", func() { fmt.Println("new") }).
		AddSubItem("Open", func() { fmt.Println("open") }).
		AddSeparator().
		AddSubItem("Exit", func() { fmt.Println("exit") })
	mainMenu.
		AddItem("Edit", nil).
		AddSubItem("Cut", func() { fmt.Println("cut") }).
		AddSubItem("Copy", func() { fmt.Println("copy") }).
		AddSubItem("Paste", func() { fmt.Println("paste") })
	mainMenu.AddItem("Help", func() { fmt.Println("help clicked") })

	panel := root.CreatePanel(layout.PercentOf(100), layout.StaticPx(200))
	panel.SetBackground(colors.DarkGray)
	panel.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	btn := panel.CreateButton(layout.StaticPx(120), layout.StaticPx(32), "Click me")
	btn.OnClick = func() {
		fmt.Println("button clicked")
	}

	panel2 := root.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	panel2.SetBackground(colors.DarkBlue)
	tools := panel2.CreateMenuBar(layout.StaticPx(26), components.MenuBarWidthAuto)
	tools.
		AddItem("Tools", nil).
		AddSubItem("Format", func() { fmt.Println("format") }).
		AddSeparator().
		AddSubItem("Preferences", func() { fmt.Println("preferences") })

	okBtn := components.NewButton(layout.StaticPx(100), layout.StaticPx(28), "OK")
	okBtn.OnClick = func() { fmt.Println("OK clicked") }
	panel2.AddButton(okBtn)

	return ui
}
