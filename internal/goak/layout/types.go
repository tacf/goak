package layout

// Sizing is how a dimension is specified.
type Sizing int

const (
	Static  Sizing = iota // fixed pixels
	Percent                 // percentage of parent (0–100)
	Auto                    // fill remaining space
)

// Size specifies width or height: Static (px), Percent (0–100), or Auto.
type Size struct {
	Kind  Sizing
	Value float64 // pixels for Static, 0–100 for Percent; ignored for Auto
}

// Alignment controls child placement inside a container.
type Alignment int

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

// StaticPx returns a static size in pixels.
func StaticPx(pixels float64) Size {
	return Size{Kind: Static, Value: pixels}
}

// PercentOf returns a percent-based size (0–100).
func PercentOf(pct float64) Size {
	return Size{Kind: Percent, Value: pct}
}

// AutoSize returns an auto (fill-remaining) size.
func AutoSize() Size {
	return Size{Kind: Auto}
}

// Rect is the computed bounds (x, y, width, height) after layout.
type Rect struct {
	X, Y, W, H float64
}

// Container is a nested layout node. Width and Height define size; Bounds is filled by Layout.
type Container struct {
	Width           Size
	Height          Size
	HorizontalAlign Alignment
	VerticalAlign   Alignment
	Children        []*Container
	Bounds          Rect // set by Layout (Pass 1 + Pass 2)
}

// NewContainer returns a container with optional children. Default size is Auto.
func NewContainer(width, height Size, children ...*Container) *Container {
	c := &Container{
		Width:           width,
		Height:          height,
		HorizontalAlign: AlignStart,
		VerticalAlign:   AlignStart,
	}
	if len(children) > 0 {
		c.Children = children
	}
	return c
}
