package weatherapp

import (
	"image"
	"image/color"
	"math"
)

type weatherKind int

const (
	weatherClear weatherKind = iota
	weatherPartlyCloudy
	weatherCloudy
	weatherFog
	weatherRain
	weatherSnow
	weatherThunder
)

func weatherKindForCode(code int) weatherKind {
	switch {
	case code == 0:
		return weatherClear
	case code >= 1 && code <= 2:
		return weatherPartlyCloudy
	case code == 3:
		return weatherCloudy
	case code == 45 || code == 48:
		return weatherFog
	case code >= 71 && code <= 77, code == 85 || code == 86:
		return weatherSnow
	case code >= 95:
		return weatherThunder
	default:
		return weatherRain
	}
}

func weatherDescription(code int) string {
	switch code {
	case 0:
		return "Clear sky"
	case 1:
		return "Mainly clear"
	case 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45:
		return "Fog"
	case 48:
		return "Rime fog"
	case 51:
		return "Light drizzle"
	case 53:
		return "Drizzle"
	case 55:
		return "Dense drizzle"
	case 56, 57:
		return "Freezing drizzle"
	case 61:
		return "Light rain"
	case 63:
		return "Rain"
	case 65:
		return "Heavy rain"
	case 66, 67:
		return "Freezing rain"
	case 71:
		return "Light snow"
	case 73:
		return "Snow"
	case 75:
		return "Heavy snow"
	case 77:
		return "Snow grains"
	case 80:
		return "Light showers"
	case 81:
		return "Rain showers"
	case 82:
		return "Heavy showers"
	case 85:
		return "Snow showers"
	case 86:
		return "Heavy snow showers"
	case 95:
		return "Thunderstorm"
	case 96, 99:
		return "Thunderstorm with hail"
	default:
		return "Mixed conditions"
	}
}

func weatherArtwork(code int, isDay bool, width, height int) image.Image {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	kind := weatherKindForCode(code)
	top, bottom := skyColors(kind, isDay)
	drawGradient(result, top, bottom)

	scale := math.Min(float64(width)/220, float64(height)/110)
	drawWeatherGlyph(result, code, isDay, width/2, height/2, scale, bottom)
	return result
}

func drawWeatherGlyph(
	target *image.RGBA,
	code int,
	isDay bool,
	centerX, centerY int,
	scale float64,
	background color.RGBA,
) {
	kind := weatherKindForCode(code)
	if kind == weatherClear {
		drawSunOrMoon(target, centerX, centerY, int(25*scale), isDay, background)
		return
	}
	if kind == weatherFog {
		drawCloud(target, centerX, centerY-int(12*scale), scale, kind)
		lineColor := color.RGBA{R: 220, G: 231, B: 239, A: 255}
		for row := 0; row < 3; row++ {
			y := centerY + int(float64(15+row*12)*scale)
			drawLine(target, centerX-int(54*scale), y, centerX+int(54*scale), y, max(1, int(3*scale)), lineColor)
		}
		return
	}

	if kind == weatherPartlyCloudy {
		drawSunOrMoon(
			target,
			centerX-int(34*scale),
			centerY-int(22*scale),
			int(20*scale),
			isDay,
			background,
		)
	}
	drawCloud(target, centerX, centerY, scale, kind)

	switch kind {
	case weatherRain:
		rain := color.RGBA{R: 91, G: 190, B: 255, A: 255}
		for column := -2; column <= 2; column++ {
			x := centerX + int(float64(column*18)*scale)
			drawLine(
				target,
				x,
				centerY+int(22*scale),
				x-int(7*scale),
				centerY+int(39*scale),
				max(1, int(3*scale)),
				rain,
			)
		}
	case weatherSnow:
		snow := color.RGBA{R: 242, G: 249, B: 255, A: 255}
		for column := -2; column <= 2; column++ {
			x := centerX + int(float64(column*19)*scale)
			y := centerY + int(float64(29+(column&1)*8)*scale)
			drawSnowflake(target, x, y, max(2, int(5*scale)), snow)
		}
	case weatherThunder:
		bolt := color.RGBA{R: 255, G: 211, B: 74, A: 255}
		drawLine(
			target,
			centerX+int(9*scale),
			centerY+int(16*scale),
			centerX-int(5*scale),
			centerY+int(31*scale),
			max(2, int(6*scale)),
			bolt,
		)
		drawLine(
			target,
			centerX-int(5*scale),
			centerY+int(31*scale),
			centerX+int(2*scale),
			centerY+int(31*scale),
			max(2, int(5*scale)),
			bolt,
		)
		drawLine(
			target,
			centerX+int(2*scale),
			centerY+int(31*scale),
			centerX-int(12*scale),
			centerY+int(47*scale),
			max(2, int(6*scale)),
			bolt,
		)
	}
}

