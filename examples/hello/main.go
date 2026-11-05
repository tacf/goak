// Command hello draws rects, circles and text with the draw list on SDL3.
// It stays on the low level backend API on purpose to show draw lists directly.
// See basic for the widget loop.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/tacf/goak"
	"github.com/tacf/goak/backend/sdl3"
	"github.com/tacf/goak/colors"
)

func main() {
	driver := flag.String("renderer", "software", "renderer driver: software, gpu, auto or an SDL driver name")
	flag.Parse()

	b, err := sdl3.New(sdl3.Config{
		Title: "goak hello", Width: 800, Height: 600,
		Renderer: *driver, Resizable: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	fmt.Println("renderer:", b.RendererName())

	fnt, err := goak.NewFont(nil, 20, b.Scale())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { fnt.Close() }()

	dl := goak.NewDrawList()
	var in goak.Input
	var t float32
	for b.NewFrame(&in) {
		if in.KeyPressed[goak.KeyEscape] {
			break
		}
		if s, ok := b.ScaleChanged(); ok {
			nf, err := goak.NewFont(nil, 20, s)
			if err != nil {
				log.Println("font:", err)
			} else {
				if t := fnt.Atlas().Texture(); t != 0 {
					b.DestroyTexture(t)
				}
				fnt.Close()
				fnt = nf
			}
		}
		t += in.DT
		dl.Reset()
		dl.AddRectFilled(goak.Vec2{X: 40, Y: 40}, goak.Vec2{X: 300, Y: 200}, colors.RGB(60, 90, 160), 12)
		dl.AddRect(goak.Vec2{X: 40, Y: 40}, goak.Vec2{X: 300, Y: 200}, colors.RGB(240, 240, 240), 12, 2)
		dl.AddCircleFilled(goak.Vec2{X: 480, Y: 120}, 60, colors.RGB(200, 80, 80))
		dl.AddCircle(goak.Vec2{X: 480, Y: 120}, 70, 3, colors.RGB(240, 200, 80))
		orbit := goak.Vec2{
			X: 480 + 70*float32(math.Cos(float64(t*2))),
			Y: 120 + 70*float32(math.Sin(float64(t*2))),
		}
		dl.AddCircleFilled(orbit, 8, colors.RGB(250, 250, 250))
		dl.AddCircleFilled(goak.Vec2{X: in.MousePos[0], Y: in.MousePos[1]}, 14, colors.RGBA(80, 220, 120, 200))
		if err := fnt.DrawText(dl, b, 60, 100, "Hello from the goak font atlas", colors.RGB(250, 250, 250)); err != nil {
			log.Println("text:", err)
		}
		fnt.Finish(dl)

		_ = b.Clear(colors.RGB(24, 24, 28))
		if err := b.Render(b.DrawData(dl)); err != nil {
			log.Println("render:", err)
		}
		_ = b.Present()
	}
}
