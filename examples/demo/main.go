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

	app.InitWindow("Goak Demo", 800, 650)
	app.SetAutoDPI(true)
	app.SetScaleHotkeysEnabled(true)
	ui := buildUI()

	app.Run(ui)
}

func buildUI() *components.UI {
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

	container := root.CreatePanel(layout.PercentOf(100), layout.PercentOf(100))
	container.SetBackground(colors.HexOr("#1e1e1e", colors.RGB(30, 30, 30)))
	container.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	buttonSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(100))
	buttonSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	buttonSection.SetAlignment(layout.AlignStart, layout.AlignCenter)

	btn1 := buttonSection.CreateButton(layout.StaticPx(120), layout.StaticPx(32), "Click Me!")
	btn1.OnClick = func() { fmt.Println("Button 1 clicked") }

	btn2 := buttonSection.CreateButton(layout.StaticPx(120), layout.StaticPx(32), "Press Me!")
	btn2.OnClick = func() { fmt.Println("Button 2 clicked") }

	checkboxSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(110))
	checkboxSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	checkboxSection.SetAlignment(layout.AlignStart, layout.AlignStart)

	cb1 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature A")
	cb1.OnChanged = func(checked bool) {
		fmt.Printf("Feature A: %v\n", checked)
	}

	cb2 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature B")
	cb2.OnChanged = func(checked bool) {
		fmt.Printf("Feature B: %v\n", checked)
	}

	cb3 := checkboxSection.CreateCheckbox(layout.StaticPx(200), layout.StaticPx(24), "Enable feature C")
	cb3.Checked = true
	cb3.OnChanged = func(checked bool) {
		fmt.Printf("Feature C: %v\n", checked)
	}

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
	radio.SelectedIndex = 0
	radio.OnChanged = func(index int, value string) {
		fmt.Printf("Radio selected: %s (index %d)\n", value, index)
	}

	sliderSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(80))
	sliderSection.SetBackground(colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)))
	sliderSection.SetAlignment(layout.AlignStart, layout.AlignStart)

	slider := sliderSection.CreateSlider(layout.StaticPx(400), layout.StaticPx(60), "Volume", 0, 100, 50)
	slider.OnChanged = func(value float64) {
		fmt.Printf("Slider value: %.1f\n", value)
	}

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
	dropdown.OnChanged = func(index int, value string) {
		fmt.Printf("Dropdown selected: %s (index %d)\n", value, index)
	}

	contextMenu := components.NewContextMenu([]components.ContextMenuItem{
		{Kind: components.ContextMenuItemAction, Label: "Copy", OnClick: func() { fmt.Println("Context: Copy") }},
		{Kind: components.ContextMenuItemAction, Label: "Paste", OnClick: func() { fmt.Println("Context: Paste") }},
		{Kind: components.ContextMenuItemSeparator},
		{Kind: components.ContextMenuItemAction, Label: "Delete", OnClick: func() { fmt.Println("Context: Delete") }},
		{Kind: components.ContextMenuItemAction, Label: "Properties", OnClick: func() { fmt.Println("Context: Properties") }},
	})
	container.AddContextMenu(contextMenu)

	infoSection := container.CreatePanel(layout.PercentOf(95), layout.StaticPx(60))
	infoSection.SetBackground(colors.HexOr("#252525", colors.RGB(37, 37, 37)))
	infoSection.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	infoBtn := infoSection.CreateButton(layout.StaticPx(250), layout.StaticPx(36), "Demo")
	infoBtn.OnClick = func() {
		fmt.Println("This demo shows all available components:")
		fmt.Println("- Buttons, Checkboxes, Radio Groups")
		fmt.Println("- Sliders, Dropdowns, Context Menus")
		fmt.Println("- Menu Bars with submenus")
		fmt.Println("Try Ctrl+/- to scale the UI!")
	}

	return ui
}
