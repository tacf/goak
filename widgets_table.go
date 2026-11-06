package goak

// TableFlags tune BeginTable.
type TableFlags uint16

// Table flags.
const (
	// TableRowBg shades every other row.
	TableRowBg TableFlags = 1 << iota
	// TableBordersV draws lines between columns.
	TableBordersV
	// TableBordersH draws lines between rows.
	TableBordersH
	// TableBordersOuter draws a frame around the table.
	TableBordersOuter
	// TableResizable lets the user drag column borders in the header row.
	TableResizable
)

// TableBorders draws all borders.
const TableBorders = TableBordersV | TableBordersH | TableBordersOuter

// TableColumnFlags tune TableSetupColumn.
type TableColumnFlags uint8

// Table column flags.
const (
	// TableColumnFixed makes the width a size in logical units. Without it
	// the value is a weight that shares the remaining width.
	TableColumnFixed TableColumnFlags = 1 << iota
)

type tableColumn struct {
	label string
	flags TableColumnFlags
	init  float32
	x, w  float32 // offset from the table left and resolved width
}

type tableState struct {
	id     ID
	n      int
	flags  TableFlags
	cols   []tableColumn
	setups int
	// Geometry.
	start   Vec2
	width   float32
	laid    bool
	sizes   []float32 // user resized widths, 0 when not resized
	sizeKey ID
	// Row and column cursor.
	row, col   int
	rowTop     float32
	rowBottom  float32
	header     bool
	bgVertex   int // first vertex of the pending row background, -1 for none
	inCell     bool
	cellMax    Vec2
	savedMax   Vec2
	savedClip  Rect
	savedRight float32
	savedIndnt float32
	hasRows    bool
}

func (t *tableState) cellPad(c *Context) Vec2 {
	return Vec2{X: c.Style.FramePadding.X, Y: 2}
}

// BeginTable starts a table with n columns. It returns false when the
// window is hidden, then EndTable must not be called. outerSize.X of 0 uses
// the available width. Declare columns with TableSetupColumn, then add rows
// with TableNextRow and TableNextColumn. Tables do not scroll, so put long
// ones in a child window.
func (c *Context) BeginTable(id string, n int, flags TableFlags, outerSize Vec2) bool {
	w := c.cur
	if w == nil || w.skip || n <= 0 {
		return false
	}
	tid := c.GetID(id)
	t := &tableState{id: tid, n: n, flags: flags, row: -1, col: -1, bgVertex: -1, start: w.cursor}
	t.cols = make([]tableColumn, n)
	for i := range t.cols {
		t.cols[i].init = 1
	}
	t.sizeKey = hashStr("#TABLESIZES", tid)
	t.sizes = StateGet(c.States, t.sizeKey, []float32(nil))
	if len(t.sizes) != n {
		t.sizes = make([]float32, n)
	}
	t.width = outerSize.X
	if t.width <= 0 {
		t.width = max(c.ContentRegionAvail().X+t.width, 1)
	}
	t.savedMax, t.savedClip, t.savedRight, t.savedIndnt = w.maxPos, w.clip, w.contentRight, w.indent
	c.tables = append(c.tables, t)
	return true
}

// TableSetupColumn names a column and sets its size. Call it after
// BeginTable and before the first row. For a stretch column v is a weight
// (0 means 1), with TableColumnFixed it is a width.
func (c *Context) TableSetupColumn(label string, flags TableColumnFlags, v float32) {
	t := c.table()
	if t.setups >= t.n || t.laid {
		return
	}
	col := &t.cols[t.setups]
	col.label, col.flags = label, flags
	if v > 0 {
		col.init = v
	}
	t.setups++
}

func (c *Context) table() *tableState {
	if len(c.tables) == 0 {
		panic("goak: table call outside BeginTable")
	}
	return c.tables[len(c.tables)-1]
}

