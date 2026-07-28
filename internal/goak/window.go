package goak

import (
	"fmt"
	"math"

	"goak/internal/goak/components"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
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

	ui          *components.UI
	beforeFrame func()

	handle       *sdl.Window
	renderer     *sdl.Renderer
	rendererName string
	font         *rendering.Font
	fontScale    float64
	mouseX       float64
	mouseY       float64
	running      bool
	destroyed    bool

	unloadSDL func()
	unloadTTF func()
}

// Config contains window creation options.
type Config struct {
	Title       string
	Width       int
	Height      int
	AutoDPI     bool
	WindowScale float64
	Renderer    RendererDriver
}

// InitWindow creates a resizable SDL3 window.
func InitWindow(title string, width, height int) *Window {
	return newWindow(Config{
		Title:       title,
		Width:       width,
		Height:      height,
		WindowScale: 1,
		Renderer:    RendererSoftware,
	})
}

func newWindow(cfg Config) *Window {
	if cfg.Width <= 0 {
		cfg.Width = 800
	}
	if cfg.Height <= 0 {
		cfg.Height = 600
	}
	if cfg.Title == "" {
		cfg.Title = "Goak"
	}

	sdlLibrary := binsdl.Load()
	ttfLibrary := binttf.Load()
	unloadSDL := sdlLibrary.Unload
	unloadTTF := ttfLibrary.Unload

	cleanupLibraries := func() {
		unloadTTF()
		unloadSDL()
	}
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		cleanupLibraries()
		panic(fmt.Errorf("goak: SDL init failed: %w", err))
	}
	if err := ttf.Init(); err != nil {
		sdl.Quit()
		cleanupLibraries()
		panic(fmt.Errorf("goak: SDL_ttf init failed: %w", err))
	}

	flags := sdl.WINDOW_RESIZABLE | sdl.WINDOW_HIGH_PIXEL_DENSITY
	handle, err := sdl.CreateWindow(cfg.Title, cfg.Width, cfg.Height, flags)
	if err != nil {
		ttf.Quit()
		sdl.Quit()
		cleanupLibraries()
		panic(fmt.Errorf("goak: could not create SDL window: %w", err))
	}

	rendererDriver := normalizeRendererDriver(cfg.Renderer)
	renderer, err := handle.CreateRenderer(string(rendererDriver))
	if err != nil {
		handle.Destroy()
		ttf.Quit()
		sdl.Quit()
		cleanupLibraries()
		panic(fmt.Errorf("goak: could not create SDL renderer %q: %w", rendererDriver, err))
	}
	_ = renderer.SetVSync(1)
	_ = renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	rendererName, nameErr := renderer.Name()
	if nameErr != nil {
		rendererName = string(rendererDriver)
	}

	fontScale := normalizeScale(cfg.WindowScale)
	if cfg.AutoDPI {
		if density, densityErr := handle.PixelDensity(); densityErr == nil && density > 0 {
			fontScale *= float64(density)
		}
	}
	font, err := rendering.NewFont(renderer, 20, float32(fontScale))
	if err != nil {
		renderer.Destroy()
		handle.Destroy()
		ttf.Quit()
		sdl.Quit()
		cleanupLibraries()
		panic(fmt.Errorf("goak: could not open default font: %w", err))
	}

	return &Window{
		title:        cfg.Title,
		width:        cfg.Width,
		height:       cfg.Height,
		autoDPI:      cfg.AutoDPI,
		windowScale:  normalizeScale(cfg.WindowScale),
		handle:       handle,
		renderer:     renderer,
		rendererName: rendererName,
		font:         font,
		fontScale:    fontScale,
		unloadSDL:    unloadSDL,
		unloadTTF:    unloadTTF,
	}
}

// RendererName returns the SDL renderer backing this window.
func (win *Window) RendererName() string {
	if win == nil {
		return ""
	}
	return win.rendererName
}

func (win *Window) attachUI(ui *components.UI) {
	win.ui = ui
}

