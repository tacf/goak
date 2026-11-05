package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
)

// Image draws a texture region. Pass colors.White as tint for no tint.
func (c *Context) Image(tex draw.TextureID, size, uv0, uv1 Vec2, tint colors.Color) {
	r := c.ItemAdd(size)
	if c.cur.skip {
		return
	}
	c.cur.DL.AddImage(tex, r.Min, r.Max, [2]float32{uv0.X, uv0.Y}, [2]float32{uv1.X, uv1.Y}, tint)
}

// ProgressBar draws frac in 0..1. A size component of 0 means the available
// width or the frame height. An empty overlay shows no text.
func (c *Context) ProgressBar(frac float32, size Vec2, overlay string) {
	if size.X <= 0 {
		size.X = max(c.ContentRegionAvail().X+size.X, 1)
	}
	if size.Y <= 0 {
		size.Y = c.frameHeight()
	}
	r := c.ItemAdd(size)
	if c.cur.skip {
		return
	}
	frac = min(max(frac, 0), 1)
	dl := c.cur.DL
	dl.AddRectFilled(r.Min, r.Max, c.col(ColFrameBg), c.Style.FrameRounding)
	if frac > 0 {
		dl.PushClipRect(Rect{Min: r.Min, Max: Vec2{X: r.Min.X + size.X*frac, Y: r.Max.Y}})
		dl.AddRectFilled(r.Min, r.Max, c.col(ColProgressBar), c.Style.FrameRounding)
		dl.PopClipRect()
	}
	dl.AddRect(r.Min, r.Max, c.col(ColBorder), c.Style.FrameRounding, 1)
	if overlay != "" {
		ts := c.TextSize(overlay)
		c.DrawText(Vec2{X: r.Min.X + (size.X-ts.X)/2, Y: r.Min.Y + (size.Y-ts.Y)/2}, overlay, c.col(ColText))
	}
}
