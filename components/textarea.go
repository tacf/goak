package components

import (
	"strconv"
	"unicode"

	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

const (
	textAreaPadding   = 6.0
	scrollbarSize     = 9.0
	textAreaLineGap   = 3.0
	wheelScrollLines  = 3.0
	minScrollbarThumb = 18.0
)

// TextArea is an editable multiline text box with optional wrapping and line
// numbers. It supports vertical scrolling and horizontal scrolling when wrap
// is disabled.
type TextArea struct {
	Element
	c            *layout.Container
	editor       textEditorState
	placeholder  string
	wrap         bool
	lineNumbers  bool
	scrollX      float64
	scrollY      float64
	revealCursor bool
	onChanged    func(TextAreaChangedEvent)
}

// NewTextArea creates a standalone multiline text area.
func NewTextArea(width, height layout.Size, text string) *TextArea {
	area := &TextArea{
		c:            layout.NewContainer(width, height),
		editor:       newTextEditorState(text, true),
		wrap:         true,
		lineNumbers:  true,
		revealCursor: true,
	}
	area.Element.init(area)
	return area
}

// Container returns the text area's underlying layout node.
func (a *TextArea) Container() *layout.Container { return a.c }

// Bounds returns the computed text area bounds.
func (a *TextArea) Bounds() layout.Rect { return a.c.Bounds }

// Text returns the current value.
func (a *TextArea) Text() string { return a.editor.string() }

// SetText replaces the text area value.
func (a *TextArea) SetText(text string) {
	previous, changed := a.editor.setText(text, true)
	if changed {
		a.revealCursor = true
		a.emitChanged(previous)
	}
}

// Placeholder returns the hint drawn when the text area is empty.
func (a *TextArea) Placeholder() string { return a.placeholder }

// SetPlaceholder updates the empty-value hint.
func (a *TextArea) SetPlaceholder(placeholder string) { a.placeholder = placeholder }

// Wrap reports whether long logical lines are visually wrapped.
func (a *TextArea) Wrap() bool { return a.wrap }

// SetWrap enables or disables visual line wrapping. Horizontal scrolling is
// reset when wrapping is enabled.
func (a *TextArea) SetWrap(wrap bool) {
	a.wrap = wrap
	a.revealCursor = true
	if wrap {
		a.scrollX = 0
	}
}

// LineNumbers reports whether logical line numbers are visible.
func (a *TextArea) LineNumbers() bool { return a.lineNumbers }

// SetLineNumbers shows or hides logical line numbers.
func (a *TextArea) SetLineNumbers(visible bool) {
	a.lineNumbers = visible
	a.revealCursor = true
}

// Cursor returns the cursor position as a rune index.
func (a *TextArea) Cursor() int { return a.editor.cursor }

// SetCursor moves the cursor to a rune index and clears the selection.
func (a *TextArea) SetCursor(position int) {
	a.editor.moveCursor(position, false)
	a.revealCursor = true
}

// Focused reports whether the text area has keyboard focus.
func (a *TextArea) Focused() bool { return a.editor.focused }

// SetFocused changes keyboard focus state.
func (a *TextArea) SetFocused(focused bool) {
	a.editor.focused = focused
	if focused {
		a.revealCursor = true
	}
}

// ScrollOffsets returns horizontal and vertical content offsets.
func (a *TextArea) ScrollOffsets() (x, y float64) { return a.scrollX, a.scrollY }

// SetScrollOffsets changes horizontal and vertical content offsets. Values are
// clamped to non-negative immediately and to content bounds during drawing.
func (a *TextArea) SetScrollOffsets(x, y float64) {
	if a.wrap {
		x = 0
	}
	a.scrollX = max(0, x)
	a.scrollY = max(0, y)
}

// ScrollBy moves the viewport by logical pixels.
func (a *TextArea) ScrollBy(dx, dy float64) {
	a.SetScrollOffsets(a.scrollX+dx, a.scrollY+dy)
}

// ScrollWheel applies SDL wheel deltas.
func (a *TextArea) ScrollWheel(x, y float64, font *rendering.Font) {
	lineHeight := 20.0
	if font != nil {
		lineHeight = font.Height() + textAreaLineGap
	}
	if x != 0 {
		a.ScrollBy(-x*lineHeight*wheelScrollLines, 0)
	}
	a.ScrollBy(0, -y*lineHeight*wheelScrollLines)
}

// SetOnChanged installs a callback for value changes.
func (a *TextArea) SetOnChanged(onChanged func(TextAreaChangedEvent)) {
	a.onChanged = onChanged
}

// HandleTextInput inserts Unicode text at the cursor.
func (a *TextArea) HandleTextInput(text string) {
	previous, changed := a.editor.replaceSelection(normalizedText(text, true))
	if changed {
		a.revealCursor = true
		a.emitChanged(previous)
	}
}

// HandleKey applies an SDL editing key. It returns whether the key was consumed.
func (a *TextArea) HandleKey(key sdl.Keycode, mod sdl.Keymod) bool {
	selecting := mod&sdl.KMOD_SHIFT != 0
	if isCommandModifier(mod) {
		previous, changed, handled := a.editor.handleClipboardKey(key, true)
		if changed {
			a.revealCursor = true
			a.emitChanged(previous)
		}
		if handled {
			a.revealCursor = true
			return true
		}
	}

	switch key {
	case sdl.K_LEFT:
		position := a.editor.cursor - 1
		if isCommandModifier(mod) {
			position = a.editor.moveWord(-1)
		}
		a.editor.moveCursor(position, selecting)
	case sdl.K_RIGHT:
		position := a.editor.cursor + 1
		if isCommandModifier(mod) {
			position = a.editor.moveWord(1)
		}
		a.editor.moveCursor(position, selecting)
	case sdl.K_UP:
		a.editor.moveCursor(a.editor.verticalPosition(-1), selecting)
	case sdl.K_DOWN:
		a.editor.moveCursor(a.editor.verticalPosition(1), selecting)
	case sdl.K_HOME:
		start, _ := a.editor.lineRange()
		if isCommandModifier(mod) {
			start = 0
		}
		a.editor.moveCursor(start, selecting)
	case sdl.K_END:
		_, end := a.editor.lineRange()
		if isCommandModifier(mod) {
			end = len(a.editor.text)
		}
		a.editor.moveCursor(end, selecting)
	case sdl.K_BACKSPACE:
		if _, _, selected := a.editor.selection(); !selected && a.editor.cursor > 0 {
			a.editor.anchor = a.editor.cursor - 1
		}
		previous, changed := a.editor.replaceSelection(nil)
		if changed {
			a.emitChanged(previous)
		}
	case sdl.K_DELETE:
		if _, _, selected := a.editor.selection(); !selected && a.editor.cursor < len(a.editor.text) {
			a.editor.anchor = a.editor.cursor + 1
		}
		previous, changed := a.editor.replaceSelection(nil)
		if changed {
			a.emitChanged(previous)
		}
	case sdl.K_RETURN, sdl.K_RETURN2, sdl.K_KP_ENTER:
		previous, changed := a.editor.replaceSelection([]rune{'\n'})
		if changed {
			a.emitChanged(previous)
		}
	case sdl.K_TAB:
		previous, changed := a.editor.replaceSelection([]rune{'\t'})
		if changed {
			a.emitChanged(previous)
		}
	default:
		return false
	}
	a.revealCursor = true
	return true
}

// SetCursorAt moves the cursor to the character nearest a logical point.
func (a *TextArea) SetCursorAt(x, y float64, font *rendering.Font) {
	if font == nil {
		return
	}
	geometry := a.geometry(font)
	lines := visualLines(a.editor.text, a.wrap, geometry.text.W, font)
	if len(lines) == 0 {
		a.editor.moveCursor(0, false)
		return
	}
	row := int((y - geometry.text.Y + a.scrollY) / geometry.lineHeight)
	row = max(0, min(row, len(lines)-1))
	line := lines[row]
	target := x - geometry.text.X + a.scrollX
	position := line.start + nearestRune(a.editor.text[line.start:line.end], target, font)
	a.editor.moveCursor(position, false)
	a.revealCursor = true
}

func (a *TextArea) emitChanged(previous string) {
	if a.onChanged != nil {
		a.onChanged(TextAreaChangedEvent{
			TextArea: a,
			Previous: previous,
			Text:     a.Text(),
		})
	}
}

type textAreaGeometry struct {
	inner      layout.Rect
	text       layout.Rect
	gutter     layout.Rect
	lineHeight float64
}

func (a *TextArea) geometry(font *rendering.Font) textAreaGeometry {
	bounds := a.Bounds()
	inner := layout.Rect{
		X: bounds.X + textAreaPadding,
		Y: bounds.Y + textAreaPadding,
		W: max(0, bounds.W-textAreaPadding*2-scrollbarSize),
		H: max(0, bounds.H-textAreaPadding*2-scrollbarSize),
	}
	gutterW := 0.0
	if a.lineNumbers {
		lineCount := logicalLineCount(a.editor.text)
		numberW, _ := font.Measure(strconv.Itoa(lineCount))
		gutterW = numberW + textAreaPadding*2
	}
	return textAreaGeometry{
		inner: inner,
		gutter: layout.Rect{
			X: inner.X,
			Y: inner.Y,
			W: gutterW,
			H: inner.H,
		},
		text: layout.Rect{
			X: inner.X + gutterW,
			Y: inner.Y,
			W: max(0, inner.W-gutterW),
			H: inner.H,
		},
		lineHeight: font.Height() + textAreaLineGap,
	}
}

// Draw renders the text area, line numbers, selection, caret, and scrollbars.
func (a *TextArea) Draw(renderer *sdl.Renderer, font *rendering.Font, theme TextAreaTheme) {
	if a == nil || renderer == nil || font == nil {
		return
	}
	bounds := a.Bounds()
	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, theme.Fill)
	stroke := theme.Stroke
	if a.Focused() {
		stroke = theme.FocusedStroke
	}
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1, stroke)

	geometry := a.geometry(font)
	lines := visualLines(a.editor.text, a.wrap, geometry.text.W, font)
	contentH := float64(len(lines)) * geometry.lineHeight
	contentW := maxLineWidth(lines, a.editor.text, font)
	a.scrollY = max(0, min(a.scrollY, max(0, contentH-geometry.text.H)))
	if a.wrap {
		a.scrollX = 0
	} else {
		a.scrollX = max(0, min(a.scrollX, max(0, contentW-geometry.text.W)))
	}
	if a.revealCursor {
		a.ensureCaretVisible(lines, geometry, font)
		a.revealCursor = false
		a.scrollY = max(0, min(a.scrollY, max(0, contentH-geometry.text.H)))
		if !a.wrap {
			a.scrollX = max(0, min(a.scrollX, max(0, contentW-geometry.text.W)))
		}
	}

	if a.lineNumbers {
		rendering.FillRect(
			renderer,
			geometry.gutter.X,
			geometry.gutter.Y,
			geometry.gutter.W,
			geometry.gutter.H,
			theme.GutterFill,
		)
		rendering.DrawLine(
			renderer,
			geometry.gutter.X+geometry.gutter.W-1,
			geometry.gutter.Y,
			geometry.gutter.H,
			1,
			theme.Stroke,
			false,
		)
	}

	clip := sdl.Rect{
		X: int32(geometry.inner.X),
		Y: int32(geometry.inner.Y),
		W: int32(geometry.inner.W),
		H: int32(geometry.inner.H),
	}
	_ = renderer.SetClipRect(&clip)
	firstRow := max(0, int(a.scrollY/geometry.lineHeight))
	lastRow := min(len(lines), int((a.scrollY+geometry.text.H)/geometry.lineHeight)+2)
	selectionStart, selectionEnd, hasSelection := a.editor.selection()
	cursorRow := cursorVisualRow(lines, a.editor.cursor)

	for row := firstRow; row < lastRow; row++ {
		line := lines[row]
		rowY := geometry.text.Y + float64(row)*geometry.lineHeight - a.scrollY
		if a.lineNumbers && !line.continuation {
			number := strconv.Itoa(line.logicalLine + 1)
			numberW, _ := font.Measure(number)
			rendering.DrawText(
				renderer,
				number,
				font,
				geometry.gutter.X+geometry.gutter.W-textAreaPadding-numberW,
				rowY,
				theme.LineNumber,
			)
		}
		textX := geometry.text.X - a.scrollX
		if hasSelection {
			start := max(selectionStart, line.start)
			end := min(selectionEnd, line.end)
			if start < end {
				beforeW, _ := font.Measure(string(a.editor.text[line.start:start]))
				selectedW, _ := font.Measure(string(a.editor.text[start:end]))
				rendering.FillRect(renderer, textX+beforeW, rowY, selectedW, font.Height(), theme.Selection)
			}
		}
		rendering.DrawText(
			renderer,
			string(a.editor.text[line.start:line.end]),
			font,
			textX,
			rowY,
			theme.Text,
		)
		if a.Focused() && caretVisible() && row == cursorRow {
			cursor := min(max(a.editor.cursor, line.start), line.end)
			cursorW, _ := font.Measure(string(a.editor.text[line.start:cursor]))
			rendering.FillRect(renderer, textX+cursorW, rowY, 1.5, font.Height(), theme.Caret)
		}
	}
	if a.Text() == "" && !a.Focused() {
		rendering.DrawText(renderer, a.placeholder, font, geometry.text.X, geometry.text.Y, theme.Placeholder)
	}
	_ = renderer.SetClipRect(nil)

	drawTextAreaScrollbars(renderer, bounds, geometry, contentW, contentH, a.scrollX, a.scrollY, theme)
}

