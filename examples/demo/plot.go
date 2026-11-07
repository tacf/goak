package main

import (
	"math"

	"github.com/tacf/goak"
)

var plotAPI struct {
	xs, sin, cos, noise []float64
	barsX, bars         []float64
}

func initPlot() {
	d := &plotAPI
	for i := 0; i < 200; i++ {
		x := float64(i) * 0.05
		d.xs = append(d.xs, x)
		d.sin = append(d.sin, math.Sin(x))
		d.cos = append(d.cos, math.Cos(x)*0.6)
	}
	for i := 0; i < 24; i++ {
		x := float64(i) * 0.42
		d.barsX = append(d.barsX, x)
		d.bars = append(d.bars, 0.25*math.Sin(x*1.7)+0.1)
		d.noise = append(d.noise, math.Sin(x)+0.2*math.Sin(float64(i)*12.9898))
	}
}

// plotSection shows BeginPlot with lines, markers and bars.
func plotSection(c *goak.Context) {
	if !c.CollapsingHeader("Plot") {
		return
	}
	d := &plotAPI
	if d.xs == nil {
		initPlot()
	}
	c.Text("Drag to pan, wheel to zoom, double click to fit")
	if c.BeginPlot("Waves", goak.Vec2{Y: 260}, 0) {
		c.PlotSetupAxes("x", "value")
		c.PlotLine("sin", d.xs, d.sin)
		c.PlotLine("cos", d.xs, d.cos)
		c.PlotScatter("samples", d.barsX, d.noise)
		c.PlotBars("bars", d.barsX, d.bars, 0.3)
		c.EndPlot()
	}
}
