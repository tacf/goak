package goak

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// Window wraps the runtime window management.
type Window struct {
	title                string
	width                int
	height               int
	autoDPI              bool
	windowScale          float64
	scaleHotkeys         bool
	onWindowScaleChanged func(float64)
	debugMode            bool
	hoveredRect          layout.Rect
	hasHoveredRect       bool

	ui *components.UI

	canvas *ebiten.Image
}

// Config holds window options.
type Config struct {
	Title       string
	Width       int
	Height      int
	AutoDPI     bool
	WindowScale float64
}

// InitWindow creates and configures a new window with the given title and size.
func InitWindow(title string, width, height int) *Window {
	return New(Config{
		Title:       title,
		Width:       width,
		Height:      height,
		AutoDPI:     false,
		WindowScale: 1,
	})
}

// New creates and configures a new window. Call Destroy when done.
func New(cfg Config) *Window {
	return &Window{
		title:       cfg.Title,
		width:       cfg.Width,
		height:      cfg.Height,
		autoDPI:     cfg.AutoDPI,
		windowScale: normalizeScale(cfg.WindowScale),
	}
}

func (win *Window) attachUI(ui *components.UI) {
	win.ui = ui
}

// SetTitle updates the window title.
func (win *Window) SetTitle(title string) {
	win.title = title
	ebiten.SetWindowTitle(title)
}

// SetAutoDPI toggles automatic HiDPI scaling using ebiten.DeviceScaleFactor.
func (win *Window) SetAutoDPI(enabled bool) {
	win.autoDPI = enabled
}

// AutoDPI reports whether automatic HiDPI scaling is enabled.
func (win *Window) AutoDPI() bool {
	return win.autoDPI
}

// SetWindowScale sets additional runtime scale applied on top of root scale
// and optional device scale. Values <= 0 are ignored.
func (win *Window) SetWindowScale(scale float64) {
	next := normalizeScale(scale)
	if next == win.windowScale {
		return
	}
	win.windowScale = next
	if win.onWindowScaleChanged != nil {
		win.onWindowScaleChanged(next)
	}
}

// WindowScale returns the runtime window scale multiplier.
func (win *Window) WindowScale() float64 {
	if win.windowScale <= 0 {
		return 1
	}
	return win.windowScale
}

// SetScaleHotkeysEnabled toggles built-in Ctrl +/- scale shortcuts.
func (win *Window) SetScaleHotkeysEnabled(enabled bool) {
	win.scaleHotkeys = enabled
}

// ScaleHotkeysEnabled reports whether built-in scale shortcuts are enabled.
func (win *Window) ScaleHotkeysEnabled() bool {
	return win.scaleHotkeys
}

// SetOnWindowScaleChanged sets a callback invoked when WindowScale changes.
func (win *Window) SetOnWindowScaleChanged(fn func(float64)) {
	win.onWindowScaleChanged = fn
}

// Run runs the window event loop until the window is closed.
func (win *Window) Run() {
	ebiten.SetWindowTitle(win.title)
	ebiten.SetWindowSize(win.width, win.height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	_ = ebiten.RunGame(win)
}

// Destroy closes the window and frees any additional resources.
func (win *Window) Destroy() {}

// Update handles input and layout.
func (win *Window) Update() error {
	if win.ui == nil {
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		win.debugMode = !win.debugMode
	}

	if win.scaleHotkeys {
		win.handleScaleHotkeys()
	}

	w, h := windowSize(win.width, win.height)

	root := win.ui.Root()
	scale := win.effectiveScale(root.Scale)

	for _, m := range win.ui.MenuBars() {
		m.SyncWidth()
	}

	layout.Layout(root.Container(), float64(w)/scale, float64(h)/scale)

	mx, my := ebiten.CursorPosition()
	lx := float64(mx) / scale
	ly := float64(my) / scale
	for _, m := range win.ui.MenuBars() {
		m.OnMouseMove(lx, ly)
	}
	win.updateHoveredElement(lx, ly)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		consumed := false
		for _, m := range win.ui.MenuBars() {
			if m.OnMouseDown(lx, ly) {
				consumed = true
				break
			}
		}
		if !consumed {
			for i, b := range win.ui.Buttons() {
				bound := b.Bounds()
				if lx >= bound.X && lx < bound.X+bound.W && ly >= bound.Y && ly < bound.Y+bound.H {
					win.ui.ButtonClicked(i)
					break
				}
			}
		}
	}

	return nil
}

