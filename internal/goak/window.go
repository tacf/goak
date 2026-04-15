package goak

import (
	"bytes"
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

	canvas     *ebiten.Image
	fontSource *text.GoTextFaceSource
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
	fs, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal("error loading font", err)
	}

	return &Window{
		title:       cfg.Title,
		width:       cfg.Width,
		height:      cfg.Height,
		autoDPI:     cfg.AutoDPI,
		windowScale: normalizeScale(cfg.WindowScale),
		fontSource:  fs,
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

// SetWindowScale sets additional runtime scale applied on top of root scale.
// Values <= 0 are ignored.
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

	logicalW, logicalH := win.logicalScreenSize()

	root := win.ui.Root()
	uiScale := win.effectiveUIScale(root.Scale)

	for _, m := range win.ui.MenuBars() {
		m.SyncWidth()
	}

	layout.Layout(root.Container(), float64(logicalW)/uiScale, float64(logicalH)/uiScale)

	mx, my := ebiten.CursorPosition()
	lx := float64(mx) / uiScale
	ly := float64(my) / uiScale
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
	uiScale := win.effectiveUIScale(root.Scale)

	// When there is no UI zoom, draw directly to the screen for the sharpest result.
	if uiScale == 1 {
		win.drawUI(screen)
		return
	}

	logicalW := math.Ceil(float64(screenW) / uiScale)
	logicalH := math.Ceil(float64(screenH) / uiScale)
	if logicalW <= 0 {
		logicalW = 1
	}
	if logicalH <= 0 {
		logicalH = 1
	}

	canvasW := int(logicalW)
	canvasH := int(logicalH)
	if win.canvas == nil || win.canvas.Bounds().Dx() != canvasW || win.canvas.Bounds().Dy() != canvasH {
		win.canvas = ebiten.NewImage(canvasW, canvasH)
	}

	win.drawUI(win.canvas)

	sx := float64(screenW) / logicalW
	sy := float64(screenH) / logicalH
	if sx <= 0 {
		sx = 1
	}
	if sy <= 0 {
		sy = 1
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(win.canvas, op)
}

func (win *Window) drawUI(dst *ebiten.Image) {
	screenW, screenH := dst.Bounds().Dx(), dst.Bounds().Dy()
	logicalW := float64(screenW)
	logicalH := float64(screenH)

	bg := colors.Black
	dst.Fill(bg)

	panelTheme := components.DefaultPanelTheme()
	buttonTheme := components.DefaultButtonTheme()
	menuTheme := components.DefaultMenuTheme()

	face := text.GoTextFace{
		Source: win.fontSource,
		Size:   20,
	}

	for _, p := range win.ui.Panels() {
		p.Draw(dst, panelTheme)
	}

	for _, b := range win.ui.Buttons() {
		b.Draw(dst, face, buttonTheme)
	}

	checkboxTheme := components.DefaultCheckboxTheme()
	for _, cb := range win.ui.Checkboxes() {
		cb.Draw(dst, face, checkboxTheme, false)
	}

	radioTheme := components.DefaultRadioTheme()
	for _, rg := range win.ui.RadioGroups() {
		rg.Draw(dst, face, radioTheme)
	}

	sliderTheme := components.DefaultSliderTheme()
	for _, s := range win.ui.Sliders() {
		s.Draw(dst, face, sliderTheme)
	}

	dropdownTheme := components.DefaultDropdownTheme()
	for _, dd := range win.ui.Dropdowns() {
		dd.Draw(dst, face, dropdownTheme)
	}

	for _, m := range win.ui.MenuBars() {
		m.DrawBar(dst, face, menuTheme)
	}

	for _, m := range win.ui.MenuBars() {
		m.DrawDropdown(dst, face, menuTheme)
	}

	contextMenuTheme := components.DefaultContextMenuTheme()
	for _, cm := range win.ui.ContextMenus() {
		cm.Draw(dst, face, contextMenuTheme)
	}

	if win.debugMode {
		if win.hasHoveredRect {
			rendering.DrawStrokeRect(dst, win.hoveredRect.X, win.hoveredRect.Y, win.hoveredRect.W, win.hoveredRect.H, 2.0, colors.Yellow)
		}
		const label = "Debug Mode"
		lw, lh := text.Measure(label, &face, 0)
		const margin = 8.0
		x := logicalW - lw - margin
		y := logicalH - lh - margin
		if x < margin {
			x = margin
		}
		if y < margin {
			y = margin
		}
		rendering.DrawText(dst, label, face, int(x), int(y), colors.Yellow)
	}
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

// LayoutF accepts the outside size in device-independent pixels and returns the
// logical screen size in pixels.
func (win *Window) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	if outsideWidth <= 0 {
		outsideWidth = float64(win.width)
	}
	if outsideHeight <= 0 {
		outsideHeight = float64(win.height)
	}
	dpi := win.currentDPIScale()
	return outsideWidth * dpi, outsideHeight * dpi
}

func (win *Window) effectiveUIScale(rootScale float64) float64 {
	return normalizeScale(rootScale) * win.WindowScale()
}

func (win *Window) currentDPIScale() float64 {
	if !win.autoDPI {
		return 1
	}
	m := ebiten.Monitor()
	if m == nil {
		return 1
	}
	dpi := m.DeviceScaleFactor()
	if dpi <= 0 {
		return 1
	}
	return dpi
}

func (win *Window) logicalScreenSize() (int, int) {
	w, h := windowSize(win.width, win.height)
	dpi := win.currentDPIScale()
	return int(math.Ceil(float64(w) * dpi)), int(math.Ceil(float64(h) * dpi))
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
