package goak

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

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

// Window config options
type Config struct {
	Title       string
	Width       int
	Height      int
	AutoDPI     bool
	WindowScale float64
}

func InitWindow(title string, width, height int) *Window {
	return newWindow(Config{
		Title:       title,
		Width:       width,
		Height:      height,
		AutoDPI:     false,
		WindowScale: 1,
	})
}

// Internal function to initialize window
func newWindow(cfg Config) *Window {
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

func (win *Window) SetTitle(title string) {
	win.title = title
	ebiten.SetWindowTitle(title)
}

func (win *Window) SetAutoDPI(enabled bool) {
	win.autoDPI = enabled
}

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

func (win *Window) ScaleHotkeysEnabled() bool {
	return win.scaleHotkeys
}

func (win *Window) SetOnWindowScaleChanged(fn func(float64)) {
	win.onWindowScaleChanged = fn
}

// PointWithinBounds returns true if the point (x, y) is inside the given rectangle.
// This is a convenience wrapper around rendering.PointWithinBounds.
func (win *Window) PointWithinBounds(x, y float64, r layout.Rect) bool {
	return rendering.PointWithinBounds(x, y, r)
}

// Run runs the window event loop until the window is closed.
func (win *Window) Run() {
	ebiten.SetWindowTitle(win.title)
	ebiten.SetWindowSize(win.width, win.height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	_ = ebiten.RunGame(win)
}

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

	for _, s := range win.ui.Sliders() {
		if s.IsDragging() {
			s.UpdateValue(lx)
		}
	}

	for _, rg := range win.ui.RadioGroups() {
		hitIndex := rg.HitTest(lx, ly)
		rg.SetHovered(hitIndex)
	}

	for _, dd := range win.ui.Dropdowns() {
		if dd.IsOpen() {
			hitIndex := dd.HitTestList(lx, ly)
			dd.SetHovered(hitIndex)
		}
	}

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
					consumed = true
					break
				}
			}
		}
		if !consumed {
			for _, cb := range win.ui.Checkboxes() {
				if rendering.PointWithinBounds(lx, ly, cb.Bounds()) {
					cb.Toggle()
					consumed = true
					break
				}
			}
		}
		if !consumed {
			for _, rg := range win.ui.RadioGroups() {
				hitIndex := rg.HitTest(lx, ly)
				if hitIndex >= 0 {
					rg.Select(hitIndex)
					consumed = true
					break
				}
			}
		}
		if !consumed {
			for _, s := range win.ui.Sliders() {
				if rendering.PointWithinBounds(lx, ly, s.Bounds()) {
					s.StartDrag()
					s.UpdateValue(lx)
					consumed = true
					break
				}
			}
		}
		if !consumed {
			for _, dd := range win.ui.Dropdowns() {
				if dd.IsOpen() {
					hitIndex := dd.HitTestList(lx, ly)
					if hitIndex >= 0 {
						dd.Select(hitIndex)
						consumed = true
						break
					}
					// Close dropdown if clicked outside the list
					listBounds := dd.ListBounds()
					if !rendering.PointWithinBounds(lx, ly, listBounds) {
						dd.Close()
						consumed = true
						break
					}
				} else if rendering.PointWithinBounds(lx, ly, dd.Bounds()) {
					dd.Open()
					consumed = true
					break
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		for _, s := range win.ui.Sliders() {
			if s.IsDragging() {
				s.StopDrag()
			}
		}
	}

	return nil
}

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

	checkboxTheme := components.DefaultCheckboxTheme()
	for _, cb := range win.ui.Checkboxes() {
		cb.Draw(win.canvas, face, checkboxTheme, false)
	}

	radioTheme := components.DefaultRadioTheme()
	for _, rg := range win.ui.RadioGroups() {
		rg.Draw(win.canvas, face, radioTheme)
	}

	sliderTheme := components.DefaultSliderTheme()
	for _, s := range win.ui.Sliders() {
		s.Draw(win.canvas, face, sliderTheme)
	}

	dropdownTheme := components.DefaultDropdownTheme()
	for _, dd := range win.ui.Dropdowns() {
		dd.Draw(win.canvas, face, dropdownTheme)
	}

	for _, m := range win.ui.MenuBars() {
		m.DrawBar(win.canvas, face, menuTheme)
	}

	for _, m := range win.ui.MenuBars() {
		m.DrawDropdown(win.canvas, face, menuTheme)
	}

	contextMenuTheme := components.DefaultContextMenuTheme()
	for _, cm := range win.ui.ContextMenus() {
		cm.Draw(win.canvas, face, contextMenuTheme)
	}

	if win.debugMode {
		if win.hasHoveredRect {
			rendering.DrawStrokeRect(win.canvas, win.hoveredRect.X, win.hoveredRect.Y, win.hoveredRect.W, win.hoveredRect.H, 2.0, colors.Yellow)
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
		dpi := ebiten.Monitor().DeviceScaleFactor()
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
				if rendering.PointWithinBounds(x, y, subRects[j]) {
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
			if rendering.PointWithinBounds(x, y, topRects[j]) {
				win.hoveredRect = topRects[j]
				win.hasHoveredRect = true
				return
			}
		}
		if rendering.PointWithinBounds(x, y, m.Bounds()) {
			win.hoveredRect = m.Bounds()
			win.hasHoveredRect = true
			return
		}
	}

	buttons := win.ui.Buttons()
	for i := len(buttons) - 1; i >= 0; i-- {
		if rendering.PointWithinBounds(x, y, buttons[i].Bounds()) {
			win.hoveredRect = buttons[i].Bounds()
			win.hasHoveredRect = true
			return
		}
	}
	panels := win.ui.Panels()
	for i := len(panels) - 1; i >= 0; i-- {
		if rendering.PointWithinBounds(x, y, panels[i].Bounds()) {
			win.hoveredRect = panels[i].Bounds()
			win.hasHoveredRect = true
			return
		}
	}
}
