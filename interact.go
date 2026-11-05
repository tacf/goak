package goak

// ButtonFlags tune ButtonBehaviorEx.
type ButtonFlags uint8

// Button flags.
const (
	// BtnNoFocus keeps the item out of Tab order and click focus.
	BtnNoFocus ButtonFlags = 1 << iota
)

// ButtonState is the result of one interaction test.
type ButtonState struct {
	Pressed bool // clicked: released inside after pressing on the item
	Hovered bool
	Held    bool
	Started bool // the press began this frame
}

// mouseOver reports whether the mouse is over r, clipped by the current
// window clip, with the window being the topmost one under the mouse.
func (c *Context) mouseOver(r Rect) bool {
	w := c.cur
	if w == nil || w.skip || c.hoveredWindow != w.Root {
		return false
	}
	return r.Intersect(w.clip).Contains(c.mouse())
}

// ButtonBehavior is ButtonBehaviorEx with default flags.
func (c *Context) ButtonBehavior(bounds Rect, id ID) (pressed, hovered, held bool) {
	s := c.ButtonBehaviorEx(bounds, id, 0)
	return s.Pressed, s.Hovered, s.Held
}

// ButtonBehaviorEx handles hover and press. Hover needs the topmost window under
// the mouse and no other active item. A press on a hovered item makes it
// active and captures the mouse until release. Pressed is true only when the
// release happens while still hovering the item that got the press, so a
// press outside followed by a release inside is not a click. Enter or Space
// also click the focused item.
func (c *Context) ButtonBehaviorEx(bounds Rect, id ID, fl ButtonFlags) ButtonState {
	var s ButtonState
	if c.disabled > 0 {
		return s
	}
	in := c.In
	left := MouseLeft
	w := c.cur
	visible := !w.skip && !bounds.Intersect(w.clip).Empty()
	if !visible && c.activeID != id {
		return s
	}
	focusable := fl&BtnNoFocus == 0 && visible
	if focusable {
		c.focusList = append(c.focusList, id)
	}
	if c.mouseOver(bounds) && (c.activeID == 0 || c.activeID == id) && !c.blankPress {
		s.Hovered = true
		c.hotID = id
	}
	if s.Hovered && in.MousePress[left] && c.activeID == 0 {
		c.activeID = id
		c.activeWindow = c.cur.Root
		c.pressClaimed = true
		s.Started = true
		if focusable {
			c.focusID = id
		}
	}
	if c.activeID == id {
		c.activeSeen = true
		released := in.MouseRel[left] || !in.MouseDown[left]
		s.Held = !released
		if released {
			s.Pressed = s.Hovered
			c.activeID, c.activeWindow = 0, nil
		}
	}
	if focusable && c.focusID == id && (in.KeyPressed[KeyEnter] || in.KeyPressed[KeySpace]) {
		s.Pressed = true
	}
	return s
}
