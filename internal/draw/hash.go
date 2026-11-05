package draw

import (
	"math"
	"time"
)

// Hash returns a 64 bit FNV-1a style hash of everything that affects the
// pixels of dd. Equal data gives equal hashes.
func (dd *DrawData) Hash() uint64 {
	const prime = 1099511628211
	h := uint64(14695981039346656037)
	mix := func(v uint32) { h = (h ^ uint64(v)) * prime }
	mixF := func(f float32) { mix(math.Float32bits(f)) }
	mixF(dd.DisplaySize.X)
	mixF(dd.DisplaySize.Y)
	mixF(dd.Scale.X)
	mixF(dd.Scale.Y)
	for _, l := range dd.Lists {
		mix(uint32(len(l.Vertices)))
		for _, v := range l.Vertices {
			mixF(v.Pos.X)
			mixF(v.Pos.Y)
			mixF(v.UV[0])
			mixF(v.UV[1])
			mix(v.Color)
		}
		for _, i := range l.Indices {
			mix(i)
		}
		for _, c := range l.Cmds {
			mixF(c.ClipRect.Min.X)
			mixF(c.ClipRect.Min.Y)
			mixF(c.ClipRect.Max.X)
			mixF(c.ClipRect.Max.Y)
			mix(uint32(c.TextureID))
			mix(uint32(c.TextureID >> 32))
			mix(uint32(c.IdxOffset))
			mix(uint32(c.IdxCount))
			mix(uint32(c.VtxOffset))
		}
	}
	return h
}

// FrameSkipper tells whether a frame needs drawing. It is optional and the
// zero value is ready to use.
//
//	if skip.Changed(dd) {
//		b.Clear(bg)
//		b.RenderDrawData(dd)
//		b.Present()
//	}
type FrameSkipper struct {
	// Interval forces a redraw after this long. Zero means one second.
	Interval time.Duration

	hash uint64
	at   time.Time
	has  bool
}

// Changed reports whether dd differs from the last frame it returned true
// for, or whether Interval has elapsed since then.
func (s *FrameSkipper) Changed(dd *DrawData) bool {
	iv := s.Interval
	if iv <= 0 {
		iv = time.Second
	}
	h, now := dd.Hash(), time.Now()
	if s.has && h == s.hash && now.Sub(s.at) < iv {
		return false
	}
	s.hash, s.at, s.has = h, now, true
	return true
}
