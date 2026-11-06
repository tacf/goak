package main

import (
	"fmt"

	"github.com/tacf/goak"
)

var tableDemo = struct {
	rows  []string
	flags goak.TableFlags
}{
	rows:  []string{"Ada", "Grace", "Linus", "Margaret", "Ken"},
	flags: goak.TableBorders | goak.TableRowBg | goak.TableResizable,
}

// tableSection shows a resizable table with a header row.
func tableSection(c *goak.Context) {
	if !c.CollapsingHeader("Table") {
		return
	}
	d := &tableDemo
	if c.BeginTable("people", 3, d.flags, goak.Vec2{}) {
		c.TableSetupColumn("ID", goak.TableColumnFixed, 50)
		c.TableSetupColumn("Name", 0, 2)
		c.TableSetupColumn("Score", 0, 1)
		c.TableHeadersRow()
		for i, name := range d.rows {
			c.TableNextColumn()
			c.Textf("%d", i+1)
			c.TableNextColumn()
			c.Text(name)
			c.TableNextColumn()
			c.Text(fmt.Sprintf("%.1f", 60+float32(i)*7.5))
		}
		c.EndTable()
	}
}
