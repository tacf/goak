package layout

import "math"

// Sizing is how a dimension is specified.
type Sizing int

const (
	Static  Sizing = iota // fixed pixels
	Percent               // percentage of parent (0–100)
	Auto                  // fill remaining space
)

// Size specifies width or height: Static (px), Percent (0–100), or Auto.
type Size struct {
	Kind  Sizing
	Value float64 // pixels for Static, 0–100 for Percent; ignored for Auto
	Min   float64 // minimum resolved size; values <= 0 disable the minimum
	Max   float64 // maximum resolved size; values <= 0 disable the maximum
}

// WithMin returns a copy of the size with a minimum pixel constraint.
func (s Size) WithMin(pixels float64) Size {
	s.Min = nonnegativeFinite(pixels)
	return s
}

// WithMax returns a copy of the size with a maximum pixel constraint.
func (s Size) WithMax(pixels float64) Size {
	s.Max = nonnegativeFinite(pixels)
	return s
}

// Alignment controls child placement inside a container.
type Alignment int

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

// Direction controls the axis along which a container stacks its children.
type Direction int

const (
	// Column stacks children from top to bottom.
	Column Direction = iota
	// Row stacks children from left to right.
	Row
)

// Insets specifies independent padding on each edge of a container.
type Insets struct {
	Top, Right, Bottom, Left float64
}

// UniformInsets returns equal insets for all four edges.
func UniformInsets(value float64) Insets {
	value = nonnegativeFinite(value)
	return Insets{Top: value, Right: value, Bottom: value, Left: value}
}

// StaticPx returns a static size in pixels.
func StaticPx(pixels float64) Size {
	return Size{Kind: Static, Value: nonnegativeFinite(pixels)}
}

// PercentOf returns a percent-based size (0–100).
func PercentOf(pct float64) Size {
	return Size{Kind: Percent, Value: min(100, nonnegativeFinite(pct))}
}

// AutoSize returns an auto (fill-remaining) size.
func AutoSize() Size {
	return Size{Kind: Auto}
}

func nonnegativeFinite(value float64) float64 {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

// Rect is the computed bounds (x, y, width, height) after layout.
type Rect struct {
	X, Y, W, H float64
}

// Container is a nested layout node. Width and Height define size; Padding
// applies a legacy uniform inset unless Insets contains a non-zero edge.
// Bounds is filled by Layout.
type Container struct {
	Width           Size
	Height          Size
	Padding         float64
	Insets          Insets
	Direction       Direction
	Gap             float64
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
