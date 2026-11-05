// Package font rasterizes text into a shared RGBA atlas and lays out glyph
// quads for the draw list. It is pure Go and must not import any SDL package.
package font

import "github.com/tacf/goak/internal/draw"

// Default and maximum atlas dimensions in pixels.
const (
	initialAtlasSize = 256
	maxAtlasSize     = 4096
	glyphPadding     = 1
)

type shelf struct{ y, h, x int }

// Atlas is an RGBA texture packed with a shelf packer. Pixels are white with
// alpha coverage so vertex color tints the glyphs.
type Atlas struct {
	w, h, maxW, maxH int
	pix              []byte
	shelves          []shelf
	nextY            int
	dirty            bool
	revision         uint64
	tex              draw.TextureID
	texW, texH       int
}

// NewAtlas creates an empty atlas of the given size capped at max.
func NewAtlas(w, h, max int) *Atlas {
	a := &Atlas{maxW: max, maxH: max}
	a.reset(w, h)
	return a
}

func (a *Atlas) reset(w, h int) {
	a.w, a.h = w, h
	a.pix = make([]byte, w*h*4)
	a.shelves = a.shelves[:0]
	a.nextY = 0
	a.dirty = true
}

// Size returns the atlas dimensions.
func (a *Atlas) Size() (w, h int) { return a.w, a.h }

// Pixels returns the RGBA pixels. The slice is owned by the atlas.
func (a *Atlas) Pixels() []byte { return a.pix }

// Revision increments whenever the atlas grows. Pixel coordinates stay valid
// but UVs normalized by the old size must be rescaled.
func (a *Atlas) Revision() uint64 { return a.revision }

// Dirty reports whether pixels changed since the last Upload.
func (a *Atlas) Dirty() bool { return a.dirty }

// Alloc reserves a w by h pixel region plus padding and returns its origin.
func (a *Atlas) Alloc(w, h int) (x, y int, ok bool) {
	cw, ch := w+glyphPadding, h+glyphPadding
	if cw > a.w || ch > a.h {
		return 0, 0, false
	}
	best := -1
	for i := range a.shelves {
		s := &a.shelves[i]
		if s.h >= ch && s.x+cw <= a.w && (best < 0 || s.h < a.shelves[best].h) {
			best = i
		}
	}
	if best < 0 {
		if a.nextY+ch > a.h {
			return 0, 0, false
		}
		a.shelves = append(a.shelves, shelf{y: a.nextY, h: ch})
		a.nextY += ch
		best = len(a.shelves) - 1
	}
	s := &a.shelves[best]
	x, y = s.x, s.y
	s.x += cw
	return x, y, true
}

// Write copies a coverage mask (one byte per pixel, row stride w) into the
// atlas at x, y as white pixels with that alpha.
func (a *Atlas) Write(x, y, w, h int, alpha []byte) {
	for row := 0; row < h; row++ {
		dst := ((y+row)*a.w + x) * 4
		src := row * w
		for col := 0; col < w; col++ {
			o := dst + col*4
			a.pix[o], a.pix[o+1], a.pix[o+2], a.pix[o+3] = 255, 255, 255, alpha[src+col]
		}
	}
	a.dirty = true
}

// grow doubles the height, then the width once the height is capped. Packed
// regions keep their pixel coordinates and pixels are copied, so only the UV
// normalization changes. Callers rescale UVs by the old over new size. False
// means the atlas is at its maximum.
func (a *Atlas) grow() bool {
	w, h := a.w, a.h
	switch {
	case h < a.maxH:
		h = min(h*2, a.maxH)
	case w < a.maxW:
		w = min(w*2, a.maxW)
	default:
		return false
	}
	pix := make([]byte, w*h*4)
	for y := 0; y < a.h; y++ {
		copy(pix[y*w*4:], a.pix[y*a.w*4:(y+1)*a.w*4])
	}
	a.w, a.h, a.pix = w, h, pix
	a.dirty = true
	a.revision++
	return true
}

// UV returns exact edge UVs for the region.
func (a *Atlas) UV(x, y, w, h int) (uvMin, uvMax [2]float32) {
	iw, ih := 1/float32(a.w), 1/float32(a.h)
	uvMin = [2]float32{float32(x) * iw, float32(y) * ih}
	uvMax = [2]float32{float32(x+w) * iw, float32(y+h) * ih}
	return
}

// Upload creates the texture on first use and updates it when dirty. It
// returns the texture id, which stays stable across updates.
func (a *Atlas) Upload(b draw.Renderer) (draw.TextureID, error) {
	if a.tex == 0 {
		id, err := b.NewTexture(a.w, a.h, a.pix)
		if err != nil {
			return 0, err
		}
		a.tex, a.texW, a.texH, a.dirty = id, a.w, a.h, false
		return id, nil
	}
	if a.dirty || a.texW != a.w || a.texH != a.h {
		if err := b.UpdateTexture(a.tex, a.w, a.h, a.pix); err != nil {
			return a.tex, err
		}
		a.texW, a.texH, a.dirty = a.w, a.h, false
	}
	return a.tex, nil
}

// Texture returns the id from the last successful Upload, or zero.
func (a *Atlas) Texture() draw.TextureID { return a.tex }
