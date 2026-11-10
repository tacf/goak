package goak

import (
	"fmt"
	"math"

	"github.com/tacf/goak/internal/draw"
)

// DockDirection selects the side occupied by a new split.
type DockDirection uint8

// Dock split directions.
const (
	DockLeft DockDirection = iota
	DockRight
	DockTop
	DockBottom
)

const dockGap float32 = 4

// DockSpaceFlags configure the central workspace.
type DockSpaceFlags uint8

const (
	// DockNoDockingInCentralNode keeps the central leaf empty.
	DockNoDockingInCentralNode DockSpaceFlags = 1 << iota
	// DockPassthruCentralNode leaves an empty center transparent to drawing and input.
	DockPassthruCentralNode
)

type dockNode struct {
	id        ID
	root      *dockNode
	children  [2]*dockNode
	direction DockDirection
	ratio     float32
	rect      Rect
	window    ID
	host      *Window
	frame     int
	central   *dockNode
	flags     DockSpaceFlags
	dirty     bool
	floating  bool
	name      string
	grip      *draw.DrawList
}

// DockSpace submits a persistent docking area outside any window, before panels.
// Its position and size are in screen coordinates. Nonpositive sizes use the
// remaining display space. The returned root ID is stable across frames.
func (c *Context) DockSpace(name string, pos, size Vec2, flags ...DockSpaceFlags) ID {
	if c.cur != nil {
		panic("goak: DockSpace inside a window")
	}
	if c.dockNodes == nil {
		c.dockNodes = make(map[ID]*dockNode)
		c.dockWindows = make(map[ID]*dockNode)
	}
	id := hashLabel(name, hashStr("#DOCKSPACE", 0))
	n := c.dockNodes[id]
	if n == nil {
		n = &dockNode{id: id, name: name}
		n.root = n
		n.central = n
		c.dockNodes[id] = n
	}
	if n.root != n {
		panic("goak: dock space ID conflicts with a split")
	}
	if size.X <= 0 {
		size.X = max(c.DisplaySize.X-pos.X+size.X, 0)
	}
	if size.Y <= 0 {
		size.Y = max(c.DisplaySize.Y-pos.Y+size.Y, 0)
	}
	n.frame = c.Frame
	n.flags = 0
	for _, flag := range flags {
		n.flags |= flag
	}
	if n.flags&DockNoDockingInCentralNode != 0 && n.central.window != 0 {
		if w := c.windows[n.central.window]; w != nil {
			c.UndockWindow(w.Name)
		} else {
			delete(c.dockWindows, n.central.window)
			n.central.window = 0
			n.dirty = true
		}
	}
	if n.dirty {
		c.compactDockNode(n)
		n.dirty = false
	}
	n.rect = Rect{Min: pos, Max: pos.Add(size)}
	c.SetNextWindowPos(pos, CondAlways)
	c.SetNextWindowSize(size, CondAlways)
	hostFlags := NoTitleBar | NoMove | NoDocking | NoBackground
	if !n.floating {
		hostFlags |= NoResize
	}
	c.Begin(fmt.Sprintf("##DOCKHOST%d", id), nil, hostFlags)
	n.host = c.cur
	n.host.dockHost = true
	if n.floating {
		n.host.dockGroup = n
		pos, size = n.host.Pos, n.host.Size
		size = size.Max(n.minSize(c.Style.WindowMinSize))
		n.rect = Rect{Min: pos, Max: pos.Add(size)}
	}
	n.host.Pos, n.host.Size = pos, size
	n.host.DL.Reset()
	n.host.DL.PushClipRect(n.rect)
	n.host.clip = n.rect
	n.layout()
	c.dockBackground(n)
	c.dockSplitters(n)
	c.End()
	if n.floating {
		if n.grip == nil {
			n.grip = draw.NewDrawList()
		}
		n.grip.Reset()
		n.grip.Scale = c.pixelScale()
		n.grip.PushClipRect(n.rect)
		c.drawGripTo(n.host, n.grip)
		n.grip.PopClipRect()
	}
	return id
}

