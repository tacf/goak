package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/tacf/goak"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

func main() {
	app := goak.NewApp()
	defer app.Destroy()

	if err := app.SetRenderer(goak.RendererSoftware); err != nil {
		log.Fatalf("could not select renderer: %v", err)
	}
	app.InitWindow("Goak Demo", 800, 650)
	if err := app.SetWindowIcon(makeWindowIcon()); err != nil {
		log.Printf("could not set window icon: %v", err)
	}
	app.SetAutoDPI(true)
	app.SetScaleHotkeysEnabled(true)
	ui := buildUI(app.RendererName())

	app.Run(ui)
}

func buildUI(rendererName string) *components.UI {
	ui := components.NewUI()
	root := ui.Root()
	root.SetAlignment(layout.AlignStart, layout.AlignStart)

	mainMenu := root.CreateMenuBar(layout.StaticPx(28), components.MenuBarWidthFull)
	mainMenu.
		AddItem("File", nil).
		AddSubItem("New", func() { fmt.Println("File -> New") }).
		AddSubItem("Open", func() { fmt.Println("File -> Open") }).
		AddSeparator().
		AddSubItem("Exit", func() { fmt.Println("File -> Exit") })
	mainMenu.
		AddItem("Edit", nil).
		AddSubItem("Cut", func() { fmt.Println("Edit -> Cut") }).
		AddSubItem("Copy", func() { fmt.Println("Edit -> Copy") }).
		AddSubItem("Paste", func() { fmt.Println("Edit -> Paste") })
	mainMenu.AddItem("Help", func() { fmt.Println("Help clicked") })

	container := root.CreatePanel(layout.PercentOf(100), layout.AutoSize())
	container.SetBackground(colors.HexOr("#1e1e1e", colors.RGB(30, 30, 30)))
	container.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	container.SetPadding(16)

	buttonSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(100))
	buttonSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	buttonSection.SetAlignment(layout.AlignStart, layout.AlignCenter)
	buttonSection.SetPadding(12)

	btn1 := buttonSection.CreateButton(layout.StaticPx(120), layout.StaticPx(32), "Click Me!")
	btn1.SetOnClick(func(components.ButtonClickEvent) { fmt.Println("Button 1 clicked") })

	btn2 := buttonSection.CreateButton(layout.StaticPx(120), layout.StaticPx(32), "Press Me!")
	btn2.SetOnClick(func(components.ButtonClickEvent) { fmt.Println("Button 2 clicked") })

	checkboxSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(110))
	checkboxSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	checkboxSection.SetAlignment(layout.AlignStart, layout.AlignStart)

	cb1 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature A")
	cb1.SetOnChanged(func(event components.CheckboxChangedEvent) {
		fmt.Printf("Feature A: %v\n", event.Checked)
	})

	cb2 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature B")
	cb2.SetOnChanged(func(event components.CheckboxChangedEvent) {
		fmt.Printf("Feature B: %v\n", event.Checked)
	})

	cb3 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature C")
	cb3.SetChecked(true)
	cb3.SetOnChanged(func(event components.CheckboxChangedEvent) {
		fmt.Printf("Feature C: %v\n", event.Checked)
	})

	radioSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(120))
	radioSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	radioSection.SetAlignment(layout.AlignStart, layout.AlignStart)

	radioOptions := []components.RadioOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
		{Label: "Option 4", Value: "opt4"},
	}
	radio := radioSection.CreateRadioGroup(layout.StaticPx(200), layout.StaticPx(110), radioOptions)
	radio.SetSelectedIndex(0)
	radio.SetOnChanged(func(event components.RadioChangedEvent) {
		fmt.Printf("Radio selected: %s (index %d)\n", event.Option.Value, event.Index)
	})

	sliderSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(80))
	sliderSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	sliderSection.SetAlignment(layout.AlignStart, layout.AlignStart)

	slider := sliderSection.CreateSlider(layout.StaticPx(400), layout.StaticPx(60), "Volume", 0, 100, 50)
	slider.SetOnChanged(func(event components.SliderChangedEvent) {
		fmt.Printf("Slider value: %.1f\n", event.Value)
	})

	dropdownSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(80))
	dropdownSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	dropdownSection.SetAlignment(layout.AlignStart, layout.AlignCenter)

	dropdownOptions := []components.DropdownOption{
		{Label: "Red", Value: "red"},
		{Label: "Green", Value: "green"},
		{Label: "Blue", Value: "blue"},
		{Label: "Yellow", Value: "yellow"},
		{Label: "Purple", Value: "purple"},
	}
	dropdown := dropdownSection.CreateDropdown(layout.StaticPx(200), layout.StaticPx(32), "Select Color", dropdownOptions)
	dropdown.SetOnChanged(func(event components.DropdownChangedEvent) {
		fmt.Printf("Dropdown selected: %s (index %d)\n", event.Option.Value, event.Index)
	})

	contextMenu := components.NewContextMenu([]components.ContextMenuItem{
		components.NewContextMenuAction("Copy", func() { fmt.Println("Context: Copy") }),
		components.NewContextMenuAction("Paste", func() { fmt.Println("Context: Paste") }),
		components.NewContextMenuSeparator(),
		components.NewContextMenuAction("Delete", func() { fmt.Println("Context: Delete") }),
		components.NewContextMenuAction("Properties", func() { fmt.Println("Context: Properties") }),
	})
	container.AddContextMenu(contextMenu)

	infoSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(60))
	infoSection.SetBackground(colors.HexOr("#252525", colors.RGB(37, 37, 37)))
	infoSection.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	infoBtn := infoSection.CreateButton(layout.StaticPx(250), layout.StaticPx(36), "Demo")
	infoBtn.SetOnClick(func(components.ButtonClickEvent) {
		fmt.Println("This demo shows all available components:")
		fmt.Println("- Buttons, Checkboxes, Radio Groups")
		fmt.Println("- Sliders, Dropdowns, Context Menus")
		fmt.Println("- Menu Bars with submenus")
		fmt.Println("Try Ctrl+/- to scale the UI!")
	})

	statusBar := root.CreatePanel(layout.PercentOf(100), layout.StaticPx(30))
	statusBar.SetBackground(colors.HexOr("#181818", colors.RGB(24, 24, 24)))
	statusBar.SetAlignment(layout.AlignEnd, layout.AlignCenter)
	statusBar.SetPadding(4)

	rendererLabel := statusBar.CreateLabel(
		layout.StaticPx(220),
		layout.AutoSize(),
		"Renderer: "+rendererName,
	)
	rendererLabel.SetAlignment(layout.AlignEnd, layout.AlignCenter)
	rendererLabel.SetColor(colors.HexOr("#a8a8a8", colors.RGB(168, 168, 168)))

	return ui
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
