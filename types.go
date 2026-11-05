package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
	"github.com/tacf/goak/internal/font"
)

// Color is an RGBA color. See package colors for a palette.
type Color = colors.Color

// Types a Backend consumes. Context.Render produces them.
type (
	// DrawData is everything a backend needs to render one frame.
	DrawData = draw.DrawData
	// DrawList holds the vertices and commands of one window.
	DrawList = draw.DrawList
	// Cmd is a batch of triangles sharing clip rect and texture.
	Cmd = draw.Cmd
	// Vertex is one draw list vertex.
	Vertex = draw.Vertex
	// TextureID identifies a backend texture.
	TextureID = draw.TextureID
	// GlyphQuad is a laid out glyph rectangle with atlas UVs.
	GlyphQuad = draw.GlyphQuad
	// Font is a TTF or OTF font rasterized into a texture atlas.
	Font = font.Font
)

// NewDrawList returns an empty draw list for custom drawing.
func NewDrawList() *DrawList { return draw.NewDrawList() }

// NewFont builds a font from TTF or OTF data (nil for the embedded one) at
// sizePx logical pixels for the given pixel scale.
func NewFont(data []byte, sizePx float64, scale float32) (*Font, error) {
	return font.New(data, sizePx, float64(scale))
}

// Corner selects rounded corners of a rect.
type Corner = draw.Corner

// Corner flags.
const (
	CornerTopLeft     = draw.CornerTopLeft
	CornerTopRight    = draw.CornerTopRight
	CornerBottomRight = draw.CornerBottomRight
	CornerBottomLeft  = draw.CornerBottomLeft
	CornerAll         = draw.CornerAll
)

// Pack converts a color to the packed Vertex.Color format.
func Pack(c Color) uint32 { return draw.Pack(c) }

// Unpack converts a packed Vertex.Color back to a Color.
func Unpack(p uint32) Color { return draw.Unpack(p) }

// FrameSkipper reports whether draw data changed since the last frame.
type FrameSkipper = draw.FrameSkipper
