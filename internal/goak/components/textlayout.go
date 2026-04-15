package components

import "github.com/hajimehoshi/ebiten/v2/text/v2"

func textTopY(label string, face text.GoTextFace, rowY, rowH float64) int {
	_, th := text.Measure(label, &face, 0)
	return int(rowY + (rowH-th)/2)
}

func textHeight(label string, face text.GoTextFace) float64 {
	_, th := text.Measure(label, &face, 0)
	return th
}
