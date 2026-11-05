package main

import (
	"math"

	"github.com/tacf/goak"
)

var plotLinesDemo struct {
	ring   [90]float32
	offset int
	phase  float64
	bars   []float32
}

// plotLinesSection shows PlotLines with a ring buffer and PlotHistogram.
func plotLinesSection(c *goak.Context) {
	if !c.CollapsingHeader("Plot lines") {
		return
	}
	d := &plotLinesDemo
	if d.bars == nil {
		for i := 0; i < 12; i++ {
			d.bars = append(d.bars, float32(3+math.Sin(float64(i)*0.9)*2))
		}
	}
	d.phase += 0.08
	d.ring[d.offset] = float32(math.Sin(d.phase) * math.Cos(d.phase*0.37))
	d.offset = (d.offset + 1) % len(d.ring)
	c.PlotLines("Signal", d.ring[:], d.offset, "sin * cos", -1, 1, goak.Vec2{Y: 70})
	c.PlotHistogram("Bars", d.bars, 0, "", 0, goak.AutoScale, goak.Vec2{Y: 70})
}
