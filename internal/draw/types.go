// Package draw is a pure Go draw list and backend interface for the
// immediate-mode UI. It must not import any SDL package.
package draw

// TextureID identifies a backend texture. Zero means no texture.
type TextureID uint64

// Vec2 is a 2D point or size.
type Vec2 struct{ X, Y float32 }

// Rect is an axis aligned rectangle given by its Min and Max corners.
type Rect struct{ Min, Max Vec2 }

// Intersect returns the overlap of r and o. The result may be empty.
func (r Rect) Intersect(o Rect) Rect {
	out := Rect{
		Min: Vec2{max(r.Min.X, o.Min.X), max(r.Min.Y, o.Min.Y)},
		Max: Vec2{min(r.Max.X, o.Max.X), min(r.Max.Y, o.Max.Y)},
	}
	if out.Max.X < out.Min.X {
		out.Max.X = out.Min.X
	}
	if out.Max.Y < out.Min.Y {
		out.Max.Y = out.Min.Y
	}
	return out
}

// Empty reports whether the rectangle has no area.
func (r Rect) Empty() bool { return r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y }

// Vertex is one draw list vertex. Color is packed with R in the low byte.
type Vertex struct {
	Pos   Vec2
	UV    [2]float32
	Color uint32
}

// Cmd is a batch of triangles sharing clip rect and texture.
// Indices are absolute into DrawList.Vertices, so VtxOffset is 0 unless a
// backend splits the list.
type Cmd struct {
	ClipRect  Rect
	TextureID TextureID
	IdxOffset int
	IdxCount  int
	VtxOffset int
	// Quads marks a Cmd made only of axis aligned quads. Each quad is 4
	// vertices (min, top right, max, bottom left) with 6 indices
	// a,a+1,a+2,a,a+2,a+3 and one color. Backends may draw it faster, and
	// may ignore the flag. False is always valid.
	Quads bool
}

// GlyphQuad is a pre-laid-out glyph rectangle with atlas UVs.
type GlyphQuad struct {
	Min, Max     Vec2
	UVMin, UVMax [2]float32
}

// DrawData is everything a backend needs to render one frame.
type DrawData struct {
	Lists       []*DrawList
	DisplaySize Vec2
	Scale       Vec2
}

// Renderer renders draw data and owns GPU or software textures.
type Renderer interface {
	// NewTexture uploads RGBA32 pixels and returns its id.
	NewTexture(w, h int, rgba []byte) (TextureID, error)
	// UpdateTexture replaces the pixels of an existing texture.
	UpdateTexture(id TextureID, w, h int, rgba []byte) error
	// DestroyTexture releases a texture.
	DestroyTexture(id TextureID)
	// Render draws one frame of draw data. Vertex positions are in logical
	// units and DrawData.Scale converts them to pixels.
	Render(*DrawData) error
}

// Add returns v + o.
func (v Vec2) Add(o Vec2) Vec2 { return Vec2{X: v.X + o.X, Y: v.Y + o.Y} }

// Sub returns v - o.
func (v Vec2) Sub(o Vec2) Vec2 { return Vec2{X: v.X - o.X, Y: v.Y - o.Y} }

// Max returns the component wise maximum of v and o.
func (v Vec2) Max(o Vec2) Vec2 { return Vec2{X: max(v.X, o.X), Y: max(v.Y, o.Y)} }

// Contains reports whether p is inside r. Empty rects contain nothing and the
// Max edges are exclusive.
func (r Rect) Contains(p Vec2) bool {
	return !r.Empty() && p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}
