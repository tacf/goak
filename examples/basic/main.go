// Command basic is the minimal explicit frame loop.
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

	b, err := sdl3.New(sdl3.Config{Title: "goak basic", Renderer: *renderer, Resizable: true})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	ctx, err := goak.NewContext(b, goak.Options{FontSize: 18})
	if err != nil {
		log.Fatal(err)
	}

	var in goak.Input
	clicks := 0
	for b.NewFrame(&in) {
		ctx.NewFrame(&in, b.DisplaySize())
		if ctx.Begin("Hello", nil, 0) {
			if ctx.Button("Click me") {
				clicks++
			}
			ctx.Textf("clicked %d times", clicks)
		}
		ctx.End()
		if err := ctx.Present(colors.RGB(24, 24, 28)); err != nil {
			log.Println("present:", err)
		}
		if in.KeyPressed[goak.KeyEscape] && ctx.FocusID() == 0 && !ctx.AnyPopupOpen() {
			break
		}
	}
}
