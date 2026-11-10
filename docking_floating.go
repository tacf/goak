package goak

import "fmt"

func (c *Context) floatingDockRoot(w *Window) *dockNode {
	if w.dockGroup != nil {
		return w.dockGroup
	}
	if n := c.dockWindows[w.ID]; n != nil && n.root.floating {
		return n.root
	}
	return nil
}

func (c *Context) floatingDockGrip(n *dockNode) Rect {
	return Rect{Min: n.rect.Max.Sub(Vec2{X: c.Style.ResizeGripSize, Y: c.Style.ResizeGripSize}), Max: n.rect.Max}
}

func (c *Context) newFloatingDock(target *Window) *dockNode {
	id := c.DockSpace(fmt.Sprintf("##FLOATDOCK%d", target.ID), target.Pos, target.Size)
	n := c.dockNodes[id]
	n.floating = true
	n.host.dockGroup = n
	c.floatingDocks = append(c.floatingDocks, n)
	c.DockWindow(target.Name, id)
	return n
}

func (c *Context) submitFloatingDocks() {
	keep := c.floatingDocks[:0]
	for _, n := range c.floatingDocks {
		if n.dirty {
			c.compactDockNode(n)
			n.dirty = false
		}
		if n.children[0] == nil {
			if w := c.windows[n.window]; w != nil {
				c.UndockWindow(w.Name)
				w.Pos, w.Size = n.rect.Min, n.rect.Max.Sub(n.rect.Min)
			}
			n.host.skip = true
			n.host.dockGroup = nil
			c.forgetDockSubtree(n)
			continue
		}
		c.DockSpace(n.name, n.rect.Min, n.rect.Max.Sub(n.rect.Min))
		keep = append(keep, n)
	}
	clear(c.floatingDocks[len(keep):])
	c.floatingDocks = keep
}
