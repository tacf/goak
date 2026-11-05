// Package textedit holds the rune buffer, cursor and selection logic of a
// single or multiline text field. It has no UI dependency.
package textedit

import (
	"strings"
	"unicode"
)

// Buffer is the editable text with a cursor and selection anchor.
type Buffer struct {
	Text   []rune
	Cursor int
	Anchor int
	Sync   string // last string seen or written
	Rev    int    // bumped by every text change
	Reveal bool   // scroll the caret into view on the next draw
	Fresh  bool
}

// Normalize converts line endings to \n and, unless multiline, newlines to
// spaces.
func Normalize(s string, multiline bool) []rune {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if !multiline {
		s = strings.ReplaceAll(s, "\n", " ")
	}
	return []rune(s)
}

// CleanPaste normalizes s and drops control runes, keeping newlines only
// for multiline text.
func CleanPaste(s string, multi bool) []rune {
	in := Normalize(s, multi)
	out := in[:0]
	for _, r := range in {
		if r == '\n' && multi || !unicode.IsControl(r) {
			out = append(out, r)
		}
	}
	return out
}

// Load adopts s when it differs from what the editor last wrote.
func (e *Buffer) Load(s string, multiline bool) {
	if e.Fresh && s == e.Sync {
		return
	}
	e.Fresh = true
	e.Rev++
	e.Sync = s
	e.Text = Normalize(s, multiline)
	e.Cursor = min(e.Cursor, len(e.Text))
	e.Anchor = min(e.Anchor, len(e.Text))
}

// Selection returns the ordered selection range and whether one exists.
func (e *Buffer) Selection() (start, end int, ok bool) {
	if e.Cursor == e.Anchor {
		return 0, 0, false
	}
	return min(e.Cursor, e.Anchor), max(e.Cursor, e.Anchor), true
}

// SelectedText returns the selected text, or "".
func (e *Buffer) SelectedText() string {
	if s, t, ok := e.Selection(); ok {
		return string(e.Text[s:t])
	}
	return ""
}

// ReplaceSelection swaps the selection, or inserts at the cursor.
func (e *Buffer) ReplaceSelection(value []rune) bool {
	start, end, sel := e.Selection()
	if !sel {
		start, end = e.Cursor, e.Cursor
	}
	if start == end && len(value) == 0 {
		return false
	}
	next := make([]rune, 0, len(e.Text)-(end-start)+len(value))
	next = append(next, e.Text[:start]...)
	next = append(next, value...)
	next = append(next, e.Text[end:]...)
	e.Text = next
	e.Rev++
	e.Cursor = start + len(value)
	e.Anchor = e.Cursor
	e.Reveal = true
	return true
}

// MoveCursor clamps pos and moves the cursor, keeping the anchor when
// selecting.
func (e *Buffer) MoveCursor(pos int, selecting bool) {
	pos = max(0, min(pos, len(e.Text)))
	if !selecting {
		e.Anchor = pos
	}
	e.Cursor = pos
	e.Reveal = true
}

// MoveWord returns the start of the previous (dir < 0) or end of the next
// word from the cursor.
func (e *Buffer) MoveWord(dir int) int {
	p := e.Cursor
	if dir < 0 {
		for p > 0 && unicode.IsSpace(e.Text[p-1]) {
			p--
		}
		for p > 0 && !unicode.IsSpace(e.Text[p-1]) {
			p--
		}
		return p
	}
	for p < len(e.Text) && unicode.IsSpace(e.Text[p]) {
		p++
	}
	for p < len(e.Text) && !unicode.IsSpace(e.Text[p]) {
		p++
	}
	return p
}

// LineRange is the rune range of the cursor line.
func (e *Buffer) LineRange() (start, end int) {
	start = e.Cursor
	for start > 0 && e.Text[start-1] != '\n' {
		start--
	}
	end = e.Cursor
	for end < len(e.Text) && e.Text[end] != '\n' {
		end++
	}
	return start, end
}

// VerticalPosition is the cursor index one line up or down.
func (e *Buffer) VerticalPosition(dir int) int {
	ls, le := e.LineRange()
	col := e.Cursor - ls
	if dir < 0 {
		if ls == 0 {
			return e.Cursor
		}
		pe := ls - 1
		ps := pe
		for ps > 0 && e.Text[ps-1] != '\n' {
			ps--
		}
		return ps + min(col, pe-ps)
	}
	if le >= len(e.Text) {
		return e.Cursor
	}
	ns := le + 1
	ne := ns
	for ne < len(e.Text) && e.Text[ne] != '\n' {
		ne++
	}
	return ns + min(col, ne-ns)
}

// WordAt returns the word range around index i.
func (e *Buffer) WordAt(i int) (start, end int) {
	i = max(0, min(i, len(e.Text)))
	isWord := func(r rune) bool { return !unicode.IsSpace(r) }
	start, end = i, i
	if i < len(e.Text) && isWord(e.Text[i]) || i > 0 && isWord(e.Text[i-1]) {
		for start > 0 && isWord(e.Text[start-1]) {
			start--
		}
		for end < len(e.Text) && isWord(e.Text[end]) {
			end++
		}
	}
	return start, end
}