// SetTitle updates the native window title.
func (win *Window) SetTitle(title string) {
	win.title = title
	if win.handle != nil {
		_ = win.handle.SetTitle(title)
	}
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
	if scale <= 0 {
		return
	}
	next := scale
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
func (win *Window) PointWithinBounds(x, y float64, r layout.Rect) bool {
	return rendering.PointWithinBounds(x, y, r)
}

// Run processes SDL events and renders frames until the window is closed.
func (win *Window) Run() {
	if win == nil || win.destroyed || win.handle == nil || win.renderer == nil {
		return
	}
	win.running = true
	var event sdl.Event
	for win.running {
		for sdl.PollEvent(&event) {
			win.handleEvent(&event)
		}
		if win.beforeFrame != nil {
			win.beforeFrame()
		}
		win.updateUI()
		win.drawUI()
	}
}

func (win *Window) setBeforeFrame(beforeFrame func()) {
	win.beforeFrame = beforeFrame
}

// Destroy releases all native resources. It is safe to call more than once.
func (win *Window) Destroy() {
	if win == nil || win.destroyed {
		return
	}
	win.destroyed = true
	win.running = false
	if win.font != nil {
		win.font.Close()
		win.font = nil
	}
	if win.renderer != nil {
		win.renderer.Destroy()
		win.renderer = nil
	}
	if win.handle != nil {
		win.handle.Destroy()
		win.handle = nil
	}
	ttf.Quit()
	sdl.Quit()
	if win.unloadTTF != nil {
		win.unloadTTF()
		win.unloadTTF = nil
	}
	if win.unloadSDL != nil {
		win.unloadSDL()
		win.unloadSDL = nil
	}
}

func (win *Window) handleEvent(event *sdl.Event) {
	switch event.Type {
	case sdl.EVENT_QUIT:
		win.running = false
	case sdl.EVENT_KEY_DOWN:
		key := event.KeyboardEvent()
		if key == nil || key.Repeat {
			return
		}
		switch key.Key {
		case sdl.K_F12:
			win.debugMode = !win.debugMode
		case sdl.K_EQUALS, sdl.K_PLUS, sdl.K_KP_PLUS:
			if win.scaleHotkeys && key.Mod&sdl.KMOD_CTRL != 0 {
				win.changeWindowScale(0.1)
			}
		case sdl.K_MINUS, sdl.K_KP_MINUS:
			if win.scaleHotkeys && key.Mod&sdl.KMOD_CTRL != 0 {
				win.changeWindowScale(-0.1)
			}
		}
	case sdl.EVENT_MOUSE_MOTION:
		mouse := event.MouseMotionEvent()
		if mouse != nil {
			win.mouseX = float64(mouse.X)
			win.mouseY = float64(mouse.Y)
		}
	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		mouse := event.MouseButtonEvent()
		if mouse == nil {
			return
		}
		win.mouseX = float64(mouse.X)
		win.mouseY = float64(mouse.Y)
		x, y := win.logicalMousePosition()
		switch mouse.Button {
		case uint8(sdl.BUTTON_LEFT):
			win.mouseDown(x, y)
		case uint8(sdl.BUTTON_RIGHT):
			win.openContextMenu(x, y)
		}
	case sdl.EVENT_MOUSE_BUTTON_UP:
		mouse := event.MouseButtonEvent()
		if mouse == nil {
			return
		}
		win.mouseX = float64(mouse.X)
		win.mouseY = float64(mouse.Y)
		if mouse.Button == uint8(sdl.BUTTON_LEFT) && win.ui != nil {
			_ = sdl.CaptureMouse(false)
			for _, slider := range win.ui.Sliders() {
				slider.StopDrag()
			}
		}
	}
}

func (win *Window) updateUI() {
	if win.ui == nil {
		return
	}
	outputW, outputH := win.outputSize()
	scale := win.renderScale()
	root := win.ui.Root()
	for _, menu := range win.ui.MenuBars() {
		menu.SyncWidth()
	}
	layout.Layout(root.Container(), float64(outputW)/scale, float64(outputH)/scale)

	x, y := win.logicalMousePosition()
	for _, menu := range win.ui.MenuBars() {
		menu.OnMouseMove(x, y)
	}
	for _, menu := range win.ui.ContextMenus() {
		if menu.IsOpen() {
			menu.SetHovered(menu.HitTest(x, y))
		}
	}
	for _, slider := range win.ui.Sliders() {
		if slider.IsDragging() {
			slider.UpdateValue(x)
		}
	}
	for _, group := range win.ui.RadioGroups() {
		group.SetHovered(group.HitTest(x, y))
	}
	for _, dropdown := range win.ui.Dropdowns() {
		if dropdown.IsOpen() {
			dropdown.SetHovered(dropdown.HitTestList(x, y))
		}
	}
	win.updateHoveredElement(x, y)
}

func (win *Window) mouseDown(x, y float64) {
	if win.ui == nil {
		return
	}

	for _, menu := range win.ui.ContextMenus() {
		if !menu.IsOpen() {
			continue
		}
		if index := menu.HitTest(x, y); index >= 0 {
			menu.Click(index)
			return
		}
		menu.Close()
	}

	for _, menu := range win.ui.MenuBars() {
		if menu.OnMouseDown(x, y) {
			return
		}
	}
	for i, button := range win.ui.Buttons() {
		if rendering.PointWithinBounds(x, y, button.Bounds()) {
			win.ui.ButtonClicked(i)
			return
		}
	}
	for _, checkbox := range win.ui.Checkboxes() {
		if rendering.PointWithinBounds(x, y, checkbox.Bounds()) {
			checkbox.Toggle()
			return
		}
	}
	for _, group := range win.ui.RadioGroups() {
		if index := group.HitTest(x, y); index >= 0 {
			group.SetSelectedIndex(index)
			return
		}
	}
	for _, slider := range win.ui.Sliders() {
		if rendering.PointWithinBounds(x, y, slider.Bounds()) {
			slider.StartDrag()
			slider.UpdateValue(x)
			_ = sdl.CaptureMouse(true)
			return
		}
	}
	for _, dropdown := range win.ui.Dropdowns() {
		if dropdown.IsOpen() {
			if index := dropdown.HitTestList(x, y); index >= 0 {
				dropdown.SetSelectedIndex(index)
				return
			}
			if !rendering.PointWithinBounds(x, y, dropdown.ListBounds()) {
				dropdown.Close()
				return
			}
		} else if rendering.PointWithinBounds(x, y, dropdown.Bounds()) {
			dropdown.Open()
			return
		}
	}
}

func (win *Window) openContextMenu(x, y float64) {
	if win.ui == nil {
		return
	}
	menus := win.ui.ContextMenus()
	for _, menu := range menus {
		menu.Close()
	}
	if len(menus) > 0 {
		menus[len(menus)-1].Open(x, y)
	}
}

func (win *Window) drawUI() {
	if win.ui == nil || win.renderer == nil {
		return
	}
	scale := win.renderScale()
	win.ensureFont(scale)
	_ = win.renderer.SetScale(float32(scale), float32(scale))
	theme := win.ui.Theme()
	bg := theme.Background
	_ = win.renderer.SetDrawColor(bg.R, bg.G, bg.B, bg.A)
	_ = win.renderer.Clear()

	for _, panel := range win.ui.Panels() {
		panel.Draw(win.renderer, theme.Panel)
	}
	for _, label := range win.ui.Labels() {
		label.Draw(win.renderer, win.font, theme.Label)
	}
	for _, button := range win.ui.Buttons() {
		button.Draw(win.renderer, win.font, theme.Button)
	}
	for _, checkbox := range win.ui.Checkboxes() {
		checkbox.Draw(win.renderer, win.font, theme.Checkbox, false)
	}
	for _, group := range win.ui.RadioGroups() {
		group.Draw(win.renderer, win.font, theme.RadioGroup)
	}
	for _, slider := range win.ui.Sliders() {
		slider.Draw(win.renderer, win.font, theme.Slider)
	}
	for _, dropdown := range win.ui.Dropdowns() {
		dropdown.Draw(win.renderer, win.font, theme.Dropdown)
	}
	for _, menu := range win.ui.MenuBars() {
		menu.DrawBar(win.renderer, win.font, theme.MenuBar)
	}
	for _, menu := range win.ui.MenuBars() {
		menu.DrawDropdown(win.renderer, win.font, theme.MenuBar)
	}
	for _, menu := range win.ui.ContextMenus() {
		menu.Draw(win.renderer, win.font, theme.ContextMenu)
	}

	if win.debugMode {
		if win.hasHoveredRect {
			rendering.DrawStrokeRect(
				win.renderer,
				win.hoveredRect.X, win.hoveredRect.Y,
				win.hoveredRect.W, win.hoveredRect.H,
				2, theme.Debug.Outline,
			)
		}
		const label = "Debug Mode"
		labelW, labelH := win.font.Measure(label)
		outputW, outputH := win.outputSize()
		logicalW := float64(outputW) / scale
		logicalH := float64(outputH) / scale
		const margin = 8.0
		x := max(margin, logicalW-labelW-margin)
		y := max(margin, logicalH-labelH-margin)
		rendering.DrawText(win.renderer, label, win.font, x, y, theme.Debug.Text)
	}
	_ = win.renderer.Present()
}

func (win *Window) ensureFont(scale float64) {
	if win.renderer == nil || (win.font != nil && math.Abs(win.fontScale-scale) < 0.01) {
		return
	}
	font, err := rendering.NewFont(win.renderer, 20, float32(scale))
	if err != nil {
		return
	}
	if win.font != nil {
		win.font.Close()
	}
	win.font = font
	win.fontScale = scale
}

func (win *Window) outputSize() (int32, int32) {
	if win.renderer != nil {
		if width, height, err := win.renderer.CurrentOutputSize(); err == nil &&
			width > 0 && height > 0 {
			return width, height
		}
	}
	return int32(win.width), int32(win.height)
}

func (win *Window) pixelDensity() float64 {
	if win.handle != nil {
		if density, err := win.handle.PixelDensity(); err == nil && density > 0 {
			return float64(density)
		}
	}
	return 1
}

func (win *Window) renderScale() float64 {
	rootScale := 1.0
	if win.ui != nil {
		rootScale = win.ui.Root().Scale
	}
	scale := win.effectiveUIScale(rootScale)
	if win.autoDPI {
		scale *= win.pixelDensity()
	}
	return max(0.01, scale)
}

func (win *Window) logicalMousePosition() (float64, float64) {
	scale := win.renderScale()
	density := win.pixelDensity()
	return win.mouseX * density / scale, win.mouseY * density / scale
}

func (win *Window) effectiveUIScale(rootScale float64) float64 {
	return normalizeScale(rootScale) * win.WindowScale()
}

func (win *Window) changeWindowScale(delta float64) {
	const (
		minScale = 0.5
		maxScale = 4.0
	)
	win.SetWindowScale(math.Max(minScale, math.Min(maxScale, win.WindowScale()+delta)))
}

func normalizeScale(value float64) float64 {
	if value <= 0 {
		return 1
	}
	return value
}

func (win *Window) updateHoveredElement(x, y float64) {
	win.hasHoveredRect = false
	if !win.debugMode || win.ui == nil {
		return
	}

	menus := win.ui.MenuBars()
	for i := len(menus) - 1; i >= 0; i-- {
		menu := menus[i]
		if menu.IsOpen() {
			rects := menu.OpenSubItemRects()
			for j := len(rects) - 1; j >= 0; j-- {
				if rendering.PointWithinBounds(x, y, rects[j]) {
					win.hoveredRect = rects[j]
					win.hasHoveredRect = true
					return
				}
			}
		}
	}
	for i := len(menus) - 1; i >= 0; i-- {
		menu := menus[i]
		rects := menu.TopItemRects()
		for j := len(rects) - 1; j >= 0; j-- {
			if rendering.PointWithinBounds(x, y, rects[j]) {
				win.hoveredRect = rects[j]
				win.hasHoveredRect = true
				return
			}
		}
		if rendering.PointWithinBounds(x, y, menu.Bounds()) {
			win.hoveredRect = menu.Bounds()
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
