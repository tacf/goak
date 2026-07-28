package rendering

import (
	"math"

	"goak/internal/goak/colors"
	"goak/internal/goak/layout"

	"github.com/Zyko0/go-sdl3/sdl"
)

// FillRect draws a filled rectangle.
func FillRect(renderer *sdl.Renderer, x, y, w, h float64, c colors.Color) {
	if renderer == nil || w <= 0 || h <= 0 {
		return
	}
	_ = renderer.SetDrawColor(c.R, c.G, c.B, c.A)
	_ = renderer.RenderFillRect(&sdl.FRect{
		X: float32(x), Y: float32(y), W: float32(w), H: float32(h),
	})
}

// DrawStrokeRect draws a rectangular outline with the given thickness.
func DrawStrokeRect(renderer *sdl.Renderer, x, y, w, h, thickness float64, c colors.Color) {
	if renderer == nil || w <= 0 || h <= 0 || thickness <= 0 {
		return
	}
	if thickness*2 >= w || thickness*2 >= h {
		FillRect(renderer, x, y, w, h, c)
		return
	}
	FillRect(renderer, x, y, w, thickness, c)
	FillRect(renderer, x, y+h-thickness, w, thickness, c)
	FillRect(renderer, x, y+thickness, thickness, h-thickness*2, c)
	FillRect(renderer, x+w-thickness, y+thickness, thickness, h-thickness*2, c)
}

// DrawLine draws a horizontal or vertical line.
func DrawLine(renderer *sdl.Renderer, x, y, length, thickness float64, c colors.Color, horizontal bool) {
	if horizontal {
		FillRect(renderer, x, y, length, thickness, c)
		return
	}
	FillRect(renderer, x, y, thickness, length, c)
}

// DrawCircleStroke draws an antialiased circle outline.
func DrawCircleStroke(renderer *sdl.Renderer, centerX, centerY, radius, thickness float64, c colors.Color) {
	if renderer == nil || radius <= 0 || thickness <= 0 {
		return
	}
	if thickness >= radius {
		DrawFilledCircle(renderer, centerX, centerY, radius, c)
		return
	}

	aa := circleAAWidth(renderer)
	inner := radius - thickness
	innerTransparent := max(0, inner-aa*0.5)
	innerOpaque := inner + aa*0.5
	outerOpaque := radius - aa*0.5
	outerTransparent := radius + aa*0.5
	opacity := float32(1)
	if innerOpaque > outerOpaque {
		innerOpaque = (inner + radius) * 0.5
		outerOpaque = innerOpaque
		opacity = float32(min(1, thickness/aa))
	}

	segments := circleSegmentCount(radius, renderer)
	opaque := floatColor(c, opacity)
	transparent := floatColor(c, 0)
	radii := [...]float64{innerTransparent, innerOpaque, outerOpaque, outerTransparent}
	ringColors := [...]sdl.FColor{transparent, opaque, opaque, transparent}
	vertices := make([]sdl.Vertex, 0, segments*len(radii))
	for ring, ringRadius := range radii {
		vertices = appendCircleRing(vertices, centerX, centerY, ringRadius, segments, ringColors[ring])
	}

	indices := make([]int32, 0, segments*6*(len(radii)-1))
	for ring := 0; ring < len(radii)-1; ring++ {
		indices = appendRingTriangles(indices, ring*segments, (ring+1)*segments, segments)
	}
	_ = renderer.RenderGeometry(nil, vertices, indices)
}

// DrawFilledCircle draws an antialiased filled circle.
func DrawFilledCircle(renderer *sdl.Renderer, centerX, centerY, radius float64, c colors.Color) {
	if renderer == nil || radius <= 0 {
		return
	}

	aa := circleAAWidth(renderer)
	innerRadius := max(0, radius-aa*0.5)
	outerRadius := radius + aa*0.5
	segments := circleSegmentCount(radius, renderer)
	opaque := floatColor(c, 1)
	transparent := floatColor(c, 0)

	vertices := make([]sdl.Vertex, 1, 1+segments*2)
	vertices[0] = sdl.Vertex{
		Position: sdl.FPoint{X: float32(centerX), Y: float32(centerY)},
		Color:    opaque,
	}
	vertices = appendCircleRing(vertices, centerX, centerY, innerRadius, segments, opaque)
	vertices = appendCircleRing(vertices, centerX, centerY, outerRadius, segments, transparent)

	indices := make([]int32, 0, segments*9)
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		indices = append(indices, 0, int32(1+i), int32(1+next))
	}
	indices = appendRingTriangles(indices, 1, 1+segments, segments)
	_ = renderer.RenderGeometry(nil, vertices, indices)
}

func circleAAWidth(renderer *sdl.Renderer) float64 {
	scale := float64(1)
	if scaleX, scaleY, err := renderer.Scale(); err == nil {
		scale = max(math.Abs(float64(scaleX)), math.Abs(float64(scaleY)))
	}
	if scale <= 0 {
		scale = 1
	}
	return 1 / scale
}

func circleSegmentCount(radius float64, renderer *sdl.Renderer) int {
	scale := float64(1)
	if scaleX, scaleY, err := renderer.Scale(); err == nil {
		scale = max(math.Abs(float64(scaleX)), math.Abs(float64(scaleY)))
	}
	physicalRadius := max(1, radius*scale)
	return min(256, max(24, int(math.Ceil(2*math.Pi*math.Sqrt(physicalRadius)))))
}

func floatColor(c colors.Color, opacity float32) sdl.FColor {
	const toFloat = 1.0 / 255.0
	return sdl.FColor{
		R: float32(c.R) * toFloat,
		G: float32(c.G) * toFloat,
		B: float32(c.B) * toFloat,
		A: float32(c.A) * toFloat * opacity,
	}
}

func appendCircleRing(vertices []sdl.Vertex, centerX, centerY, radius float64, segments int, color sdl.FColor) []sdl.Vertex {
	for i := 0; i < segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		vertices = append(vertices, sdl.Vertex{
			Position: sdl.FPoint{
				X: float32(centerX + math.Cos(angle)*radius),
				Y: float32(centerY + math.Sin(angle)*radius),
			},
			Color: color,
		})
	}
	return vertices
}

func appendRingTriangles(indices []int32, innerStart, outerStart, segments int) []int32 {
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		inner := int32(innerStart + i)
		innerNext := int32(innerStart + next)
		outer := int32(outerStart + i)
		outerNext := int32(outerStart + next)
		indices = append(indices,
			inner, outer, outerNext,
			inner, outerNext, innerNext,
		)
	}
	return indices
}

// DrawText renders text at the specified top-left position.
func DrawText(renderer *sdl.Renderer, str string, font *Font, x, y float64, c colors.Color) {
	if renderer == nil || font == nil || str == "" {
		return
	}
	pen := float32(x)
	origin := pen
	for _, r := range str {
		glyph := font.glyph(r)
		if glyph.texture != nil {
			_ = glyph.texture.SetColorMod(c.R, c.G, c.B)
			_ = glyph.texture.SetAlphaMod(c.A)
			_ = renderer.RenderTexture(glyph.texture, nil, &sdl.FRect{
				X: pen, Y: float32(y), W: glyph.width, H: glyph.height,
			})
		}
		pen += font.advanceAt(r, pen-origin)
	}
}

// PointWithinBounds returns true if the point (x, y) is inside the given rectangle.
func PointWithinBounds(x, y float64, r layout.Rect) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}
