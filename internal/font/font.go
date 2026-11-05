package font

import (
	"fmt"
	"image"
	"math"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type glyph struct {
	advance      float32 // logical
	x, y, w, h   int     // atlas region in pixels, w or h zero when blank
	offX, offY   float32 // logical offset from pen and baseline
	sizeW, sizeH float32 // logical quad size
}

// Font rasterizes glyphs lazily into an atlas at size*scale pixels while all
// metrics stay logical. It is not safe for concurrent use.
type Font struct {
	face    font.Face
	scale   float32
	height  float32
	ascent  float32
	space   float32
	atlas   *Atlas
	glyphs  map[rune]*glyph
	scratch []draw.GlyphQuad
	spans   []span
	mask    []byte
}

// New opens a font from TTF or OTF data. Nil or empty data selects the
// embedded Go Regular. sizePx is the logical pixel size and scale is the
// logical to device pixel factor.
func New(data []byte, sizePx, scale float64) (*Font, error) {
	if len(data) == 0 {
		data = goregular.TTF
	}
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	if sizePx <= 0 || math.IsNaN(sizePx) || math.IsInf(sizePx, 0) {
		return nil, fmt.Errorf("font: invalid size %v", sizePx)
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("font: parse: %w", err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: max(1, sizePx*scale), DPI: 72, Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, fmt.Errorf("font: face: %w", err)
	}
	m := face.Metrics()
	f := &Font{
		face:   face,
		scale:  float32(scale),
		height: float32(m.Height) / 64 / float32(scale),
		ascent: float32(m.Ascent) / 64 / float32(scale),
		atlas:  NewAtlas(initialAtlasSize, initialAtlasSize, maxAtlasSize),
		glyphs: make(map[rune]*glyph),
	}
	if adv, ok := face.GlyphAdvance(' '); ok && adv > 0 {
		f.space = float32(adv) / 64 / f.scale
	} else {
		f.space = f.height * 0.5
	}
	return f, nil
}

// Scale returns the logical to device pixel factor the font was built for.
func (f *Font) Scale() float32 { return f.scale }

// Close releases the face.
func (f *Font) Close() {
	if f != nil && f.face != nil {
		_ = f.face.Close()
		f.face = nil
	}
}

// Atlas returns the glyph atlas.
func (f *Font) Atlas() *Atlas { return f.atlas }

// LineHeight returns the recommended line height.
func (f *Font) LineHeight() float64 {
	if f == nil {
		return 0
	}
	return float64(f.height)
}

// Measure returns the size of a single line string.
func (f *Font) Measure(s string) (w, h float64) {
	if f == nil {
		return 0, 0
	}
	var width float32
	for _, r := range s {
		width += f.advanceAt(r, width)
	}
	return float64(width), float64(f.height)
}

func (f *Font) advanceAt(r rune, offset float32) float32 {
	if r == '\t' {
		stop := f.space * 4
		if stop <= 0 {
			return f.space
		}
		adv := stop - float32(math.Mod(float64(offset), float64(stop)))
		if adv < 0.01 {
			return stop
		}
		return adv
	}
	return f.advanceOf(r)
}

// advanceOf returns the logical advance without touching the atlas.
func (f *Font) advanceOf(r rune) float32 {
	if r == '\n' || r == '\r' {
		return 0
	}
	if g, ok := f.glyphs[r]; ok {
		return g.advance
	}
	adv, ok := f.face.GlyphAdvance(r)
	if !ok {
		adv, ok = f.face.GlyphAdvance('?')
	}
	if !ok {
		return f.space
	}
	return float32(adv) / 64 / f.scale
}

// resolve maps a rune to the rune actually rasterized. Runes without a glyph
// fall back to '?'.
func (f *Font) resolve(r rune) rune {
	if _, ok := f.face.GlyphAdvance(r); ok {
		return r
	}
	return '?'
}

func (f *Font) glyph(r rune) *glyph {
	if g, ok := f.glyphs[r]; ok {
		return g
	}
	blank := r == ' ' || r == '\t' || r == '\n' || r == '\r'
	if !blank {
		// Share the fallback glyph so it is rasterized once.
		if rr := f.resolve(r); rr != r {
			g := f.glyph(rr)
			f.glyphs[r] = g
			return g
		}
	}
	g := &glyph{advance: f.advanceOf(r)}
	f.glyphs[r] = g
	if blank {
		return g
	}
	f.place(r, g)
	return g
}

// place rasterizes r into the atlas, growing and repacking when full.
func (f *Font) place(r rune, g *glyph) {
	dr, mask, maskp, ok := f.raster(r)
	if !ok {
		return
	}
	w, h := dr.Dx(), dr.Dy()
	g.offX, g.offY = float32(dr.Min.X)/f.scale, float32(dr.Min.Y)/f.scale
	g.sizeW, g.sizeH = float32(w)/f.scale, float32(h)/f.scale
	if w == 0 || h == 0 {
		return
	}
	for {
		x, y, ok := f.atlas.Alloc(w, h)
		if ok {
			g.x, g.y, g.w, g.h = x, y, w, h
			f.copyMask(x, y, w, h, dr, mask, maskp)
			return
		}
		if !f.atlas.grow() {
			g.w, g.h = 0, 0
			return
		}
	}
}

func (f *Font) raster(r rune) (image.Rectangle, image.Image, image.Point, bool) {
	dr, mask, maskp, _, ok := f.face.Glyph(fixed.Point26_6{}, f.resolve(r))
	return dr, mask, maskp, ok
}

func (f *Font) copyMask(x, y, w, h int, dr image.Rectangle, mask image.Image, maskp image.Point) {
	if cap(f.mask) < w*h {
		f.mask = make([]byte, w*h)
	}
	buf := f.mask[:w*h]
	if a, ok := mask.(*image.Alpha); ok {
		for row := 0; row < h; row++ {
			o := a.PixOffset(maskp.X, maskp.Y+row)
			copy(buf[row*w:(row+1)*w], a.Pix[o:o+w])
		}
	} else {
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				_, _, _, av := mask.At(maskp.X+col, maskp.Y+row).RGBA()
				buf[row*w+col] = byte(av >> 8)
			}
		}
	}
	f.atlas.Write(x, y, w, h, buf)
}