// layoutTable resolves column widths once the setup calls are done.
func (c *Context) layoutTable(t *tableState) {
	if t.laid {
		return
	}
	t.laid = true
	var fixed, weights float32
	for i := range t.cols {
		col := &t.cols[i]
		switch {
		case t.sizes[i] > 0:
			fixed += t.sizes[i]
		case col.flags&TableColumnFixed != 0:
			fixed += col.init
		default:
			weights += col.init
		}
	}
	rest := max(t.width-fixed, 0)
	x := float32(0)
	for i := range t.cols {
		col := &t.cols[i]
		switch {
		case t.sizes[i] > 0:
			col.w = t.sizes[i]
		case col.flags&TableColumnFixed != 0:
			col.w = col.init
		case weights > 0:
			col.w = rest * col.init / weights
		}
		col.w = max(col.w, 8)
		col.x = x
		x += col.w
	}
}

// TableNextRow ends the current row and starts a new one.
func (c *Context) TableNextRow() { c.tableRow(false) }

func (c *Context) tableRow(header bool) {
	t := c.table()
	c.layoutTable(t)
	w := c.cur
	c.endCell(t)
	if t.row >= 0 {
		c.endRow(t)
	}
	t.row++
	t.col = -1
	t.header = header
	t.hasRows = true
	pad := t.cellPad(c)
	if t.row == 0 {
		t.rowTop = t.start.Y
	} else {
		t.rowTop = t.rowBottom
	}
	t.rowBottom = t.rowTop + c.lineHeight() + 2*pad.Y
	t.bgVertex = -1
	dl := w.DL
	left, right := t.start.X, t.start.X+t.width
	switch {
	case header:
		dl.AddRectFilled(Vec2{X: left, Y: t.rowTop}, Vec2{X: right, Y: t.rowBottom}, c.col(ColTableHeaderBg), 0)
	case t.flags&TableRowBg != 0 && t.row%2 == 0:
		before := len(dl.Vertices)
		dl.AddRectFilled(Vec2{X: left, Y: t.rowTop}, Vec2{X: right, Y: t.rowTop + 1}, c.col(ColTableRowBgAlt), 0)
		if len(dl.Vertices) == before+4 {
			t.bgVertex = before
		}
	}
}

// endRow finishes the current row, stretching its background to the final
// height and drawing the line below it.
func (c *Context) endRow(t *tableState) {
	w := c.cur
	dl := w.DL
	if t.bgVertex >= 0 && t.bgVertex+3 < len(dl.Vertices) {
		dl.Vertices[t.bgVertex+2].Pos.Y = t.rowBottom
		dl.Vertices[t.bgVertex+3].Pos.Y = t.rowBottom
	}
	t.bgVertex = -1
	if t.flags&TableBordersH != 0 {
		y := t.rowBottom - 0.5
		dl.AddLine(Vec2{X: t.start.X, Y: y}, Vec2{X: t.start.X + t.width, Y: y}, c.col(ColTableBorder), 1)
	}
}

// TableNextColumn moves to the next column, starting a new row after the
// last one. It reports whether the column is visible.
func (c *Context) TableNextColumn() bool {
	t := c.table()
	if t.row < 0 || t.col >= t.n-1 {
		c.tableRow(false)
	}
	return c.enterCell(t, t.col+1)
}

// TableSetColumnIndex moves to column i of the current row.
func (c *Context) TableSetColumnIndex(i int) bool {
	t := c.table()
	if t.row < 0 {
		c.tableRow(false)
	}
	if i < 0 || i >= t.n {
		return false
	}
	return c.enterCell(t, i)
}

