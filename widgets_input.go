package goak

import (
	"strings"
	"unicode"

	"github.com/tacf/goak/internal/textedit"
)

// InputTextFlags tune InputText.
type InputTextFlags uint32

// Input text flags.
const (
	// InputPassword shows each rune as '*' and disables copy and cut.
	InputPassword InputTextFlags = 1 << iota
	// InputReadOnly allows selection and copy but no edits.
	InputReadOnly
	// InputEnterReturnsTrue makes the widget return true only when Enter is
	// pressed. In a multiline input Enter then no longer inserts a newline.
	InputEnterReturnsTrue
)

const caretBlink = 0.5

// InputText edits *s on one line. It returns true when the text changed,
// or when Enter was pressed with InputEnterReturnsTrue. The label is drawn
// to the right of the frame. Keys: arrows, Home, End, Backspace, Delete,
// Ctrl+arrows for words, Shift to select, Ctrl+A, Ctrl+C, Ctrl+X, Ctrl+V.
// A double click selects a word and the mouse drag selects a range.
func (c *Context) InputText(label string, s *string, flags InputTextFlags) bool {
	return c.inputText(label, "", s, Vec2{}, flags, false)
}

// InputTextWithHint is InputText showing hint while the text is empty.
func (c *Context) InputTextWithHint(label, hint string, s *string, flags InputTextFlags) bool {
	return c.inputText(label, hint, s, Vec2{}, flags, false)
}

// InputTextMultiline edits *s over several lines in a frame of the given
// size. A size component <= 0 means the available space plus that value,
// and a zero height means five lines. Long lines scroll horizontally and
// there is no word wrap.
func (c *Context) InputTextMultiline(label string, s *string, size Vec2, flags InputTextFlags) bool {
	return c.inputText(label, "", s, size, flags, true)
}

func (c *Context) textEditState(id ID) *textEdit {
	key := hashStr("#TEXT", id)
	e := StateGet[*textEdit](c.States, key, nil)
	if e == nil {
		e = &textEdit{}
		StateSetKeep(c.States, key, e, LongAge)
	}
	return e
}

// buildSpans returns the line ranges and fills the display line cache.
// It does nothing while the text revision is unchanged.
func (c *Context) buildSpans(e *textEdit, multi, pw bool) [][2]int {
	if e.cacheOK && e.cacheRev == e.Rev && e.cachePW == pw && e.cacheMult == multi {
		return e.spans
	}
	e.cacheOK, e.cacheRev, e.cachePW, e.cacheMult, e.maxWOK = true, e.Rev, pw, multi, false
	e.spans = e.spans[:0]
	if !multi {
		e.spans = append(e.spans, [2]int{0, len(e.Text)})
	} else {
		start := 0
		for i, r := range e.Text {
			if r == '\n' {
				e.spans = append(e.spans, [2]int{start, i})
				start = i + 1
			}
		}
		e.spans = append(e.spans, [2]int{start, len(e.Text)})
	}
	e.lines = e.lines[:0]
	for _, sp := range e.spans {
		e.lines = append(e.lines, c.disp(e, sp[0], sp[1], pw))
	}
	return e.spans
}

// maxWidth is the widest line, measured once per text revision.
func (c *Context) maxWidth(e *textEdit) float32 {
	if !e.maxWOK {
		e.maxW = 0
		for _, l := range e.lines {
			e.maxW = max(e.maxW, c.measure(l))
		}
		e.maxWOK = true
	}
	return e.maxW
}

func (e *textEdit) lineOf(spans [][2]int, idx int) int {
	for i, sp := range spans {
		if idx <= sp[1] {
			return i
		}
	}
	return len(spans) - 1
}

func (c *Context) disp(e *textEdit, a, b int, pw bool) string {
	if a >= b {
		return ""
	}
	if pw {
		return strings.Repeat("*", b-a)
	}
	return string(e.Text[a:b])
}

func (c *Context) measure(s string) float32 {
	if s == "" || c.Font == nil {
		return 0
	}
	w, _ := c.Font.Measure(s)
	return float32(w)
}