// DockSpaceOverViewport fills the display below a submitted main menu bar.
func (c *Context) DockSpaceOverViewport(name string, flags ...DockSpaceFlags) ID {
	var pos Vec2
	if bar := c.WindowByName("##MainMenuBar"); bar != nil && bar.lastFrame == c.Frame && !bar.skip {
		pos.Y = min(bar.Pos.Y+bar.Size.Y, c.DisplaySize.Y)
	}
	return c.DockSpace(name, pos, Vec2{}, flags...)
}

// DockSpaceCentralNode returns the current central leaf, or zero for an invalid root.
func (c *Context) DockSpaceCentralNode(root ID) ID {
	if n := c.dockNodes[root]; n != nil && n.root == n {
		return n.central.id
	}
	return 0
}

// compactDockNode preserves occupied leaves and the central workspace.
func (c *Context) compactDockNode(n *dockNode) bool {
	if n.children[0] == nil {
		return n.window != 0 || (n == n.root.central && !n.root.floating)
	}
	a, b := n.children[0], n.children[1]
	left, right := c.compactDockNode(a), c.compactDockNode(b)
	if left && right {
		return true
	}
	keep, drop := a, b
	if right {
		keep, drop = b, a
	}
	c.forgetDockSubtree(drop)
	delete(c.dockNodes, keep.id)
	n.children, n.window = keep.children, keep.window
	n.direction, n.ratio = keep.direction, keep.ratio
	if n.window != 0 {
		c.dockWindows[n.window] = n
	}
	if n.root.central == keep {
		n.root.central = n
	}
	if n.root.central == nil {
		leaf := n
		for leaf.children[0] != nil {
			leaf = leaf.children[0]
		}
		n.root.central = leaf
	}
	return left || right
}

func (c *Context) forgetDockSubtree(n *dockNode) {
	if n.children[0] != nil {
		c.forgetDockSubtree(n.children[0])
		c.forgetDockSubtree(n.children[1])
	}
	if n.root.central == n {
		n.root.central = nil
	}
	delete(c.dockNodes, n.id)
}

func (n *dockNode) emptyPassthruCenter(p Vec2) bool {
	return n.flags&DockPassthruCentralNode != 0 && n.central.window == 0 && n.central.rect.Contains(p)
}

func (c *Context) dockBackground(n *dockNode) {
	if n.children[0] != nil {
		c.dockBackground(n.children[0])
		c.dockBackground(n.children[1])
		return
	}
	if n == n.root.central && n.window == 0 && n.root.flags&DockPassthruCentralNode != 0 {
		return
	}
	c.cur.DL.AddRectFilled(n.rect.Min, n.rect.Max, c.col(ColWindowBg), 0)
}

func (c *Context) dockHostPassesInput(w *Window, p Vec2) bool {
	if !w.dockHost {
		return false
	}
	for _, n := range c.dockNodes {
		if n.root == n && n.host == w && n.frame == c.Frame-1 {
			return n.emptyPassthruCenter(p)
		}
	}
	return false
}

// SplitDockNode splits a leaf, returning the requested side and the remainder.
// ratio is the fraction assigned to the requested side, clamped to 0.1..0.9.
// An existing panel moves to the remainder. Invalid requests return zero IDs.
// Build an initial layout once, then keep submitting DockSpace each frame.
func (c *Context) SplitDockNode(id ID, direction DockDirection, ratio float32) (ID, ID) {
	n := c.dockNodes[id]
	if n == nil || n.children[0] != nil || direction > DockBottom || math.IsNaN(float64(ratio)) {
		return 0, 0
	}
	for i := range n.children {
		var childID ID
		for childID == 0 || c.dockNodes[childID] != nil {
			c.dockSerial++
			childID = hashStr(fmt.Sprintf("#DOCKNODE%d", c.dockSerial), n.root.id)
		}
		child := &dockNode{id: childID, root: n.root}
		c.dockNodes[childID] = child
		n.children[i] = child
	}
	n.direction, n.ratio = direction, max(0.1, min(ratio, 0.9))
	if n.root.central == n {
		n.root.central = n.children[1]
	}
	if n.window != 0 {
		n.children[1].window = n.window
		c.dockWindows[n.window] = n.children[1]
		n.window = 0
	}
	n.layout()
	return n.children[0].id, n.children[1].id
}