func (c *Context) enterCell(t *tableState, i int) bool {
	c.endCell(t)
	w := c.cur
	t.col = i
	col := t.cols[i]
	pad := t.cellPad(c)
	x0 := t.start.X + col.x
	cell := Rect{Min: Vec2{X: x0, Y: w.clip.Min.Y}, Max: Vec2{X: x0 + col.w, Y: w.clip.Max.Y}}
	t.savedClip = w.clip
	w.clip = w.clip.Intersect(cell)
	w.DL.PushClipRect(cell)
	t.savedMax = w.maxPos
	w.maxPos = Vec2{X: x0, Y: t.rowTop}
	t.savedRight, t.savedIndnt = w.contentRight, w.indent
	w.indent = x0 + pad.X - w.contentMin.X
	w.contentRight = x0 + col.w - pad.X
	w.cursor = Vec2{X: w.lineStartX(), Y: t.rowTop + pad.Y}
	w.onSameLine = false
	w.nextWidthSet = false
	w.itemWidths = w.itemWidths[:0]
	t.inCell = true
	return true
}

// endCell leaves the current column and records how tall it was.
func (c *Context) endCell(t *tableState) {
	if !t.inCell {
		return
	}
	w := c.cur
	t.inCell = false
	w.DL.PopClipRect()
	w.clip = t.savedClip
	pad := t.cellPad(c)
	t.rowBottom = max(t.rowBottom, w.maxPos.Y+pad.Y)
	w.maxPos = t.savedMax.Max(w.maxPos)
	w.contentRight, w.indent = t.savedRight, t.savedIndnt
}

// TableHeadersRow adds a row with the column names. With TableResizable the
// borders between headers can be dragged.
func (c *Context) TableHeadersRow() {
	t := c.table()
	c.tableRow(true)
	w := c.cur
	for i := 0; i < t.n; i++ {
		c.enterCell(t, i)
		c.DrawText(w.cursor, DisplayLabel(t.cols[i].label), c.col(ColText))
		w.maxPos = w.maxPos.Max(Vec2{X: w.cursor.X, Y: t.rowTop + c.lineHeight()})
		c.endCell(t)
		if t.flags&TableResizable != 0 && i < t.n-1 {
			c.tableResizeHandle(t, i)
		}
	}
}

// tableResizeHandle lets the user drag the right border of column i.
func (c *Context) tableResizeHandle(t *tableState, i int) {
	col := t.cols[i]
	x := t.start.X + col.x + col.w
	r := Rect{Min: Vec2{X: x - 3, Y: t.rowTop}, Max: Vec2{X: x + 3, Y: t.rowBottom}}
	id := hashUint(0x7B, uint64(i), t.id)
	s := c.ButtonBehaviorEx(r, id, BtnNoFocus)
	if c.activeID == id && s.Held {
		t.sizes[i] = max(c.mouse().X-(t.start.X+col.x), 20)
		StateSetKeep(c.States, t.sizeKey, t.sizes, LongAge)
		// Following columns shift now, the next frame lays them out.
		x = t.start.X + col.x + t.sizes[i]
	}
	if s.Hovered || s.Held {
		c.cur.DL.AddLine(Vec2{X: x, Y: t.rowTop}, Vec2{X: x, Y: t.rowBottom}, c.col(ColCheckMark), 2)
	}
}

// EndTable ends BeginTable.
func (c *Context) EndTable() {
	t := c.table()
	w := c.cur
	c.layoutTable(t)
	c.endCell(t)
	if t.row >= 0 {
		c.endRow(t)
	}
	bottom := t.start.Y
	if t.hasRows {
		bottom = t.rowBottom
	}
	dl := w.DL
	border := c.col(ColTableBorder)
	if t.flags&TableBordersV != 0 {
		for i := 1; i < t.n; i++ {
			x := t.start.X + t.cols[i].x - 0.5
			dl.AddLine(Vec2{X: x, Y: t.start.Y}, Vec2{X: x, Y: bottom}, border, 1)
		}
	}
	if t.flags&TableBordersOuter != 0 {
		dl.AddRect(t.start, Vec2{X: t.start.X + t.width, Y: bottom}, border, 0, 1)
	}
	c.tables = c.tables[:len(c.tables)-1]
	w.cursor = t.start
	w.onSameLine = false
	c.ItemAdd(Vec2{X: t.width, Y: bottom - t.start.Y})
}
