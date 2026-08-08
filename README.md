# goak

Go Application Kit is a small library for building cross-platform UI apps.
It uses SDL3 for windowing, input, and GPU-accelerated rendering, with SDL_ttf
for text, and focuses on rapid prototyping.


Features:
- Layout management inspired by Clay.h
- Handy utilities like named colors (Example, `colors.LightGray`)
- Themes through a central unified structured (External configuration may come soon)
- SDL3 rendering with resizable and HiDPI-aware windows
- Selectable SDL renderers, with software rendering as the default
- Embedded SDL3, SDL_ttf, and default font assets
- Native window icons from any standard Go `image.Image`
- Image components with contain, cover, stretch, and native-size fitting
- Unicode text inputs and multiline text areas with wrapping, line numbers,
  selection, clipboard editing, and horizontal/vertical scrolling
- Public facing retained-mode ui and custom-scene API for application-
  specific interfaces such as editors, canvases, and media tools

![Demo app showcasing most widgets and features](images/demo.png)

## Run examples

```bash
go run ./examples/<example>
```

Check `examples/demo/main.go` for a widget/features showcase. The basic example
also generates and installs a window icon entirely in Go code. The dispatch
example shows how to bridge a background channel into UI updates safely.
The content example demonstrates images, a text input, and a scrollable text
area; toggle wrapping there to switch between wrapping and horizontal scrolling.

**Build:** `go build -o bin/basic ./examples/basic`. On Windows, optionally add
`-ldflags="-H windowsgui"` to build a GUI executable without an accompanying
console window; omit it during development to retain console logs and panic output.

## Using the library

```go
import (
	"github.com/tacf/goak"
	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)
```

```go
panel.CreateImage(layout.StaticPx(120), layout.StaticPx(80), sourceImage)

name := panel.CreateTextInput(
	layout.PercentOf(100),
	layout.StaticPx(36),
	"",
)
name.SetPlaceholder("Name")

notes := panel.CreateTextArea(
	layout.PercentOf(100),
	layout.StaticPx(180),
	"Long editable text...",
)
notes.SetWrap(true)
notes.SetLineNumbers(true)
```

### Custom scenes

Applications with interfaces that do not fit retained widgets can implement
`goak.Scene`. This gives the framework two integration levels:

- `App.Run(ui)` hosts Goak's retained components and automatic layout.
- `App.RunScene(scene)` hosts a custom interface while retaining Goak's window,
  renderer selection, dispatch queue, input normalization, HiDPI coordinates,
  clipboard, cursor, text-input lifecycle, and frame presentation.

Scenes receive renderer-pixel mouse coordinates and stable keyboard chords
such as `ctrl+shift+p`. Implement `SceneInitializer`, `SceneEventHandler`,
`SceneUpdater`, or `SceneCloser` only when those hooks are needed. Returning
true from a quit event lets the application defer shutdown, for example to
confirm unsaved work. The SDL renderer remains available through
`SceneContext.Renderer()` as an advanced drawing escape hatch. See
`examples/scene` for the minimal lifecycle.
