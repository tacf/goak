// Package mathx has small numeric helpers.
package mathx

import "math"

// Finite32 reports whether f is neither NaN nor infinite.
func Finite32(f float32) bool { return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0) }
