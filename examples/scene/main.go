package main

import (
	"log"

	"github.com/tacf/goak"
)

type scene struct{}

func (scene) Draw(ctx *goak.SceneContext) {
	renderer := ctx.Renderer()
	_ = renderer.SetDrawColor(24, 27, 34, 255)
	_ = renderer.Clear()
}

func main() {
	app := goak.NewApp()
	if err := app.TryInitWindowWithConfig(goak.Config{
		Title: "Goak Scene", Width: 900, Height: 600,
		Renderer: goak.RendererAuto,
	}); err != nil {
		log.Fatal(err)
	}
	defer app.Destroy()
	if err := app.RunScene(scene{}); err != nil {
		log.Fatal(err)
	}
}
