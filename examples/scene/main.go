package main

import (
	"log"

	"github.com/tacf/goak"
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"

	"github.com/Zyko0/go-sdl3/sdl"
)

type scene struct {
	background sdl.Color
}

func (s *scene) Draw(ctx *goak.SceneContext) {
	renderer := ctx.Renderer()
	_ = renderer.SetDrawColor(s.background.R, s.background.G, s.background.B, s.background.A)
	_ = renderer.Clear()
}

func sceneUI(s *scene) *components.UI {
	ui := components.NewUI()
	ui.SetFontSize(18)
	root := ui.Root()
	root.SetInsets(layout.Insets{Top: 16, Right: 16, Left: 16})

	bar := root.CreatePanel(layout.PercentOf(100), layout.StaticPx(56))
	bar.SetBackground(colors.RGBA(33, 38, 48, 235))
	bar.SetDirection(layout.Row)
	bar.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	bar.SetGap(12)
	bar.SetInsets(layout.Insets{Right: 12, Left: 12})
	bar.CreateLabel(layout.StaticPx(90), layout.StaticPx(32), "Scene UI")
	input := bar.CreateTextInput(
		layout.AutoSize().WithMax(320), layout.StaticPx(34), "",
	)
	input.SetPlaceholder("Retained controls over a custom scene")
	button := bar.CreateButton(layout.StaticPx(120), layout.StaticPx(34), "Change color")
	button.SetOnClick(func(components.ButtonClickEvent) {
		s.background = sdl.Color{R: s.background.B, G: s.background.R, B: s.background.G, A: 255}
	})
	return ui
}

func main() {
	app := goak.NewApp()
	if err := app.TryInitWindowWithConfig(goak.Config{
		Title: "Goak Scene", Width: 900, Height: 600,
		Renderer: goak.RendererAuto, AutoDPI: true,
	}); err != nil {
		log.Fatal(err)
	}
	defer app.Destroy()
	s := &scene{background: sdl.Color{R: 24, G: 27, B: 34, A: 255}}
	if err := app.RunSceneWithUI(s, sceneUI(s)); err != nil {
		log.Fatal(err)
	}
}