func skyColors(kind weatherKind, isDay bool) (color.RGBA, color.RGBA) {
	if !isDay {
		return color.RGBA{R: 18, G: 27, B: 58, A: 255},
			color.RGBA{R: 46, G: 64, B: 105, A: 255}
	}
	switch kind {
	case weatherClear:
		return color.RGBA{R: 56, G: 151, B: 236, A: 255},
			color.RGBA{R: 127, G: 207, B: 255, A: 255}
	case weatherRain, weatherThunder:
		return color.RGBA{R: 53, G: 73, B: 99, A: 255},
			color.RGBA{R: 103, G: 127, B: 150, A: 255}
	case weatherSnow, weatherFog:
		return color.RGBA{R: 112, G: 139, B: 160, A: 255},
			color.RGBA{R: 183, G: 205, B: 218, A: 255}
	default:
		return color.RGBA{R: 73, G: 119, B: 163, A: 255},
			color.RGBA{R: 133, G: 171, B: 199, A: 255}
	}
}

func drawGradient(target *image.RGBA, top, bottom color.RGBA) {
	height := target.Bounds().Dy()
	width := target.Bounds().Dx()
	for y := range height {
		ratio := float64(y) / float64(max(1, height-1))
		line := color.RGBA{
			R: uint8(float64(top.R)*(1-ratio) + float64(bottom.R)*ratio),
			G: uint8(float64(top.G)*(1-ratio) + float64(bottom.G)*ratio),
			B: uint8(float64(top.B)*(1-ratio) + float64(bottom.B)*ratio),
			A: 255,
		}
		for x := range width {
			target.SetRGBA(x, y, line)
		}
	}
}

func drawSunOrMoon(target *image.RGBA, x, y, radius int, isDay bool, sky color.RGBA) {
	if radius < 2 {
		radius = 2
	}
	if isDay {
		sun := color.RGBA{R: 255, G: 211, B: 70, A: 255}
		for ray := range 8 {
			angle := float64(ray) * math.Pi / 4
			drawLine(
				target,
				x+int(math.Cos(angle)*float64(radius+5)),
				y+int(math.Sin(angle)*float64(radius+5)),
				x+int(math.Cos(angle)*float64(radius+13)),
				y+int(math.Sin(angle)*float64(radius+13)),
				2,
				sun,
			)
		}
		fillCircle(target, x, y, radius, sun)
		return
	}
	moon := color.RGBA{R: 246, G: 242, B: 211, A: 255}
	fillCircle(target, x, y, radius, moon)
	fillCircle(target, x+radius/3, y-radius/4, radius*4/5, sky)
}

func drawCloud(target *image.RGBA, centerX, centerY int, scale float64, kind weatherKind) {
	cloud := color.RGBA{R: 235, G: 242, B: 247, A: 255}
	if kind == weatherCloudy || kind == weatherThunder {
		cloud = color.RGBA{R: 183, G: 196, B: 207, A: 255}
	}
	fillCircle(target, centerX-int(28*scale), centerY, max(2, int(18*scale)), cloud)
	fillCircle(target, centerX, centerY-int(10*scale), max(2, int(27*scale)), cloud)
	fillCircle(target, centerX+int(31*scale), centerY, max(2, int(20*scale)), cloud)
	fillRect(
		target,
		centerX-int(47*scale),
		centerY,
		centerX+int(50*scale),
		centerY+int(19*scale),
		cloud,
	)
}

func drawSnowflake(target *image.RGBA, x, y, radius int, c color.RGBA) {
	drawLine(target, x-radius, y, x+radius, y, 1, c)
	drawLine(target, x, y-radius, x, y+radius, 1, c)
	drawLine(target, x-radius, y-radius, x+radius, y+radius, 1, c)
	drawLine(target, x-radius, y+radius, x+radius, y-radius, 1, c)
}

func fillCircle(target *image.RGBA, centerX, centerY, radius int, c color.RGBA) {
	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius && image.Pt(x, y).In(target.Bounds()) {
				target.SetRGBA(x, y, c)
			}
		}
	}
}

func fillRect(target *image.RGBA, left, top, right, bottom int, c color.RGBA) {
	rect := image.Rect(left, top, right, bottom).Intersect(target.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			target.SetRGBA(x, y, c)
		}
	}
}

func drawLine(target *image.RGBA, x0, y0, x1, y1, thickness int, c color.RGBA) {
	distance := int(math.Max(math.Abs(float64(x1-x0)), math.Abs(float64(y1-y0))))
	if distance == 0 {
		fillCircle(target, x0, y0, max(1, thickness/2), c)
		return
	}
	for step := 0; step <= distance; step++ {
		ratio := float64(step) / float64(distance)
		x := int(math.Round(float64(x0) + float64(x1-x0)*ratio))
		y := int(math.Round(float64(y0) + float64(y1-y0)*ratio))
		fillCircle(target, x, y, max(1, thickness/2), c)
	}
}
