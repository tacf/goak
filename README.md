# goak

An immediate-mode UI framework in pure Go. It uses SDL3 for windowing, input
and rendering, with no cgo and no C toolchain. The same UI runs on SDL's
software renderer or a GPU renderer by changing one option.

The API follows the design of Dear ImGui (https://github.com/ocornut/imgui).

Features:

- Immediate-mode API: `if ctx.Button("Save") { save() }`
- Backend interface with an SDL3 implementation
- Software rendering by default, GPU and other SDL drivers by option
- Pure Go font atlas built from any TTF or OTF, with a default font embedded
- HiDPI aware, with the font rebuilt when the display scale changes
- Windows with drag, resize, collapse, close, scrolling and stacking
- Button, checkbox, radio, sliders, drag, progress bar, images
- Single and multiline text input with selection and clipboard
- Combo, list box, popups, tooltips, menus, trees and tabs
- Dark and light styles

## Run examples

```bash
go run ./examples/basic
go run ./examples/hello
go run ./examples/windows
go run ./examples/demo
```

All examples take `-renderer software|gpu`.

**Build:** `go build -o bin/demo ./examples/demo`. The build works with
`CGO_ENABLED=0` and cross compiles to Linux, Windows and macOS.

## Using the library

The application owns the loop. The API lives in the top level `goak` package
and a `Backend` supplies the window, input and rendering.

```go
package main

import (
    "log"

    "github.com/tacf/goak"
    "github.com/tacf/goak/backend/sdl3"
    "github.com/tacf/goak/colors"
)

func main() {
    b, err := sdl3.New(sdl3.Config{Title: "Hello", Resizable: true})
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
        if ctx.Begin("Window", nil, 0) {
            if ctx.Button("Click me") {
                clicks++
            }
            ctx.Textf("clicked %d times", clicks)
        }
        ctx.End()
        _ = ctx.Present(colors.RGB(24, 24, 28))
    }
}
```

`b.NewFrame` polls events, fills the input, sets the frame time and returns
false when the window is closed. Quit on Escape in your own loop:

```go
if in.KeyPressed[goak.KeyEscape] && ctx.FocusID() == 0 && !ctx.AnyPopupOpen() {
    break
}
```

### Backends

A backend implements `goak.Backend`: frame events, display size and scale,
clipboard, texture upload, rendering of `goak.DrawData`, clear and present.
`backend/sdl3` is the SDL3 implementation. To target another platform or
graphics API, implement the interface and pass it to `goak.NewContext`.

### Renderer

Set `sdl3.Config.Renderer` to `"software"` (default), `"gpu"`, `"auto"` or an
SDL driver name such as `"opengl"` or `"vulkan"`.

```go
b, err := sdl3.New(sdl3.Config{Title: "App", Renderer: "gpu"})
```

### Display scale changes

A context made by `NewContext` rebuilds its font in `NewFrame` when the backend
scale changes. The old atlas texture is destroyed and the old font is kept if
the new one fails.

### Skipping unchanged frames

Off by default. `goak.FrameSkipper` compares a hash of the draw data and forces
a redraw every second. Use `Render` and the backend directly to apply it.

```go
var skip goak.FrameSkipper
// ...
dd := ctx.Render()
if skip.Changed(dd) {
    _ = b.Clear(colors.RGB(24, 24, 28))
    _ = b.Render(dd)
    _ = b.Present()
}
```

### Widgets

```go
ctx.Checkbox("Enabled", &enabled)
ctx.SliderFloat("Volume", &volume, 0, 1, "%.2f")
ctx.InputText("Name", &name, 0)
ctx.Combo("Mode", &mode, []string{"Fast", "Balanced", "Slow"})

if ctx.BeginTabBar("tabs") {
    if ctx.BeginTabItem("General", nil) {
        ctx.Text("General settings")
        ctx.EndTabItem()
    }
    ctx.EndTabBar()
}
```

Widget IDs come from the label. Use `##` to hide part of a label and `###` to
keep an ID stable while the text changes. `ctx.PushID` scopes repeated labels.

### Custom drawing and images

`ctx.DL()` returns the current window draw list for custom shapes. Create
textures through the backend and draw them with `ctx.Image`.
