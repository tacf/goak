# goak

Go Application Kit is a small library for building cross-platform UI apps.
It uses Ebiten for windowing/input/rendering and focuses on rapid prototyping.

It includes a layout layer inspired by Clay.h and simple named colors like
`colors.LightGray`.

## Run examples

```bash
go run ./examples/basic
go run ./examples/menubar
go run ./examples/scaling
```

**Build:** `go build -o bin/basic ./examples/basic` (on Windows add `-ldflags="-H windowsgui"`).

## Using the library

```go
import (
	"goak/internal/goak"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
)
```

App flow: `goak.NewApp()` -> `InitWindow(title, w, h)` -> build UI -> `app.Run(ui)`.