func (a *TextArea) ensureCaretVisible(lines []visualLine, geometry textAreaGeometry, font *rendering.Font) {
	if !a.Focused() || len(lines) == 0 {
		return
	}
	row := cursorVisualRow(lines, a.editor.cursor)
	line := lines[row]
	cursor := min(max(a.editor.cursor, line.start), line.end)
	cursorX, _ := font.Measure(string(a.editor.text[line.start:cursor]))
	cursorY := float64(row) * geometry.lineHeight
	if cursorY < a.scrollY {
		a.scrollY = cursorY
	} else if cursorY+geometry.lineHeight > a.scrollY+geometry.text.H {
		a.scrollY = cursorY + geometry.lineHeight - geometry.text.H
	}
	if !a.wrap {
		if cursorX < a.scrollX {
			a.scrollX = cursorX
		} else if cursorX+2 > a.scrollX+geometry.text.W {
			a.scrollX = cursorX + 2 - geometry.text.W
		}
	}
	a.scrollX = max(0, a.scrollX)
	a.scrollY = max(0, a.scrollY)
}

type visualLine struct {
	start        int
	end          int
	logicalLine  int
	continuation bool
}

func visualLines(text []rune, wrap bool, width float64, font *rendering.Font) []visualLine {
	return visualLinesMeasured(text, wrap, width, func(value string) float64 {
		measured, _ := font.Measure(value)
		return measured
	})
}

