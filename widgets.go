package goak

func (c *Context) frameColor(s ButtonState, id ID, base, hov, act Col) Col {
	switch {
	case s.Held && c.activeID == id:
		return act
	case s.Hovered:
		return hov
	}
	return base
}

func (c *Context) frameHeight() float32 { return c.lineHeight() + 2*c.Style.FramePadding.Y }
