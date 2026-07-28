package rendering

import (
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"golang.org/x/image/font/gofont/goregular"
)

type glyph struct {
	texture *sdl.Texture
	width   float32
	height  float32
	advance float32
}

// Font is an SDL_ttf font backed by the Go regular font embedded in the
// executable. Glyphs are rasterized lazily and cached as SDL textures.
type Font struct {
	handle *ttf.Font
	render *sdl.Renderer
	height float64
	space  float32
	scale  float32
	glyphs map[rune]*glyph
}

// NewFont opens the embedded default font at the requested logical point size.
// rasterScale is the logical-to-device-pixel scale. Glyph textures are
// rasterized at their final physical size while their metrics remain logical.
func NewFont(renderer *sdl.Renderer, size, rasterScale float32) (*Font, error) {
	if rasterScale <= 0 || math.IsNaN(float64(rasterScale)) || math.IsInf(float64(rasterScale), 0) {
		rasterScale = 1
	}
	stream, err := sdl.IOFromBytes(goregular.TTF)
	if err != nil {
		return nil, err
	}
	handle, err := ttf.OpenFontIO(stream, true, max(1, size*rasterScale))
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	font := &Font{
		handle: handle,
		render: renderer,
		height: float64(handle.Height()) / float64(rasterScale),
		scale:  rasterScale,
		glyphs: make(map[rune]*glyph),
	}
	if metrics, metricErr := handle.GlyphMetrics(uint32(' ')); metricErr == nil &&
		metrics != nil && metrics.Advance > 0 {
		font.space = float32(metrics.Advance) / rasterScale
	} else {
		font.space = float32(font.height * 0.5)
	}
	return font, nil
}

// Close releases the cached SDL textures and the SDL_ttf font.
func (f *Font) Close() {
	if f == nil {
		return
	}
	for _, glyph := range f.glyphs {
		if glyph.texture != nil {
			glyph.texture.Destroy()
		}
	}
	f.glyphs = nil
	if f.handle != nil {
		f.handle.Close()
		f.handle = nil
	}
}

// Height returns the font's recommended line height.
func (f *Font) Height() float64 {
	if f == nil {
		return 0
	}
	return f.height
}

// Measure returns the dimensions of a single-line string.
func (f *Font) Measure(text string) (float64, float64) {
	if f == nil {
		return 0, 0
	}
	var width float32
	for _, r := range text {
		width += f.advanceAt(r, width)
	}
	return float64(width), f.height
}

func (f *Font) advanceAt(r rune, offset float32) float32 {
	if r == '\t' {
		stop := f.space * 4
		if stop <= 0 {
			return f.space
		}
		advance := stop - float32(math.Mod(float64(offset), float64(stop)))
		if advance < 0.01 {
			return stop
		}
		return advance
	}
	return f.glyph(r).advance
}

func (f *Font) glyph(r rune) *glyph {
	if cached, ok := f.glyphs[r]; ok {
		return cached
	}
	result := &glyph{advance: f.space}
	if metrics, err := f.handle.GlyphMetrics(uint32(r)); err == nil && metrics != nil {
		result.advance = float32(metrics.Advance) / f.scale
	}
	if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
		surface, err := f.handle.RenderTextBlended(string(r), sdl.Color{
			R: 255, G: 255, B: 255, A: 255,
		})
		if err == nil && surface != nil {
			if texture, textureErr := f.render.CreateTextureFromSurface(surface); textureErr == nil && texture != nil {
				_ = texture.SetBlendMode(sdl.BLENDMODE_BLEND)
				_ = texture.SetScaleMode(sdl.SCALEMODE_NEAREST)
				result.texture = texture
				result.width = float32(texture.W) / f.scale
				result.height = float32(texture.H) / f.scale
			}
			surface.Destroy()
		}
	}
	f.glyphs[r] = result
	return result
}