// indexAt returns the rune index within [a, b) closest to x.
func (c *Context) indexAt(e *textEdit, a, b int, x float32, pw bool) int {
	if x <= 0 || b <= a {
		return a
	}
	// Smallest k whose prefix width reaches x, by binary search.
	lo, hi := a+1, b+1
	for lo < hi {
		m := (lo + hi) / 2
		if c.measure(c.disp(e, a, m, pw)) >= x {
			hi = m
		} else {
			lo = m + 1
		}
	}
	if lo > b {
		return b
	}
	prev := c.measure(c.disp(e, a, lo-1, pw))
	w := c.measure(c.disp(e, a, lo, pw))
	if x < (prev+w)/2 {
		return lo - 1
	}
	return lo
}

func (c *Context) inputText(label, hint string, s *string, size Vec2, flags InputTextFlags, multi bool) bool {
	if s == nil {
		return false
	}
	id := c.ItemID(label)
	text := DisplayLabel(label)
	st := &c.Style
	fp := st.FramePadding
	lh := c.lineHeight()
	labelW := float32(0)
	if text != "" {
		labelW = st.ItemInnerSpacing.X + c.TextSize(text).X
	}
	w := c.CalcItemWidth()
	h := lh + 2*fp.Y
	if multi {
		w = size.X
		if w <= 0 {
			w = max(c.ContentRegionAvail().X+size.X-labelW, 1)
		}
		h = size.Y
		if h <= 0 {
			h = 5*lh + 2*fp.Y
		}
	}
	r := c.ItemAdd(Vec2{X: w + labelW, Y: h})
	if c.cur.skip {
		return false
	}
	frame := Rect{Min: r.Min, Max: Vec2{X: r.Min.X + w, Y: r.Max.Y}}
	inner := Rect{Min: frame.Min.Add(fp), Max: frame.Max.Sub(fp)}
	pw := flags&InputPassword != 0
	ro := flags&InputReadOnly != 0
	enterRet := flags&InputEnterReturnsTrue != 0

	e := c.textEditState(id)
	e.Load(*s, multi)
	beh := c.ButtonBehaviorEx(frame, id, 0)
	focused := c.focusID == id
	in := c.In

	// Wheel scroll of multiline text.
	if multi && beh.Hovered && !c.wheelUsed && in.Wheel[1] != 0 {
		e.scrollY -= in.Wheel[1] * st.ScrollStep * lh
		c.wheelUsed = true
	}

	spans := c.buildSpans(e, multi, pw)
	lineAt := func(my float32) int {
		if !multi {
			return 0
		}
		li := int((my - inner.Min.Y + e.scrollY) / lh)
		if my-inner.Min.Y+e.scrollY < 0 {
			li = 0
		}
		return min(max(li, 0), len(spans)-1)
	}
	mouseIndex := func() int {
		sp := spans[lineAt(c.mouse().Y)]
		return c.indexAt(e, sp[0], sp[1], c.mouse().X-(inner.Min.X-e.scrollX), pw)
	}
	if beh.Started {
		e.noDrag = false
		idx := mouseIndex()
		switch clicks := in.MouseClicks[MouseLeft]; {
		case clicks >= 3:
			sp := spans[lineAt(c.mouse().Y)]
			e.Anchor, e.Cursor = sp[0], sp[1]
			e.noDrag = true
		case clicks == 2:
			e.Anchor, e.Cursor = e.WordAt(idx)
			e.noDrag = true
		default:
			e.MoveCursor(idx, in.Has(ModShift))
		}
		e.Reveal = true
	} else if c.activeID == id && beh.Held && !e.noDrag {
		e.Cursor = mouseIndex()
		e.Reveal = true
	}
	if focused && !e.wasFocused && !beh.Started {
		e.Anchor, e.Cursor = 0, len(e.Text)
		e.Reveal = true
	}
	e.wasFocused = focused

	before, beforeCur, beforeAnc := len(e.Text), e.Cursor, e.Anchor
	edited, entered := false, false
	if focused {
		edited, entered = c.textKeys(e, in, flags, multi)
		if edited {
			*s = string(e.Text)
			e.Sync = *s
		}
	}
	if before != len(e.Text) || beforeCur != e.Cursor || beforeAnc != e.Anchor {
		e.blink = 0
		spans = c.buildSpans(e, multi, pw)
	}
	e.blink += in.DT

	c.drawInput(e, id, frame, inner, spans, hint, pw, ro, multi, focused, beh)
	c.frameLabel(frame, text)
	if enterRet {
		return entered
	}
	return edited
}