// DockWindow assigns a title to an empty leaf, even before its first Begin.
// It returns false for invalid nodes, occupied leaves or NoDocking windows.
func (c *Context) DockWindow(title string, nodeID ID) bool {
	n := c.dockNodes[nodeID]
	id := hashLabel(title, 0)
	w := c.windows[id]
	if n == nil || n.children[0] != nil || (n.window != 0 && n.window != id) ||
		(n == n.root.central && n.root.flags&DockNoDockingInCentralNode != 0) ||
		(w != nil && (w.top || w.Flags&NoDocking != 0)) {
		return false
	}
	if old := c.dockWindows[id]; old != nil && old != n {
		old.window = 0
		old.root.dirty = true
	}
	n.window = id
	c.dockWindows[id] = n
	return true
}

// UndockWindow restores the panel's last floating position and size.
func (c *Context) UndockWindow(title string) {
	id := hashLabel(title, 0)
	if n := c.dockWindows[id]; n != nil {
		n.window = 0
		n.root.dirty = true
		delete(c.dockWindows, id)
	}
	if w := c.windows[id]; w != nil && w.docked {
		w.Pos, w.Size = w.floatPos, w.floatSize
		w.docked = false
	}
}

// WindowDockNode returns the assigned leaf ID, or zero for a floating window.
func (c *Context) WindowDockNode(title string) ID {
	if n := c.dockWindows[hashLabel(title, 0)]; n != nil {
		return n.id
	}
	return 0
}

func (n *dockNode) layout() {
	if n.children[0] == nil {
		return
	}
	a, b := n.children[0], n.children[1]
	a.rect, b.rect = n.rect, n.rect
	if n.direction == DockLeft || n.direction == DockRight {
		span := max(n.rect.Max.X-n.rect.Min.X-dockGap, 0)
		if n.direction == DockLeft {
			a.rect.Max.X = n.rect.Min.X + span*n.ratio
			b.rect.Min.X = min(a.rect.Max.X+dockGap, n.rect.Max.X)
		} else {
			a.rect.Min.X = n.rect.Max.X - span*n.ratio
			b.rect.Max.X = max(a.rect.Min.X-dockGap, n.rect.Min.X)
		}
	} else {
		span := max(n.rect.Max.Y-n.rect.Min.Y-dockGap, 0)
		if n.direction == DockTop {
			a.rect.Max.Y = n.rect.Min.Y + span*n.ratio
			b.rect.Min.Y = min(a.rect.Max.Y+dockGap, n.rect.Max.Y)
		} else {
			a.rect.Min.Y = n.rect.Max.Y - span*n.ratio
			b.rect.Max.Y = max(a.rect.Min.Y-dockGap, n.rect.Min.Y)
		}
	}
	a.layout()
	b.layout()
}

func (c *Context) applyDock(w *Window) bool {
	n := c.dockWindows[w.ID]
	if n == nil {
		return false
	}
	if w.Flags&NoDocking != 0 || w.top {
		c.UndockWindow(w.Name)
		return false
	}
	if !w.docked {
		w.floatPos, w.floatSize = w.Pos, w.Size
		w.docked = true
	}
	w.Pos, w.Size = n.rect.Min, n.rect.Max.Sub(n.rect.Min)
	w.skip = n.root.frame != c.Frame || n.rect.Empty()
	return true
}

func (n *dockNode) splitterRect() Rect {
	a, b := n.children[0].rect, n.children[1].rect
	r := n.rect
	switch n.direction {
	case DockLeft:
		r.Min.X, r.Max.X = a.Max.X, b.Min.X
	case DockRight:
		r.Min.X, r.Max.X = b.Max.X, a.Min.X
	case DockTop:
		r.Min.Y, r.Max.Y = a.Max.Y, b.Min.Y
	case DockBottom:
		r.Min.Y, r.Max.Y = b.Max.Y, a.Min.Y
	}
	return r
}

