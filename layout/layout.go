// Package layout exposes Goak's two-pass layout primitives.
package layout

import internal "goak/internal/goak/layout"

type (
	Alignment = internal.Alignment
	Container = internal.Container
	Rect      = internal.Rect
	Size      = internal.Size
	Sizing    = internal.Sizing
)

const (
	Static  = internal.Static
	Percent = internal.Percent
	Auto    = internal.Auto

	AlignStart  = internal.AlignStart
	AlignCenter = internal.AlignCenter
	AlignEnd    = internal.AlignEnd
)

var (
	AutoSize     = internal.AutoSize
	Layout       = internal.Layout
	NewContainer = internal.NewContainer
	PercentOf    = internal.PercentOf
	StaticPx     = internal.StaticPx
)
