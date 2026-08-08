package weatherapp

import (
	"image"
	"image/color"
	"math"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	gridOuterPadding = 6
	gridCardGap      = 7
	gridCardRadius   = 12
)

func forecastGridArtwork(daily dailyWeather, width, height int) image.Image {
	width = max(width, 7*90)
	height = max(height, 180)
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{R: 17, G: 24, B: 39, A: 255}
	fillRect(result, 0, 0, width, height, background)

	renderScale := math.Min(float64(width)/850, float64(height)/210)
	pixels := func(value int) int {
		return max(1, int(math.Round(float64(value)*renderScale)))
	}
	dayFace := newGridFace(14 * renderScale)
	detailFace := newGridFace(11.5 * renderScale)
	defer dayFace.Close()
	defer detailFace.Close()

	outerPadding := pixels(gridOuterPadding)
	cardGap := pixels(gridCardGap)
	cardRadius := pixels(gridCardRadius)
	availableW := width - outerPadding*2 - cardGap*(forecastDays-1)
	cardW := availableW / forecastDays
	cardH := height - outerPadding*2
	for index := range forecastDays {
		left := outerPadding + index*(cardW+cardGap)
		if index == forecastDays-1 {
			cardW = width - outerPadding - left
		}
		card := image.Rect(left, outerPadding, left+cardW, outerPadding+cardH)
		fill := color.RGBA{R: 29, G: 40, B: 59, A: 255}
		stroke := color.RGBA{R: 47, G: 62, B: 83, A: 255}
		if index == 0 {
			fill = color.RGBA{R: 43, G: 62, B: 88, A: 255}
			stroke = color.RGBA{R: 78, G: 108, B: 145, A: 255}
		}
		drawRoundedCard(result, card, cardRadius, pixels(1), fill, stroke)

		centerX := (card.Min.X + card.Max.X) / 2
		textWidth := cardW - pixels(12)
		drawCenteredText(result, dayFace, cardDay(daily.Time[index]), centerX, card.Min.Y+pixels(23), color.RGBA{R: 225, G: 231, B: 239, A: 255}, textWidth)
		drawCenteredText(result, detailFace, cardDate(daily.Time[index]), centerX, card.Min.Y+pixels(40), color.RGBA{R: 148, G: 163, B: 184, A: 255}, textWidth)

		iconY := card.Min.Y + pixels(77)
		drawWeatherGlyph(result, daily.WeatherCode[index], true, centerX, iconY, 0.42*renderScale, fill)

		drawCenteredText(
			result,
			detailFace,
			weatherDescription(daily.WeatherCode[index]),
			centerX,
			card.Min.Y+pixels(116),
			color.RGBA{R: 226, G: 232, B: 240, A: 255},
			textWidth,
		)
		drawCenteredText(
			result,
			detailFace,
			formatTemperatureRange(daily.TemperatureMin[index], daily.TemperatureMax[index]),
			centerX,
			card.Min.Y+pixels(139),
			color.RGBA{R: 251, G: 191, B: 36, A: 255},
			textWidth,
		)
		drawCenteredText(
			result,
			detailFace,
			formatRain(daily.PrecipitationProbability[index]),
			centerX,
			card.Min.Y+pixels(158),
			color.RGBA{R: 125, G: 211, B: 252, A: 255},
			textWidth,
		)
		drawCenteredText(
			result,
			detailFace,
			formatWind(daily.WindSpeedMax[index]),
			centerX,
			card.Min.Y+pixels(177),
			color.RGBA{R: 203, G: 213, B: 225, A: 255},
			textWidth,
		)
	}
	return result
}

var (
	gridFontOnce sync.Once
	gridFont     *opentype.Font
)

func newGridFace(size float64) font.Face {
	gridFontOnce.Do(func() {
		gridFont, _ = opentype.Parse(goregular.TTF)
	})
	if gridFont == nil {
		return basicfont.Face7x13
	}
	face, err := opentype.NewFace(gridFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return basicfont.Face7x13
	}
	return face
}

func drawCenteredText(
	target *image.RGBA,
	face font.Face,
	text string,
	centerX, baselineY int,
	textColor color.RGBA,
	maxWidth int,
) {
	text = fitGridText(face, text, maxWidth)
	drawer := font.Drawer{
		Dst:  target,
		Src:  image.NewUniform(textColor),
		Face: face,
	}
	textWidth := drawer.MeasureString(text).Ceil()
	drawer.Dot = fixed.P(centerX-textWidth/2, baselineY)
	drawer.DrawString(text)
}

func fitGridText(face font.Face, text string, maxWidth int) string {
	if font.MeasureString(face, text).Ceil() <= maxWidth {
		return text
	}
	const suffix = "..."
	runes := []rune(text)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := strings.TrimSpace(string(runes)) + suffix
		if font.MeasureString(face, candidate).Ceil() <= maxWidth {
			return candidate
		}
	}
	return suffix
}

func drawRoundedCard(target *image.RGBA, rect image.Rectangle, radius, strokeWidth int, fill, stroke color.RGBA) {
	fillRoundedRect(target, rect, radius, stroke)
	fillRoundedRect(target, rect.Inset(strokeWidth), max(1, radius-strokeWidth), fill)
}

func fillRoundedRect(target *image.RGBA, rect image.Rectangle, radius int, fill color.RGBA) {
	rect = rect.Intersect(target.Bounds())
	if rect.Empty() {
		return
	}
	radius = min(radius, min(rect.Dx(), rect.Dy())/2)
	fillRect(target, rect.Min.X+radius, rect.Min.Y, rect.Max.X-radius, rect.Max.Y, fill)
	fillRect(target, rect.Min.X, rect.Min.Y+radius, rect.Max.X, rect.Max.Y-radius, fill)
	fillCircle(target, rect.Min.X+radius, rect.Min.Y+radius, radius, fill)
	fillCircle(target, rect.Max.X-radius-1, rect.Min.Y+radius, radius, fill)
	fillCircle(target, rect.Min.X+radius, rect.Max.Y-radius-1, radius, fill)
	fillCircle(target, rect.Max.X-radius-1, rect.Max.Y-radius-1, radius, fill)
}

func cardDay(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("Mon")
}

func cardDate(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}
	return parsed.Format("2 Jan")
}

func formatTemperatureRange(minimum, maximum float64) string {
	return formatRounded(minimum) + " / " + formatRounded(maximum) + "°C"
}

func formatRain(probability int) string {
	return "Rain " + formatInteger(probability) + "%"
}

func formatWind(speed float64) string {
	return "Wind " + formatRounded(speed) + " km/h"
}

func formatRounded(value float64) string {
	return formatInteger(int(math.Round(value)))
}

func formatInteger(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		position--
		digits[position] = '-'
	}
	return string(digits[position:])
}