func visualLinesMeasured(text []rune, wrap bool, width float64, measure func(string) float64) []visualLine {
	lines := make([]visualLine, 0, logicalLineCount(text))
	lineStart := 0
	lineNumber := 0
	for {
		lineEnd := lineStart
		for lineEnd < len(text) && text[lineEnd] != '\n' {
			lineEnd++
		}
		if !wrap || width <= 0 {
			lines = append(lines, visualLine{start: lineStart, end: lineEnd, logicalLine: lineNumber})
		} else {
			lines = appendWrappedLine(lines, text, lineStart, lineEnd, lineNumber, width, measure)
		}
		if lineEnd >= len(text) {
			break
		}
		lineStart = lineEnd + 1
		lineNumber++
	}
	return lines
}

func appendWrappedLine(
	lines []visualLine,
	text []rune,
	start, end, logicalLine int,
	width float64,
	measure func(string) float64,
) []visualLine {
	if start == end {
		return append(lines, visualLine{start: start, end: end, logicalLine: logicalLine})
	}
	segmentStart := start
	continuation := false
	for segmentStart < end {
		segmentEnd := segmentStart
		lastBreak := -1
		for candidate := segmentStart + 1; candidate <= end; candidate++ {
			if measure(string(text[segmentStart:candidate])) > width {
				break
			}
			segmentEnd = candidate
			if unicode.IsSpace(text[candidate-1]) {
				lastBreak = candidate
			}
		}
		if segmentEnd == segmentStart {
			segmentEnd = segmentStart + 1
		} else if segmentEnd < end && lastBreak > segmentStart {
			segmentEnd = lastBreak
		}
		lines = append(lines, visualLine{
			start:        segmentStart,
			end:          segmentEnd,
			logicalLine:  logicalLine,
			continuation: continuation,
		})
		segmentStart = segmentEnd
		continuation = true
	}
	return lines
}