// Draw renders the UI by delegating to each component's draw methods.
func (win *Window) Draw(screen *ebiten.Image) {
	if win.ui == nil {
		return
	}

	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	if screenW <= 0 || screenH <= 0 {
		return
	}

	root := win.ui.Root()
	scale := win.effectiveScale(root.Scale)

	logicalW := int(float64(screenW) / scale)
	logicalH := int(float64(screenH) / scale)
	if logicalW <= 0 {
		logicalW = 1
	}
	if logicalH <= 0 {
		logicalH = 1
	}

	if win.canvas == nil || win.canvas.Bounds().Dx() != logicalW || win.canvas.Bounds().Dy() != logicalH {
		win.canvas = ebiten.NewImage(logicalW, logicalH)
	}

	bg := colors.Black
	win.canvas.Fill(bg)

	panelTheme := components.DefaultPanelTheme()
	buttonTheme := components.DefaultButtonTheme()
	menuTheme := components.DefaultMenuTheme()
	face := basicfont.Face7x13

	for _, p := range win.ui.Panels() {
		p.Draw(win.canvas, panelTheme)
	}

	for _, b := range win.ui.Buttons() {
		b.Draw(win.canvas, face, buttonTheme)
	}

	// Layer 1: menu bar strips and top-level labels.
	for _, m := range win.ui.MenuBars() {
		m.DrawBar(win.canvas, face, menuTheme)
	}

	// Layer 2 (top-most): dropdowns always render above everything else.
	for _, m := range win.ui.MenuBars() {
		m.DrawDropdown(win.canvas, face, menuTheme)
	}

	if win.debugMode {
		if win.hasHoveredRect {
			drawDebugOutline(win.canvas, win.hoveredRect, colors.Yellow)
		}
		const label = "Debug Mode"
		lw := font.MeasureString(face, label).Ceil()
		lh := face.Metrics().Height.Ceil()
		const margin = 8
		x := logicalW - lw - margin
		y := logicalH - margin
		if x < margin {
			x = margin
		}
		if y < lh+margin {
			y = lh + margin
		}
		text.Draw(win.canvas, label, face, x, y, colors.Yellow)
	}

	screen.Fill(bg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	screen.DrawImage(win.canvas, op)
}

// Layout lets Ebiten adjust to the outside window size.
func (win *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 {
		outsideWidth = win.width
	}
	if outsideHeight <= 0 {
		outsideHeight = win.height
	}
	return outsideWidth, outsideHeight
}

func (win *Window) effectiveScale(rootScale float64) float64 {
	scale := normalizeScale(rootScale)
	scale *= win.WindowScale()
	if win.autoDPI {
		dpi := ebiten.DeviceScaleFactor()
		if dpi > 0 {
			scale *= dpi
		}
	}
	return scale
}

func (win *Window) handleScaleHotkeys() {
	if !isCtrlPressed() {
		return
	}
	step := 0.1
	minScale := 0.5
	maxScale := 4.0
	cur := win.WindowScale()
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		next := cur + step
		if next > maxScale {
			next = maxScale
		}
		win.SetWindowScale(next)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		next := cur - step
		if next < minScale {
			next = minScale
		}
		win.SetWindowScale(next)
	}
}

func isCtrlPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
}

func normalizeScale(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

func windowSize(fallbackW, fallbackH int) (int, int) {
	w, h := ebiten.WindowSize()
	if w <= 0 {
		w = fallbackW
	}
	if h <= 0 {
		h = fallbackH
	}
	return w, h
}

func (win *Window) updateHoveredElement(x, y float64) {
	win.hasHoveredRect = false
	if !win.debugMode || win.ui == nil {
		return
	}

	menus := win.ui.MenuBars()
	for i := len(menus) - 1; i >= 0; i-- {
		m := menus[i]
		if m.IsOpen() {
			subRects := m.OpenSubItemRects()
			for j := len(subRects) - 1; j >= 0; j-- {
				if pointInRect(x, y, subRects[j]) {
					win.hoveredRect = subRects[j]
					win.hasHoveredRect = true
					return
				}
			}
		}
	}
	for i := len(menus) - 1; i >= 0; i-- {
		m := menus[i]
		topRects := m.TopItemRects()
		for j := len(topRects) - 1; j >= 0; j-- {
			if pointInRect(x, y, topRects[j]) {
				win.hoveredRect = topRects[j]
				win.hasHoveredRect = true
				return
			}
		}
		if pointInRect(x, y, m.Bounds()) {
			win.hoveredRect = m.Bounds()
			win.hasHoveredRect = true
			return
		}
	}

	buttons := win.ui.Buttons()
	for i := len(buttons) - 1; i >= 0; i-- {
		if pointInRect(x, y, buttons[i].Bounds()) {
			win.hoveredRect = buttons[i].Bounds()
			win.hasHoveredRect = true
			return
		}
	}
	panels := win.ui.Panels()
	for i := len(panels) - 1; i >= 0; i-- {
		if pointInRect(x, y, panels[i].Bounds()) {
			win.hoveredRect = panels[i].Bounds()
			win.hasHoveredRect = true
			return
		}
	}
}

func drawDebugOutline(dst *ebiten.Image, r layout.Rect, c colors.Color) {
	const t = 2.0
	ebitenutil.DrawRect(dst, r.X, r.Y, r.W, t, c)
	ebitenutil.DrawRect(dst, r.X, r.Y+r.H-t, r.W, t, c)
	ebitenutil.DrawRect(dst, r.X, r.Y, t, r.H, c)
	ebitenutil.DrawRect(dst, r.X+r.W-t, r.Y, t, r.H, c)
}

func pointInRect(x, y float64, r layout.Rect) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}
