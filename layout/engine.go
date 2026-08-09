// Package inspired by Clay.h ideas
package layout

// Layout runs the two-pass layout: Pass 1 resolves sizes, Pass 2 assigns positions.
// root.Bounds is set to (0, 0, viewW, viewH). Call on window resize with new viewW, viewH.
func Layout(root *Container, viewW, viewH float64) {
	if root == nil {
		return
	}
	pass1Size(root, nonnegativeFinite(viewW), nonnegativeFinite(viewH))
	pass2Position(root, 0, 0)
}

// pass1Size (Pass 1): resolve each node's width and height from parent-available space.
// Fills Bounds.W and Bounds.H only.
func pass1Size(c *Container, availW, availH float64) {
	w := resolveSize(c.Width, availW)
	h := resolveSize(c.Height, availH)
	pass1Resolved(c, w, h)
}

// pass1Resolved lays out a node whose size has already been resolved by its
// parent. Keeping this separate prevents percentage sizes from being applied
// a second time during recursion.
func pass1Resolved(c *Container, w, h float64) {
	c.Bounds.W = w
	c.Bounds.H = h

	_, _, contentW, contentH := contentBox(c, w, h)
	if len(c.Children) == 0 {
		return
	}

	if directionOf(c) == Row {
		resolveMainAxis(c.Children, contentW, gaps(c), true)
		for _, child := range c.Children {
			child.Bounds.H = resolveSize(child.Height, contentH)
		}
	} else {
		resolveMainAxis(c.Children, contentH, gaps(c), false)
		for _, child := range c.Children {
			child.Bounds.W = resolveSize(child.Width, contentW)
		}
	}

	for _, child := range c.Children {
		pass1Resolved(child, child.Bounds.W, child.Bounds.H)
	}
}

// resolveMainAxis preserves fixed and percentage dimensions, then shares the
// remaining space between Auto dimensions while respecting constraints. It
// writes directly to child bounds and performs no per-pass allocations.
func resolveMainAxis(children []*Container, available, reserved float64, horizontal bool) {
	fixed := max(0, reserved)
	active := 0
	for _, child := range children {
		size := axisSize(child, horizontal)
		if size.Kind != Auto {
			resolved := resolveSize(size, available)
			setAxisBounds(child, horizontal, resolved)
			fixed += resolved
			continue
		}
		minimum, maximum := sizeLimits(size)
		setAxisBounds(child, horizontal, minimum)
		fixed += minimum
		if maximum <= 0 || minimum < maximum {
			active++
		}
	}

	remaining := max(0, available-fixed)
	if remaining <= 0 {
		return
	}

	for remaining > 1e-9 && active > 0 {
		share := remaining / float64(active)
		nextActive := 0
		used := 0.0
		for _, child := range children {
			size := axisSize(child, horizontal)
			if size.Kind != Auto {
				continue
			}
			current := axisBounds(child, horizontal)
			_, maximum := sizeLimits(size)
			if maximum > 0 && current >= maximum {
				continue
			}
			addition := share
			if maximum > 0 && current+addition >= maximum {
				addition = max(0, maximum-current)
			}
			current += addition
			setAxisBounds(child, horizontal, current)
			used += addition
			if maximum <= 0 || current < maximum {
				nextActive++
			}
		}
		if used <= 1e-9 {
			break
		}
		remaining -= used
		active = nextActive
	}
}

func axisSize(child *Container, horizontal bool) Size {
	if horizontal {
		return child.Width
	}
	return child.Height
}

func axisBounds(child *Container, horizontal bool) float64 {
	if horizontal {
		return child.Bounds.W
	}
	return child.Bounds.H
}

func setAxisBounds(child *Container, horizontal bool, value float64) {
	if horizontal {
		child.Bounds.W = value
		return
	}
	child.Bounds.H = value
}

