package main

import (
	"github.com/tacf/goak"
	"github.com/tacf/goak/colors"
)

var scopeDemo struct {
	disabled bool
	flag     bool
}

// scopeSection shows groups, disabled state and style color overrides.
func scopeSection(c *goak.Context) {
	if !c.CollapsingHeader("Groups and disabled") {
		return
	}
	c.Checkbox("Disable the widgets below", &scopeDemo.disabled)
	if scopeDemo.disabled {
		c.BeginDisabled()
	}
	c.Button("Button")
	c.Checkbox("Flag", &scopeDemo.flag)
	if scopeDemo.disabled {
		c.EndDisabled()
	}
	c.BeginGroup()
	c.Button("One")
	c.Button("Two")
	c.EndGroup()
	c.SameLine(0, -1)
	c.Text("A group is one item")
	c.PushStyleColor(goak.ColButton, colors.RGB(160, 50, 50))
	c.PushStyleColor(goak.ColButtonHovered, colors.RGB(190, 70, 70))
	c.Button("Danger")
	c.PopStyleColor(2)
}