func (n *dockNode) minSize(leaf Vec2) Vec2 {
	if n.children[0] == nil {
		return leaf
	}
	a, b := n.children[0].minSize(leaf), n.children[1].minSize(leaf)
	if n.direction == DockLeft || n.direction == DockRight {
		return Vec2{X: a.X + b.X + dockGap, Y: max(a.Y, b.Y)}
	}
	return Vec2{X: max(a.X, b.X), Y: a.Y + b.Y + dockGap}
}

func (c *Context) dockSplitters(n *dockNode) {
	if n.children[0] == nil {
		return
	}
	r := n.splitterRect()
	s := c.ButtonBehaviorEx(r, hashStr("#DOCKSPLIT", n.id), BtnNoFocus)
	if s.Started {
		c.dragOffset = c.mouse().Sub(r.Min)
	}
	if s.Held {
		p := c.mouse().Sub(c.dragOffset)
		a := n.children[0].minSize(c.Style.WindowMinSize)
		b := n.children[1].minSize(c.Style.WindowMinSize)
		span, offset, amin, bmin := n.rect.Max.X-n.rect.Min.X-dockGap, p.X-n.rect.Min.X, a.X, b.X
		switch n.direction {
		case DockRight:
			offset = n.rect.Max.X - p.X - dockGap
		case DockTop:
			span, offset, amin, bmin = n.rect.Max.Y-n.rect.Min.Y-dockGap, p.Y-n.rect.Min.Y, a.Y, b.Y
		case DockBottom:
			span, offset, amin, bmin = n.rect.Max.Y-n.rect.Min.Y-dockGap, n.rect.Max.Y-p.Y-dockGap, a.Y, b.Y
		}
		if span > 0 {
			lo, hi := amin/span, 1-bmin/span
			if lo > hi {
				lo, hi = amin/max(amin+bmin, 1), amin/max(amin+bmin, 1)
			}
			n.ratio = max(lo, min(offset/span, hi))
			n.layout()
			r = n.splitterRect()
		}
	}
	col := ColSeparator
	if s.Held {
		col = ColResizeGripActive
	} else if s.Hovered {
		col = ColResizeGripHovered
	}
	c.cur.DL.AddRectFilled(r.Min, r.Max, c.col(col), 0)
	c.dockSplitters(n.children[0])
	c.dockSplitters(n.children[1])
}

func (n *dockNode) leafAt(p Vec2) *dockNode {
	if !n.rect.Contains(p) {
		return nil
	}
	if n.children[0] == nil {
		return n
	}
	if leaf := n.children[0].leafAt(p); leaf != nil {
		return leaf
	}
	return n.children[1].leafAt(p)
}

func (c *Context) dockTarget(w *Window) (*dockNode, *Window) {
	p := c.mouse()
	layers := c.layers()
	for i := len(layers) - 1; i >= 0; i-- {
		target := layers[i]
		if target == w || target.skip || target.noInputs || target.lastFrame != c.Frame || !target.rect().Contains(p) {
			continue
		}
		if n := c.dockWindows[target.ID]; n != nil && !n.rect.Empty() {
			return n, nil
		}
		if target.dockHost {
			for _, n := range c.dockNodes {
				if n.root == n && n.host == target && n.frame == c.Frame {
					return n.leafAt(p), nil
				}
			}
		}
		if !target.top && !target.collapsed && target.Flags&NoDocking == 0 {
			return nil, target
		}
		return nil, nil
	}
	return nil, nil
}

const dockCenter DockDirection = 4