func resolveSize(size Size, parent float64) float64 {
	value := parent
	switch size.Kind {
	case Static:
		value = nonnegativeFinite(size.Value)
	case Percent:
		value = parent * (min(100, nonnegativeFinite(size.Value)) / 100)
	case Auto:
		// Auto fills its available space when it is resolved independently.
	default:
		// Unknown sizing modes retain the historical fill behavior.
	}
	return constrain(value, size)
}

func constrain(value float64, size Size) float64 {
	minimum, maximum := sizeLimits(size)
	if minimum > 0 {
		value = max(value, minimum)
	}
	if maximum > 0 {
		value = min(value, maximum)
	}
	return value
}

func sizeLimits(size Size) (minimum, maximum float64) {
	minimum = nonnegativeFinite(size.Min)
	maximum = nonnegativeFinite(size.Max)
	if maximum > 0 && maximum < minimum {
		maximum = minimum
	}
	return minimum, maximum
}

// pass2Position (Pass 2): assign x,y to each node according to its direction.
// Fills Bounds.X and Bounds.Y.
func pass2Position(c *Container, x, y float64) {
	c.Bounds.X = x
	c.Bounds.Y = y

	insetX, insetY, contentW, contentH := contentBox(c, c.Bounds.W, c.Bounds.H)
	contentX := x + insetX
	contentY := y + insetY
	gap := nonnegativeFinite(c.Gap)

	if directionOf(c) == Row {
		totalWidth := totalMainSize(c.Children, gap, func(child *Container) float64 {
			return child.Bounds.W
		})
		cx := alignedStart(contentX, contentW, totalWidth, c.HorizontalAlign)
		for _, child := range c.Children {
			cy := alignItem(contentY, contentH, child.Bounds.H, c.VerticalAlign)
			pass2Position(child, cx, cy)
			cx += child.Bounds.W + gap
		}
		return
	}

	totalHeight := totalMainSize(c.Children, gap, func(child *Container) float64 {
		return child.Bounds.H
	})
	cy := alignedStart(contentY, contentH, totalHeight, c.VerticalAlign)
	for _, child := range c.Children {
		cx := alignItem(contentX, contentW, child.Bounds.W, c.HorizontalAlign)
		pass2Position(child, cx, cy)
		cy += child.Bounds.H + gap
	}
}

func totalMainSize(children []*Container, gap float64, sizeOf func(*Container) float64) float64 {
	if len(children) == 0 {
		return 0
	}
	total := gap * float64(len(children)-1)
	for _, child := range children {
		total += sizeOf(child)
	}
	return total
}

func alignedStart(start, available, content float64, alignment Alignment) float64 {
	switch alignment {
	case AlignCenter:
		return start + (available-content)/2
	case AlignEnd:
		return start + available - content
	default:
		return start
	}
}

func alignItem(start, available, size float64, alignment Alignment) float64 {
	return alignedStart(start, available, size, alignment)
}

func directionOf(c *Container) Direction {
	if c.Direction == Row {
		return Row
	}
	return Column
}

func gaps(c *Container) float64 {
	if len(c.Children) < 2 {
		return 0
	}
	return nonnegativeFinite(c.Gap) * float64(len(c.Children)-1)
}

func contentBox(c *Container, width, height float64) (x, y, contentW, contentH float64) {
	insets := c.Insets
	if insets == (Insets{}) {
		insets = UniformInsets(c.Padding)
	}
	x, contentW = insetAxis(width, insets.Left, insets.Right)
	y, contentH = insetAxis(height, insets.Top, insets.Bottom)
	return x, y, contentW, contentH
}

// insetAxis keeps content non-negative. Oversized asymmetric insets are
// reduced proportionally, preserving their relative placement.
func insetAxis(size, start, end float64) (offset, content float64) {
	if size <= 0 {
		return 0, 0
	}
	start = nonnegativeFinite(start)
	end = nonnegativeFinite(end)
	if total := start + end; total > size {
		scale := size / total
		start *= scale
		end *= scale
	}
	return start, max(0, size-start-end)
}

func insetSize(size, padding float64) (offset, content float64) {
	return insetAxis(size, padding, padding)
}
