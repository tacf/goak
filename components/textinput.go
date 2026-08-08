package components

import (
	"time"

	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

const textInputPadding = 7.0

// TextInput is an editable single-line text box.
type TextInput struct {
	c           *layout.Container
	editor      textEditorState
	placeholder string
	scrollX     float64
	onChanged   func(TextInputChangedEvent)
	onSubmitted func(TextInputSubmittedEvent)
}

// NewTextInput creates a standalone single-line text input.
func NewTextInput(width, height layout.Size, text string) *TextInput {
	return &TextInput{
		c:      layout.NewContainer(width, height),
		editor: newTextEditorState(text, false),
	}
}

// Container returns the input's underlying layout node.
func (i *TextInput) Container() *layout.Container { return i.c }

// Bounds returns the computed input bounds.
func (i *TextInput) Bounds() layout.Rect { return i.c.Bounds }

// Text returns the current value.
func (i *TextInput) Text() string { return i.editor.string() }

// SetText replaces the input value.
func (i *TextInput) SetText(text string) {
	previous, changed := i.editor.setText(text, false)
	if changed {
		i.emitChanged(previous)
	}
}

// Placeholder returns the hint drawn when the input is empty.
func (i *TextInput) Placeholder() string { return i.placeholder }

// SetPlaceholder updates the empty-value hint.
func (i *TextInput) SetPlaceholder(placeholder string) { i.placeholder = placeholder }

// Cursor returns the cursor position as a rune index.
func (i *TextInput) Cursor() int { return i.editor.cursor }

// SetCursor moves the cursor to a rune index and clears the selection.
func (i *TextInput) SetCursor(position int) {
	i.editor.moveCursor(position, false)
}

// Focused reports whether the input has keyboard focus.
func (i *TextInput) Focused() bool { return i.editor.focused }

// SetFocused changes keyboard focus state.
func (i *TextInput) SetFocused(focused bool) {
	i.editor.focused = focused
}

// SetOnChanged installs a callback for value changes.
func (i *TextInput) SetOnChanged(onChanged func(TextInputChangedEvent)) {
	i.onChanged = onChanged
}

// SetOnSubmitted installs a callback for Enter/Return.
func (i *TextInput) SetOnSubmitted(onSubmitted func(TextInputSubmittedEvent)) {
	i.onSubmitted = onSubmitted
}

// HandleTextInput inserts Unicode text at the cursor.
func (i *TextInput) HandleTextInput(text string) {
	previous, changed := i.editor.replaceSelection(normalizedText(text, false))
	if changed {
		i.emitChanged(previous)
	}
}

// HandleKey applies an SDL editing key. It returns whether the key was consumed.
func (i *TextInput) HandleKey(key sdl.Keycode, mod sdl.Keymod) bool {
	selecting := mod&sdl.KMOD_SHIFT != 0
	if isCommandModifier(mod) {
		previous, changed, handled := i.editor.handleClipboardKey(key, false)
		if changed {
			i.emitChanged(previous)
		}
		if handled {
			return true
		}
	}

	switch key {
	case sdl.K_LEFT:
		position := i.editor.cursor - 1
		if isCommandModifier(mod) {
			position = i.editor.moveWord(-1)
		}
		i.editor.moveCursor(position, selecting)
	case sdl.K_RIGHT:
		position := i.editor.cursor + 1
		if isCommandModifier(mod) {
			position = i.editor.moveWord(1)
		}
		i.editor.moveCursor(position, selecting)
	case sdl.K_HOME, sdl.K_UP:
		i.editor.moveCursor(0, selecting)
	case sdl.K_END, sdl.K_DOWN:
		i.editor.moveCursor(len(i.editor.text), selecting)
	case sdl.K_BACKSPACE:
		if _, _, selected := i.editor.selection(); !selected && i.editor.cursor > 0 {
			i.editor.anchor = i.editor.cursor - 1
		}
		previous, changed := i.editor.replaceSelection(nil)
		if changed {
			i.emitChanged(previous)
		}
	case sdl.K_DELETE:
		if _, _, selected := i.editor.selection(); !selected && i.editor.cursor < len(i.editor.text) {
			i.editor.anchor = i.editor.cursor + 1
		}
		previous, changed := i.editor.replaceSelection(nil)
		if changed {
			i.emitChanged(previous)
		}
	case sdl.K_RETURN, sdl.K_RETURN2, sdl.K_KP_ENTER:
		if i.onSubmitted != nil {
			i.onSubmitted(TextInputSubmittedEvent{TextInput: i, Text: i.Text()})
		}
	default:
		return false
	}
	return true
}

// SetCursorAt moves the cursor to the character nearest a logical x coordinate.
func (i *TextInput) SetCursorAt(x float64, font *rendering.Font) {
	if font == nil {
		return
	}
	target := x - i.Bounds().X - textInputPadding + i.scrollX
	i.editor.moveCursor(nearestRune(i.editor.text, target, font), false)
}

func (i *TextInput) emitChanged(previous string) {
	if i.onChanged != nil {
		i.onChanged(TextInputChangedEvent{
			TextInput: i,
			Previous:  previous,
			Text:      i.Text(),
		})
	}
}

// Draw renders the input box, selection, text, and caret.
func (i *TextInput) Draw(renderer *sdl.Renderer, font *rendering.Font, theme TextInputTheme) {
	if i == nil || renderer == nil || font == nil {
		return
	}
	bounds := i.Bounds()
	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	stroke := theme.Stroke
	if i.Focused() {
		stroke = theme.FocusedStroke
	}
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1, stroke)

	viewportW := max(0, bounds.W-textInputPadding*2)
	cursorW, _ := font.Measure(string(i.editor.text[:i.editor.cursor]))
	if cursorW-i.scrollX > viewportW {
		i.scrollX = cursorW - viewportW
	}
	if cursorW-i.scrollX < 0 {
		i.scrollX = cursorW
	}
	textW, _ := font.Measure(i.Text())
	i.scrollX = max(0, min(i.scrollX, max(0, textW-viewportW)))

	clip := sdl.Rect{
		X: int32(bounds.X + textInputPadding),
		Y: int32(bounds.Y + 1),
		W: int32(viewportW),
		H: int32(max(0, bounds.H-2)),
	}
	_ = renderer.SetClipRect(&clip)
	textY := bounds.Y + (bounds.H-font.Height())/2
	textX := bounds.X + textInputPadding - i.scrollX
	if i.Text() == "" && !i.Focused() {
		rendering.DrawText(renderer, i.placeholder, font, textX, textY, theme.Placeholder)
	} else {
		if start, end, selected := i.editor.selection(); selected {
			beforeW, _ := font.Measure(string(i.editor.text[:start]))
			selectedW, _ := font.Measure(string(i.editor.text[start:end]))
			rendering.FillRect(renderer, textX+beforeW, textY, selectedW, font.Height(), theme.Selection)
		}
		rendering.DrawText(renderer, i.Text(), font, textX, textY, theme.Text)
		if i.Focused() && caretVisible() {
			rendering.FillRect(renderer, textX+cursorW, textY, 1.5, font.Height(), theme.Caret)
		}
	}
	_ = renderer.SetClipRect(nil)
}

func nearestRune(text []rune, target float64, font *rendering.Font) int {
	if target <= 0 {
		return 0
	}
	var previous float64
	for index := 1; index <= len(text); index++ {
		width, _ := font.Measure(string(text[:index]))
		if target < (previous+width)/2 {
			return index - 1
		}
		previous = width
	}
	return len(text)
}

func caretVisible() bool {
	return time.Now().UnixMilli()%1000 < 550
}