func (c *Context) dockTargets(n *dockNode, paint bool) (DockDirection, bool) {
	if c.dockPreview == nil {
		c.dockPreview = draw.NewDrawList()
	}
	selected, hit := dockCenter, false
	center := Vec2{X: (n.rect.Min.X + n.rect.Max.X) / 2, Y: (n.rect.Min.Y + n.rect.Max.Y) / 2}
	const button, spacing float32 = 26, 4
	for side := DockLeft; side <= dockCenter; side++ {
		if side == dockCenter {
			if n.window != 0 || (n == n.root.central && n.root.flags&DockNoDockingInCentralNode != 0) {
				continue
			}
		} else {
			minimum := c.Style.WindowMinSize
			if (side <= DockRight && n.rect.Max.X-n.rect.Min.X < 2*minimum.X+dockGap) ||
				(side >= DockTop && n.rect.Max.Y-n.rect.Min.Y < 2*minimum.Y+dockGap) {
				continue
			}
		}
		p := center
		switch side {
		case DockLeft:
			p.X -= button + spacing
		case DockRight:
			p.X += button + spacing
		case DockTop:
			p.Y -= button + spacing
		case DockBottom:
			p.Y += button + spacing
		}
		r := Rect{Min: p.Sub(Vec2{X: button / 2, Y: button / 2}), Max: p.Add(Vec2{X: button / 2, Y: button / 2})}.Intersect(n.rect)
		if r.Empty() {
			continue
		}
		col := c.col(ColTab)
		if r.Contains(c.mouse()) {
			selected, hit = side, true
			col = c.col(ColTabHovered)
		}
		if !paint {
			continue
		}
		c.dockPreview.AddRectFilled(r.Min, r.Max, col, 2)
		c.dockPreview.AddRect(r.Min, r.Max, c.col(ColResizeGripActive), 2, 1)
		mark := Rect{Min: p.Sub(Vec2{X: 7, Y: 7}), Max: p.Add(Vec2{X: 7, Y: 7})}
		if side != dockCenter {
			mark = dockHalf(mark, side)
		}
		c.dockPreview.AddRectFilled(mark.Min, mark.Max, c.col(ColText), 0)
	}
	return selected, hit
}

func dockHalf(r Rect, side DockDirection) Rect {
	switch side {
	case DockLeft:
		r.Max.X = (r.Min.X + r.Max.X - dockGap) / 2
	case DockRight:
		r.Min.X = (r.Min.X + r.Max.X + dockGap) / 2
	case DockTop:
		r.Max.Y = (r.Min.Y + r.Max.Y - dockGap) / 2
	case DockBottom:
		r.Min.Y = (r.Min.Y + r.Max.Y + dockGap) / 2
	}
	return r
}

func (c *Context) finishDockDrag() {
	if c.dockPreview != nil {
		c.dockPreview.Reset()
	}
	w := c.dockDragging
	if w == nil {
		return
	}
	if c.In.KeyPressed[KeyEscape] {
		c.activeID, c.activeWindow = 0, nil
		c.dockDragging = nil
		return
	}
	if w.lastFrame != c.Frame || w.skip || w.Flags&(NoMove|NoDocking) != 0 {
		c.dockDragging = nil
		return
	}
	released := !c.In.MouseDown[MouseLeft]
	n, floating := c.dockTarget(w)
	if floating != nil {
		n = &dockNode{rect: floating.rect(), window: floating.ID}
		n.root = n
	}
	if n != nil && !c.In.Has(ModShift) {
		if c.dockPreview == nil {
			c.dockPreview = draw.NewDrawList()
		}
		c.dockPreview.Scale = c.pixelScale()
		c.dockPreview.PushClipRect(n.root.rect)
		side, hit := c.dockTargets(n, false)
		if hit && released {
			if floating != nil {
				n = c.newFloatingDock(floating)
			}
			id := n.id
			if side != dockCenter {
				id, _ = c.SplitDockNode(id, side, 0.5)
			}
			c.DockWindow(w.Name, id)
		} else if hit {
			r := n.rect
			if side != dockCenter {
				r = dockHalf(r, side)
			}
			col := c.col(ColResizeGripActive)
			col.A = 100
			c.dockPreview.AddRectFilled(r.Min, r.Max, col, 0)
			c.dockPreview.AddRect(r.Min, r.Max, c.col(ColResizeGripActive), 0, 2)
		}
		if !released {
			c.dockTargets(n, true)
		}
		c.dockPreview.PopClipRect()
		if released {
			c.dockPreview.Reset()
		}
	}
	if released {
		c.dockDragging = nil
	}
}
