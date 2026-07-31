package components

import (
	"strings"
	"unicode"

	"github.com/Zyko0/go-sdl3/sdl"
)

type textEditorState struct {
	text    []rune
	cursor  int
	anchor  int
	focused bool
}

func newTextEditorState(text string, multiline bool) textEditorState {
	state := textEditorState{text: normalizedText(text, multiline)}
	state.cursor = len(state.text)
	state.anchor = state.cursor
	return state
}

func normalizedText(text string, multiline bool) []rune {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if !multiline {
		text = strings.ReplaceAll(text, "\n", " ")
	}
	return []rune(text)
}

func (e *textEditorState) string() string {
	return string(e.text)
}

func (e *textEditorState) setText(text string, multiline bool) (previous string, changed bool) {
	previous = e.string()
	next := normalizedText(text, multiline)
	if string(next) == previous {
		return previous, false
	}
	e.text = next
	e.cursor = min(e.cursor, len(e.text))
	e.anchor = e.cursor
	return previous, true
}

func (e *textEditorState) selection() (start, end int, ok bool) {
	if e.cursor == e.anchor {
		return 0, 0, false
	}
	return min(e.cursor, e.anchor), max(e.cursor, e.anchor), true
}

func (e *textEditorState) selectedText() string {
	start, end, ok := e.selection()
	if !ok {
		return ""
	}
	return string(e.text[start:end])
}

func (e *textEditorState) replaceSelection(value []rune) (previous string, changed bool) {
	previous = e.string()
	start, end, selected := e.selection()
	if !selected {
		start, end = e.cursor, e.cursor
	}
	if start == end && len(value) == 0 {
		return previous, false
	}
	next := make([]rune, 0, len(e.text)-(end-start)+len(value))
	next = append(next, e.text[:start]...)
	next = append(next, value...)
	next = append(next, e.text[end:]...)
	e.text = next
	e.cursor = start + len(value)
	e.anchor = e.cursor
	return previous, e.string() != previous
}

func (e *textEditorState) moveCursor(position int, selecting bool) {
	position = max(0, min(position, len(e.text)))
	if !selecting {
		e.anchor = position
	}
	e.cursor = position
}

func (e *textEditorState) moveWord(direction int) int {
	position := e.cursor
	if direction < 0 {
		for position > 0 && unicode.IsSpace(e.text[position-1]) {
			position--
		}
		for position > 0 && !unicode.IsSpace(e.text[position-1]) {
			position--
		}
		return position
	}
	for position < len(e.text) && unicode.IsSpace(e.text[position]) {
		position++
	}
	for position < len(e.text) && !unicode.IsSpace(e.text[position]) {
		position++
	}
	return position
}

func (e *textEditorState) lineRange() (start, end int) {
	start = e.cursor
	for start > 0 && e.text[start-1] != '\n' {
		start--
	}
	end = e.cursor
	for end < len(e.text) && e.text[end] != '\n' {
		end++
	}
	return start, end
}

func (e *textEditorState) verticalPosition(direction int) int {
	lineStart, lineEnd := e.lineRange()
	column := e.cursor - lineStart
	if direction < 0 {
		if lineStart == 0 {
			return e.cursor
		}
		previousEnd := lineStart - 1
		previousStart := previousEnd
		for previousStart > 0 && e.text[previousStart-1] != '\n' {
			previousStart--
		}
		return previousStart + min(column, previousEnd-previousStart)
	}
	if lineEnd >= len(e.text) {
		return e.cursor
	}
	nextStart := lineEnd + 1
	nextEnd := nextStart
	for nextEnd < len(e.text) && e.text[nextEnd] != '\n' {
		nextEnd++
	}
	return nextStart + min(column, nextEnd-nextStart)
}

func (e *textEditorState) handleClipboardKey(key sdl.Keycode, multiline bool) (previous string, changed, handled bool) {
	switch key {
	case sdl.K_A:
		e.anchor = 0
		e.cursor = len(e.text)
		return "", false, true
	case sdl.K_C:
		if selected := e.selectedText(); selected != "" {
			_ = sdl.SetClipboardText(selected)
		}
		return "", false, true
	case sdl.K_X:
		if selected := e.selectedText(); selected != "" {
			_ = sdl.SetClipboardText(selected)
			previous, changed = e.replaceSelection(nil)
		}
		return previous, changed, true
	case sdl.K_V:
		value, err := sdl.GetClipboardText()
		if err != nil {
			return "", false, true
		}
		previous, changed = e.replaceSelection(normalizedText(value, multiline))
		return previous, changed, true
	default:
		return "", false, false
	}
}

func isCommandModifier(mod sdl.Keymod) bool {
	return mod&(sdl.KMOD_CTRL|sdl.KMOD_GUI) != 0
}