// Layout returns one quad per visible glyph for s with the top left of the
// line at x, y in logical units.
func (f *Font) Layout(s string, x, y float32) []draw.GlyphQuad {
	return f.LayoutAppend(nil, s, x, y)
}

// LayoutAppend appends the quads of s to dst. Tabs advance to the next stop
// of four spaces like rendering.Font. Newline and carriage return have no
// advance and no glyph. UVs reflect the atlas size on return.
func (f *Font) LayoutAppend(dst []draw.GlyphQuad, s string, x, y float32) []draw.GlyphQuad {
	if f == nil {
		return dst
	}
	return f.layoutOnce(dst, s, x, y)
}

func (f *Font) layoutOnce(dst []draw.GlyphQuad, s string, x, y float32) []draw.GlyphQuad {
	sc := f.scale
	baseline := float32(math.Round(float64((y+f.ascent)*sc))) / sc
	base := len(dst)
	w0, h0 := f.atlas.Size()
	var pen float32
	for _, r := range s {
		g := f.glyph(r)
		// Growth keeps pixel rects, so rescale quads made at the old size.
		if w1, h1 := f.atlas.Size(); w1 != w0 || h1 != h0 {
			sx, sy := float32(w0)/float32(w1), float32(h0)/float32(h1)
			for i := base; i < len(dst); i++ {
				q := &dst[i]
				q.UVMin[0] *= sx
				q.UVMax[0] *= sx
				q.UVMin[1] *= sy
				q.UVMax[1] *= sy
			}
			w0, h0 = w1, h1
		}
		if g.w > 0 {
			px := float32(math.Round(float64((x+pen)*sc))) / sc
			q := draw.GlyphQuad{
				Min: draw.Vec2{X: px + g.offX, Y: baseline + g.offY},
			}
			q.Max = draw.Vec2{X: q.Min.X + g.sizeW, Y: q.Min.Y + g.sizeH}
			q.UVMin, q.UVMax = f.atlas.UV(g.x, g.y, g.w, g.h)
			dst = append(dst, q)
		}
		pen += f.advanceAt(r, pen)
	}
	return dst
}

// span is a run of text vertices in a draw list with the atlas size their
// UVs were normalized by.
type span struct {
	dl         *draw.DrawList
	start, end int
	w, h       int
}

// DrawText lays out s at x, y, uploads the atlas through b when needed and
// appends the glyphs to dl tinted by c. Call Finish(dl) before rendering so
// text drawn before an atlas growth gets current UVs.
func (f *Font) DrawText(dl *draw.DrawList, b draw.Renderer, x, y float32, s string, c colors.Color) error {
	if f == nil || s == "" {
		return nil
	}
	f.scratch = f.LayoutAppend(f.scratch[:0], s, x, y)
	tex, err := f.atlas.Upload(b)
	if err != nil {
		return err
	}
	if len(f.spans) > 1024 {
		f.spans = f.spans[:0]
	}
	start := len(dl.Vertices)
	dl.AddText(tex, f.scratch, c)
	if end := len(dl.Vertices); end > start {
		w, h := f.atlas.Size()
		f.spans = append(f.spans, span{dl, start, end, w, h})
	}
	return nil
}

// Finish renormalizes the UVs of text previously drawn into dl if the atlas
// grew since, then forgets dl. Safe to call when nothing changed.
func (f *Font) Finish(dl *draw.DrawList) {
	if f == nil {
		return
	}
	w, h := f.atlas.Size()
	keep := f.spans[:0]
	for _, sp := range f.spans {
		if sp.dl != dl {
			// Drop spans of lists that were reset.
			if sp.end <= len(sp.dl.Vertices) {
				keep = append(keep, sp)
			}
			continue
		}
		if sp.end > len(dl.Vertices) || (sp.w == w && sp.h == h) {
			continue
		}
		sx, sy := float32(sp.w)/float32(w), float32(sp.h)/float32(h)
		for i := sp.start; i < sp.end; i++ {
			dl.Vertices[i].UV[0] *= sx
			dl.Vertices[i].UV[1] *= sy
		}
	}
	f.spans = keep
}