// textKeys applies this frame's keys and text. It reports whether the
// text changed and whether Enter was pressed on a single line or with
// InputEnterReturnsTrue.
func (c *Context) textKeys(e *textEdit, in *Input, flags InputTextFlags, multi bool) (edited, entered bool) {
	pw := flags&InputPassword != 0
	ro := flags&InputReadOnly != 0
	ctrl := in.Has(ModCtrl) || in.Has(ModSuper)
	shift := in.Has(ModShift)
	pressed := func(k Key) bool { return in.KeyPressed[k] }

	if pressed(KeyLeft) {
		switch {
		case ctrl:
			e.MoveCursor(e.MoveWord(-1), shift)
		case !shift && e.Cursor != e.Anchor:
			s, _, _ := e.Selection()
			e.MoveCursor(s, false)
		default:
			e.MoveCursor(e.Cursor-1, shift)
		}
	}
	if pressed(KeyRight) {
		switch {
		case ctrl:
			e.MoveCursor(e.MoveWord(1), shift)
		case !shift && e.Cursor != e.Anchor:
			_, en, _ := e.Selection()
			e.MoveCursor(en, false)
		default:
			e.MoveCursor(e.Cursor+1, shift)
		}
	}
	if multi && pressed(KeyUp) {
		e.MoveCursor(e.VerticalPosition(-1), shift)
	}
	if multi && pressed(KeyDown) {
		e.MoveCursor(e.VerticalPosition(1), shift)
	}
	if pressed(KeyHome) {
		ls := 0
		if multi && !ctrl {
			ls, _ = e.LineRange()
		}
		e.MoveCursor(ls, shift)
	}
	if pressed(KeyEnd) {
		le := len(e.Text)
		if multi && !ctrl {
			_, le = e.LineRange()
		}
		e.MoveCursor(le, shift)
	}
	if ctrl {
		if pressed(KeyA) {
			e.Anchor, e.Cursor = 0, len(e.Text)
		}
		if pressed(KeyC) && !pw {
			if t := e.SelectedText(); t != "" && c.Clipboard != nil {
				c.Clipboard.Set(t)
			}
		}
		if pressed(KeyX) && !pw && !ro && c.Clipboard != nil {
			if t := e.SelectedText(); t != "" {
				c.Clipboard.Set(t)
				edited = e.ReplaceSelection(nil) || edited
			}
		}
		if pressed(KeyV) && !ro && c.Clipboard != nil {
			if v := textedit.CleanPaste(c.Clipboard.Get(), multi); len(v) > 0 {
				edited = e.ReplaceSelection(v) || edited
			}
		}
	}
	if !ro {
		if pressed(KeyBackspace) {
			if e.Cursor == e.Anchor && e.Cursor > 0 {
				if ctrl {
					e.Anchor = e.MoveWord(-1)
				} else {
					e.Anchor = e.Cursor - 1
				}
			}
			edited = e.ReplaceSelection(nil) || edited
		}
		if pressed(KeyDelete) {
			if e.Cursor == e.Anchor && e.Cursor < len(e.Text) {
				if ctrl {
					e.Anchor = e.MoveWord(1)
				} else {
					e.Anchor = e.Cursor + 1
				}
			}
			edited = e.ReplaceSelection(nil) || edited
		}
	}
	if pressed(KeyEnter) {
		if !multi || flags&InputEnterReturnsTrue != 0 {
			entered = true
		} else if !ro {
			edited = e.ReplaceSelection([]rune{'\n'}) || edited
		}
	}
	if !ro && !ctrl && in.Text != "" {
		var rs []rune
		for _, r := range in.Text {
			if r == '\n' && multi || !unicode.IsControl(r) {
				rs = append(rs, r)
			}
		}
		if len(rs) > 0 {
			edited = e.ReplaceSelection(rs) || edited
		}
	}
	return edited, entered
}

