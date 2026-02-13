// Package inspired by Clay.h ideas
package layout

// Layout runs the two-pass layout: Pass 1 resolves sizes, Pass 2 assigns positions.
// root.Bounds is set to (0, 0, viewW, viewH). Call on window resize with new viewW, viewH.
func Layout(root *Container, viewW, viewH float64) {
	pass1Size(root, viewW, viewH)
	pass2Position(root, 0, 0)
}

// pass1Size (Pass 1): resolve each node's width and height from parent-available space.
// Fills Bounds.W and Bounds.H only.
func pass1Size(c *Container, availW, availH float64) {
	w := resolveSize(c.Width, availW)
	h := resolveSize(c.Height, availH)
	c.Bounds.W = w
	c.Bounds.H = h

	contentW := w
	contentH := h
	if len(c.Children) == 0 {
		return
	}

	// Fixed/percent children claim space; Auto children share remaining.
	var autoCountW, autoCountH int
	var fixedW, fixedH float64
	for _, child := range c.Children {
		if child.Width.Kind == Auto {
			autoCountW++
		} else {
			fixedW += resolveSize(child.Width, contentW)
		}
		if child.Height.Kind == Auto {
			autoCountH++
		} else {
			fixedH += resolveSize(child.Height, contentH)
		}
	}
	remainingW := contentW - fixedW
	if remainingW < 0 {
		remainingW = 0
	}
	remainingH := contentH - fixedH
	if remainingH < 0 {
		remainingH = 0
	}
	childW := remainingW
	childH := remainingH
	if autoCountW > 0 {
		childW = remainingW / float64(autoCountW)
	}
	if autoCountH > 0 {
		childH = remainingH / float64(autoCountH)
	}

	for _, child := range c.Children {
		cw := resolveSize(child.Width, contentW)
		ch := resolveSize(child.Height, contentH)
		if child.Width.Kind == Auto {
			cw = childW
		}
		if child.Height.Kind == Auto {
			ch = childH
		}
		pass1Size(child, cw, ch)
	}
}

func resolveSize(s Size, parent float64) float64 {
	switch s.Kind {
	case Static:
		return s.Value
	case Percent:
		return parent * (s.Value / 100)
	case Auto:
		return parent
	default:
		return parent
	}
}

// pass2Position (Pass 2): assign x,y to each node. Children stacked vertically.
// Fills Bounds.X and Bounds.Y.
func pass2Position(c *Container, x, y float64) {
	c.Bounds.X = x
	c.Bounds.Y = y

	var totalChildH float64
	for _, child := range c.Children {
		totalChildH += child.Bounds.H
	}

	cy := y
	switch c.VerticalAlign {
	case AlignCenter:
		cy = y + (c.Bounds.H-totalChildH)/2
	case AlignEnd:
		cy = y + (c.Bounds.H - totalChildH)
	}

	for _, child := range c.Children {
		cx := x
		switch c.HorizontalAlign {
		case AlignCenter:
			cx = x + (c.Bounds.W-child.Bounds.W)/2
		case AlignEnd:
			cx = x + (c.Bounds.W - child.Bounds.W)
		}
		child.Bounds.X = cx
		child.Bounds.Y = cy
		pass2Position(child, cx, cy)
		cy += child.Bounds.H
	}
}
