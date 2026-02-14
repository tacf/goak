package rendering

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

// FillRect draws a filled rectangle.
func FillRect(dst *ebiten.Image, x, y, w, h float64, c colors.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, true)
}

// DrawStrokeRect draws a rectangular outline with the given thickness.
func DrawStrokeRect(dst *ebiten.Image, x, y, w, h, thickness float64, c colors.Color) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(thickness), c, true)
	vector.DrawFilledRect(dst, float32(x), float32(y+h-thickness), float32(w), float32(thickness), c, true)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(thickness), float32(h), c, true)
	vector.DrawFilledRect(dst, float32(x+w-thickness), float32(y), float32(thickness), float32(h), c, true)
}

// DrawLine draws a horizontal or vertical line.
func DrawLine(dst *ebiten.Image, x, y, length, thickness float64, c colors.Color, horizontal bool) {
	if horizontal {
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(length), float32(thickness), c, true)
	} else {
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(thickness), float32(length), c, true)
	}
}

// DrawFilledCircle draws a filled circle.
func DrawFilledCircle(dst *ebiten.Image, centerX, centerY, radius float64, c colors.Color) {
	vector.DrawFilledCircle(dst, float32(centerX), float32(centerY), float32(radius), c, true)
}

// DrawCircleStroke draws a circle outline.
func DrawCircleStroke(dst *ebiten.Image, centerX, centerY, radius, thickness float64, c colors.Color) {
	vector.StrokeCircle(dst, float32(centerX), float32(centerY), float32(radius), float32(thickness), c, true)
}

// DrawText renders text at the specified position.
func DrawText(dst *ebiten.Image, str string, face font.Face, x, y int, c colors.Color) {
	text.Draw(dst, str, face, x, y, c)
}

// PointWithinBounds returns true if the point (x, y) is inside the given rectangle.
func PointWithinBounds(x, y float64, r layout.Rect) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}
