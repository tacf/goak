// Command windows shows the immediate-mode context with two windows.
package main

import (
	"flag"
	"log"

	"github.com/tacf/goak"
	"github.com/tacf/goak/backend/sdl3"
	"github.com/tacf/goak/colors"
)

func main() {
	renderer := flag.String("renderer", "software", "renderer driver: software, gpu, auto or an SDL driver name")
	flag.Parse()

	b, err := sdl3.New(sdl3.Config{Title: "goak windows", Width: 800, Height: 600, Renderer: *renderer, Resizable: true})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	ctx, err := goak.NewContext(b, goak.Options{FontSize: 18})
	if err != nil {
		log.Fatal(err)
	}

	var (
		in       goak.Input
		clicks   int
		checked  bool
		showInfo = true
	)
	for b.NewFrame(&in) {
		ctx.NewFrame(&in, b.DisplaySize())
		ctx.SetNextWindowPos(goak.Vec2{X: 40, Y: 40}, goak.CondOnce)
		ctx.SetNextWindowSize(goak.Vec2{X: 300, Y: 220}, goak.CondOnce)
		if ctx.Begin("Controls", nil, 0) {
			if ctx.Button("Click me") {
				clicks++
			}
			ctx.SameLine(0, -1)
			ctx.Textf("clicked %d times", clicks)
			ctx.Checkbox("Show info window", &showInfo)
			ctx.Checkbox("Another flag##2", &checked)
			ctx.Separator()
			if ctx.BeginChild("log", goak.Vec2{X: 0, Y: 0}, 0) {
				for i := 0; i < 20; i++ {
					ctx.Textf("scrollable line %d", i)
				}
			}
			ctx.EndChild()
		}
		ctx.End()

		ctx.SetNextWindowPos(goak.Vec2{X: 220, Y: 120}, goak.CondOnce)
		if ctx.Begin("Info", &showInfo, goak.AutoSize) {
			ctx.Text("Drag titles to move.")
			ctx.Text("Drag the corner to resize.")
			if checked {
				ctx.Text("Flag is on")
			}
		}
		ctx.End()
		_ = ctx.Present(colors.RGB(24, 24, 28))
		if in.KeyPressed[goak.KeyEscape] && ctx.FocusID() == 0 && !ctx.AnyPopupOpen() {
			break
		}
	}
}
