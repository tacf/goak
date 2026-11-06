package main

import "github.com/tacf/goak"

var colorDemo = struct {
	rgb  [3]float32
	rgba [4]float32
	pick [4]float32
}{
	rgb:  [3]float32{0.9, 0.5, 0.2},
	rgba: [4]float32{0.3, 0.6, 0.9, 0.8},
	pick: [4]float32{0.8, 0.2, 0.4, 1},
}

// colorSection shows ColorEdit and ColorPicker.
func colorSection(c *goak.Context) {
	if !c.CollapsingHeader("Colors") {
		return
	}
	d := &colorDemo
	c.ColorEdit3("Edit RGB", &d.rgb)
	c.ColorEdit4("Edit RGBA", &d.rgba)
	c.ColorPicker4("##picker", &d.pick)
}
