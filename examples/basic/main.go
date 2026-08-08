package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"goak"
	"goak/colors"
	"goak/components"
	"goak/layout"
)

func main() {
	app := goak.NewApp()
	defer app.Destroy()

	app.InitWindow("Canvas", 800, 600)
	if err := app.SetWindowIcon(makeWindowIcon()); err != nil {
		log.Printf("could not set window icon: %v", err)
	}
	app.SetAutoDPI(true)
	ui := buildUI()

	app.Run(ui)
}

// makeWindowIcon creates a transparent 64x64 icon with a simple "G" mark.
func makeWindowIcon() image.Image {
	const size = 64
	icon := image.NewRGBA(image.Rect(0, 0, size, size))
	background := color.RGBA{R: 31, G: 36, B: 48, A: 255}
	accent := color.RGBA{R: 74, G: 158, B: 255, A: 255}

	for y := range size {
		for x := range size {
			// Rounded-square background.
			cornerX := math.Max(12, math.Min(float64(x), 51))
			cornerY := math.Max(12, math.Min(float64(y), 51))
			if math.Hypot(float64(x)-cornerX, float64(y)-cornerY) <= 9 {
				icon.SetRGBA(x, y, background)
			}

			// Circular G, opened on the right with a horizontal crossbar.
			distance := math.Hypot(float64(x)-31.5, float64(y)-31.5)
			onRing := distance >= 15 && distance <= 23
			openRight := x >= 39 && y >= 25 && y <= 33
			crossbar := x >= 31 && x <= 51 && y >= 31 && y <= 37
			if (onRing && !openRight) || crossbar {
				icon.SetRGBA(x, y, accent)
			}
		}
	}
	return icon
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
	btn.SetOnClick(func(components.ButtonClickEvent) {
		fmt.Println("button clicked")
	})

	panel2 := root.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	panel2.SetBackground(colors.DarkBlue)
	tools := panel2.CreateMenuBar(layout.StaticPx(26), components.MenuBarWidthAuto)
	tools.
		AddItem("Tools", nil).
		AddSubItem("Format", func() { fmt.Println("format") }).
		AddSeparator().
		AddSubItem("Preferences", func() { fmt.Println("preferences") })

	okBtn := components.NewButton(layout.StaticPx(100), layout.StaticPx(28), "OK")
	okBtn.SetOnClick(func(components.ButtonClickEvent) { fmt.Println("OK clicked") })
	panel2.AddButton(okBtn)

	return ui
}
