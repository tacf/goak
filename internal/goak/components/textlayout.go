package components

import "github.com/tacf/goak/internal/goak/rendering"

func textTopY(label string, font *rendering.Font, rowY, rowH float64) float64 {
	_, height := font.Measure(label)
	return rowY + (rowH-height)/2
}

func textHeight(label string, font *rendering.Font) float64 {
	_, height := font.Measure(label)
	return height
}
