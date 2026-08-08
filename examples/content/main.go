package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/tacf/goak"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

func main() {
	app := goak.NewApp()
	defer app.Destroy()
	app.InitWindow("Goak Content Components", 720, 540)

	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.PercentOf(100), layout.PercentOf(100))
	panel.SetPadding(12)
	panel.SetBackground(colors.HexOr("#202020", colors.RGB(32, 32, 32)))

	picture := panel.CreateImage(layout.PercentOf(100), layout.StaticPx(90), sampleImage())
	picture.SetFit(components.ImageFitContain)

	input := panel.CreateTextInput(layout.PercentOf(100), layout.StaticPx(38), "")
	input.SetPlaceholder("Type a title and press Enter")
	input.SetOnSubmitted(func(event components.TextInputSubmittedEvent) {
		fmt.Printf("submitted: %q\n", event.Text)
	})

	area := panel.CreateTextArea(
		layout.PercentOf(100),
		layout.AutoSize(),
		"Goak text areas wrap long lines by default.\n"+
			"They also display logical line numbers and scroll vertically.\n"+
			"Disable wrapping below, then use Shift+wheel to scroll a long line horizontally.\n"+
			"0123456789 abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ "+
			"0123456789 abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	)
	area.SetWrap(true)
	area.SetLineNumbers(true)

	wrap := panel.CreateCheckbox(layout.PercentOf(100), layout.StaticPx(28), "Wrap long lines")
	wrap.SetChecked(true)
	wrap.SetOnChanged(func(event components.CheckboxChangedEvent) {
		area.SetWrap(event.Checked)
	})

	app.Run(ui)
}

func sampleImage() image.Image {
	const size = 96
	result := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			cell := (x/12 + y/12) % 2
			pixel := color.RGBA{R: 45, G: 51, B: 65, A: 255}
			if cell == 0 {
				pixel = color.RGBA{R: 74, G: 158, B: 255, A: 255}
			}
			result.SetRGBA(x, y, pixel)
		}
	}
	return result
}