func logicalLineCount(text []rune) int {
	count := 1
	for _, char := range text {
		if char == '\n' {
			count++
		}
	}
	return count
}

func cursorVisualRow(lines []visualLine, cursor int) int {
	row := 0
	for index, line := range lines {
		if cursor >= line.start && cursor <= line.end {
			row = index
		}
		if cursor < line.start {
			break
		}
	}
	return row
}

func maxLineWidth(lines []visualLine, text []rune, font *rendering.Font) float64 {
	var width float64
	for _, line := range lines {
		lineWidth, _ := font.Measure(string(text[line.start:line.end]))
		width = max(width, lineWidth)
	}
	return width
}

func drawTextAreaScrollbars(
	renderer *sdl.Renderer,
	bounds layout.Rect,
	geometry textAreaGeometry,
	contentW, contentH, scrollX, scrollY float64,
	theme TextAreaTheme,
) {
	if contentH > geometry.text.H && geometry.text.H > 0 {
		trackX := bounds.X + bounds.W - scrollbarSize
		trackY := geometry.inner.Y
		trackH := geometry.inner.H
		rendering.FillRect(renderer, trackX, trackY, scrollbarSize, trackH, theme.ScrollbarTrack)
		thumbH := max(minScrollbarThumb, trackH*geometry.text.H/contentH)
		thumbY := trackY
		if contentH > geometry.text.H {
			thumbY += (trackH - thumbH) * scrollY / (contentH - geometry.text.H)
		}
		rendering.FillRect(renderer, trackX+2, thumbY, scrollbarSize-4, thumbH, theme.ScrollbarThumb)
	}
	if contentW > geometry.text.W && geometry.text.W > 0 {
		trackX := geometry.text.X
		trackY := bounds.Y + bounds.H - scrollbarSize
		trackW := geometry.text.W
		rendering.FillRect(renderer, trackX, trackY, trackW, scrollbarSize, theme.ScrollbarTrack)
		thumbW := max(minScrollbarThumb, trackW*geometry.text.W/contentW)
		thumbX := trackX
		if contentW > geometry.text.W {
			thumbX += (trackW - thumbW) * scrollX / (contentW - geometry.text.W)
		}
		rendering.FillRect(renderer, thumbX, trackY+2, thumbW, scrollbarSize-4, theme.ScrollbarThumb)
	}
}
