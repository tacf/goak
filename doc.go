// Package goak is an immediate-mode UI toolkit in pure Go.
//
// The application owns the loop. A Backend supplies the window, input and
// rendering, and a Context turns widget calls into draw data:
//
//	b, _ := sdl3.New(sdl3.Config{Title: "Hello"})
//	defer b.Close()
//	ctx, _ := goak.NewContext(b, goak.Options{FontSize: 18})
//
//	var in goak.Input
//	for b.NewFrame(&in) {
//		ctx.NewFrame(&in, b.DisplaySize())
//		if ctx.Begin("Window", nil, 0) {
//			if ctx.Button("Click me") {
//				// ...
//			}
//		}
//		ctx.End()
//		ctx.Present(colors.RGB(24, 24, 28))
//	}
//
// New backends implement Backend. The SDL3 one lives in backend/sdl3.
package goak
