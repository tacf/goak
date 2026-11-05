package goak

// Options configures NewContext.
type Options struct {
	// FontData is TTF or OTF data. Nil uses the embedded font.
	FontData []byte
	// FontSize is the font size in logical pixels. Zero means 18.
	FontSize float64
}

// NewContext returns a context bound to b. It builds the font at the
// backend scale, uses the backend clipboard and follows scale changes in
// NewFrame.
func NewContext(b Backend, o Options) (*Context, error) {
	if o.FontSize <= 0 {
		o.FontSize = 18
	}
	f, err := NewFont(o.FontData, o.FontSize, b.Scale())
	if err != nil {
		return nil, err
	}
	c := New(f, b)
	c.Clipboard = b.Clipboard()
	c.platform = b
	c.fontData, c.fontPx = o.FontData, o.FontSize
	return c, nil
}

// Present finishes the frame, clears the target to bg, draws it and shows
// it. It needs a context made by NewContext.
func (c *Context) Present(bg Color) error {
	dd := c.Render()
	if c.platform == nil {
		return nil
	}
	if err := c.platform.Clear(bg); err != nil {
		return err
	}
	if err := c.platform.Render(dd); err != nil {
		return err
	}
	return c.platform.Present()
}