func (c *Context) drawInput(e *textEdit, id ID, frame, inner Rect, spans [][2]int, hint string, pw, ro, multi, focused bool, beh ButtonState) {
	st := &c.Style
	lh := c.lineHeight()
	col := c.frameColor(beh, id, ColFrameBg, ColFrameBgHovered, ColFrameBgHovered)
	c.drawFrame(frame, col, focused)
	visW := inner.Max.X - inner.Min.X
	visH := inner.Max.Y - inner.Min.Y

	// Keep the caret visible and the scroll in range.
	cl := e.lineOf(spans, e.Cursor)
	csp := spans[cl]
	cx := c.measure(c.disp(e, csp[0], e.Cursor, pw))
	if e.Reveal {
		if cx-e.scrollX > visW-2 {
			e.scrollX = cx - visW + 2
		}
		if cx < e.scrollX {
			e.scrollX = cx
		}
		if multi {
			top := float32(cl) * lh
			if top < e.scrollY {
				e.scrollY = top
			}
			if top+lh > e.scrollY+visH {
				e.scrollY = top + lh - visH
			}
		}
		e.Reveal = false
	}
	maxW := float32(0)
	if e.scrollX > 0 || focused {
		maxW = c.maxWidth(e)
	}
	e.scrollX = min(max(e.scrollX, 0), max(maxW+2-visW, 0))
	total := float32(len(spans)) * lh
	e.scrollY = min(max(e.scrollY, 0), max(total-visH, 0))
	if !multi {
		e.scrollY = 0
	}

	dl := c.cur.DL
	dl.PushClipRect(inner)
	ox := inner.Min.X - e.scrollX
	oy := inner.Min.Y - e.scrollY
	if len(e.Text) == 0 && hint != "" {
		c.DrawText(inner.Min, hint, c.col(ColTextDisabled))
	}
	selA, selB, hasSel := e.Selection()
	for i, sp := range spans {
		y := oy + float32(i)*lh
		if y+lh < inner.Min.Y || y > inner.Max.Y {
			continue
		}
		if hasSel && focused && selB >= sp[0] && selA <= sp[1] {
			a, b := max(selA, sp[0]), min(selB, sp[1])
			x0 := c.measure(c.disp(e, sp[0], a, pw))
			x1 := c.measure(c.disp(e, sp[0], b, pw))
			if selB > sp[1] && multi {
				x1 += 4
			}
			dl.AddRectFilled(Vec2{X: ox + x0, Y: y}, Vec2{X: ox + x1, Y: y + lh}, c.col(ColTextSelectedBg), 0)
		}
		if sp[1] > sp[0] {
			c.DrawText(Vec2{X: ox, Y: y}, e.lines[i], c.col(ColText))
		}
	}
	if focused && !ro && int(e.blink/caretBlink)%2 == 0 {
		y := oy + float32(cl)*lh
		dl.AddLine(Vec2{X: ox + cx, Y: y}, Vec2{X: ox + cx, Y: y + lh}, c.col(ColCaret), 1)
	}
	dl.PopClipRect()

	if multi && total > visH {
		track := Rect{Min: Vec2{X: frame.Max.X - 4, Y: frame.Min.Y + 1}, Max: Vec2{X: frame.Max.X - 1, Y: frame.Max.Y - 1}}
		th := track.Max.Y - track.Min.Y
		gh := max(th*visH/total, st.GrabMinSize)
		gy := track.Min.Y + (th-gh)*e.scrollY/max(total-visH, 1)
		dl.AddRectFilled(Vec2{X: track.Min.X, Y: gy}, Vec2{X: track.Max.X, Y: gy + gh}, c.col(ColScrollbarGrab), 1)
	}
}
