package goak

import (
	"errors"
	"fmt"
	"math"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

var ErrInvalidWindowScale = errors.New("goak: window scale must be finite and greater than zero")

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
	scene       Scene
	sceneCtx    *SceneContext
	beforeFrame func()

	handle             *sdl.Window
	renderer           *sdl.Renderer
	rendererName       string
	font               *rendering.Font
	fontAttempted      bool
	fontAttemptScale   float64
	fontAttemptSize    float32
	fontAttemptRev     uint64
	fontErr            error
	openFont           func(*sdl.Renderer, []byte, float32, float32) (*rendering.Font, error)
	mouseX             float64
	mouseY             float64
	focusedInput       *components.TextInput
	focusedArea        *components.TextArea
	uiMouseCapture     bool
	sceneMouseCapture  bool
	mouseCaptureActive bool
	captureMouse       func(bool) error
	uiPointerPress     uint64
	sceneTextInput     bool
	textInputActive    bool
	running            bool
	destroyed          bool

	unloadSDL func()
	unloadTTF func()
}

// Config contains window creation options.
type Config struct {
	Title           string
	Width           int
	Height          int
	AutoDPI         bool
	WindowScale     float64
	Renderer        RendererDriver
	SkipDefaultFont bool // custom scenes may provide their own font system
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
	win, err := createWindow(cfg)
	if err != nil {
		panic(err)
	}
	return win
}

func createWindow(cfg Config) (*Window, error) {
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
		return nil, fmt.Errorf("goak: SDL init failed: %w", err)
	}
	if err := ttf.Init(); err != nil {
		sdl.Quit()
		cleanupLibraries()
		return nil, fmt.Errorf("goak: SDL_ttf init failed: %w", err)
	}

	flags := sdl.WINDOW_RESIZABLE | sdl.WINDOW_HIGH_PIXEL_DENSITY
	handle, err := sdl.CreateWindow(cfg.Title, cfg.Width, cfg.Height, flags)
	if err != nil {
		ttf.Quit()
		sdl.Quit()
		cleanupLibraries()
		return nil, fmt.Errorf("goak: could not create SDL window: %w", err)
	}

	rendererDriver := normalizeRendererDriver(cfg.Renderer)
	rendererNameHint := string(rendererDriver)
	if rendererDriver == RendererAuto {
		rendererNameHint = ""
	}
	renderer, err := handle.CreateRenderer(rendererNameHint)
	if err != nil {
		handle.Destroy()
		ttf.Quit()
		sdl.Quit()
		cleanupLibraries()
		return nil, fmt.Errorf("goak: could not create SDL renderer %q: %w", rendererDriver, err)
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
	var font *rendering.Font
	if !cfg.SkipDefaultFont {
		font, err = rendering.NewFont(renderer, 20, float32(fontScale))
		if err != nil {
			renderer.Destroy()
			handle.Destroy()
			ttf.Quit()
			sdl.Quit()
			cleanupLibraries()
			return nil, fmt.Errorf("goak: could not open default font: %w", err)
		}
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
		unloadSDL:    unloadSDL,
		unloadTTF:    unloadTTF,
	}, nil
}

// RendererName returns the SDL renderer backing this window.
func (win *Window) RendererName() string {
	if win == nil {
		return ""
	}
	return win.rendererName
}

// UIFontError reports the most recent retained-font configuration failure.
// An unchanged failed configuration is attempted only once.
func (win *Window) UIFontError() error {
	if win == nil {
		return nil
	}
	return win.fontErr
}

// resetRendererState establishes the host contract between custom Scene
// drawing and retained drawing. A Scene may freely change renderer state
// during Draw; Goak restores the window target and coordinate system before
// drawing the next layer.
func (win *Window) resetRendererState(scale float32) error {
	if win == nil || win.renderer == nil {
		return ErrWindowNotInitialized
	}
	return resetRendererStateOn(win.renderer, scale)
}

type rendererStateSetter interface {
	SetRenderTarget(*sdl.Texture) error
	SetLogicalPresentation(int32, int32, sdl.RendererLogicalPresentation) error
	SetViewport(*sdl.Rect) error
	SetClipRect(*sdl.Rect) error
	SetScale(float32, float32) error
	SetColorScale(float32) error
	SetDrawBlendMode(sdl.BlendMode) error
}

func resetRendererStateOn(renderer rendererStateSetter, scale float32) error {
	return errors.Join(
		renderer.SetRenderTarget(nil),
		renderer.SetLogicalPresentation(0, 0, sdl.LOGICAL_PRESENTATION_DISABLED),
		renderer.SetViewport(nil),
		renderer.SetClipRect(nil),
		renderer.SetScale(scale, scale),
		renderer.SetColorScale(1),
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND),
	)
}

func (win *Window) attachUI(ui *components.UI) {
	win.uiPointerPress = 0
	win.setUI(ui)
	win.scene = nil
	win.sceneCtx = nil
}

func (win *Window) attachScene(scene Scene, ctx *SceneContext) {
	win.uiPointerPress = 0
	win.setUI(nil)
	win.sceneTextInput = false
	_ = win.syncTextInput()
	win.scene = scene
	win.sceneCtx = ctx
}

func (win *Window) detachScene() {
	win.uiPointerPress = 0
	win.sceneTextInput = false
	win.sceneMouseCapture = false
	_ = win.syncTextInput()
	_ = win.syncMouseCapture()
	win.scene = nil
	win.sceneCtx = nil
}

func (win *Window) setUI(ui *components.UI) {
	if win.ui == ui {
		return
	}
	win.releaseUI()
	win.ui = ui
	win.fontAttempted = false
	win.fontErr = nil
}

func (win *Window) releaseUI() {
	win.clearTextFocus()
	win.releaseUIMouseCapture()
	if win.ui != nil {
		win.closePopups()
		for index := 0; index < win.ui.ComponentCount(); index++ {
			switch component := win.ui.Component(index).(type) {
			case *components.Slider:
				component.StopDrag()
			case *components.Image:
				component.Close()
			}
		}
	}
	win.ui = nil
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
// Non-positive and non-finite values return ErrInvalidWindowScale.
func (win *Window) SetWindowScale(scale float64) error {
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		return ErrInvalidWindowScale
	}
	next := scale
	if next == win.windowScale {
		return nil
	}
	win.windowScale = next
	if win.onWindowScaleChanged != nil {
		win.onWindowScaleChanged(next)
	}
	if win.ui != nil {
		win.ui.InvalidateLayout()
	}
	return nil
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
func (win *Window) Run() error {
	if win == nil || win.destroyed || win.handle == nil || win.renderer == nil {
		return ErrWindowNotInitialized
	}
	win.running = true
	// Establish retained bounds before the first event can target the UI.
	win.updateUI()
	if win.ui != nil && win.fontErr != nil {
		return win.fontErr
	}
	var event sdl.Event
	for win.running {
		for sdl.PollEvent(&event) {
			win.handleEvent(&event)
		}
		if win.beforeFrame != nil {
			win.beforeFrame()
		}
		if win.scene != nil {
			if updater, ok := win.scene.(SceneUpdater); ok {
				updater.Update()
			}
			// Scene updates may mutate or replace the retained tree.
			win.updateUI()
			_ = win.resetRendererState(1)
			win.scene.Draw(win.sceneCtx)
			win.drawUI(false)
			_ = win.renderer.Present()
		} else {
			win.updateUI()
			win.drawUI(true)
			_ = win.renderer.Present()
		}
	}
	return nil
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
	win.sceneMouseCapture = false
	win.uiMouseCapture = false
	_ = win.syncMouseCapture()
	if win.font != nil {
		win.font.Close()
		win.font = nil
	}
	win.releaseUI()
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
	if win.scene == nil {
		if event.Type == sdl.EVENT_QUIT {
			win.running = false
			return
		}
		if win.ui != nil && (!win.ui.Visible() || !win.ui.Interactive()) {
			win.reconcileUIInputState()
		}
		if win.ui != nil && win.ui.Visible() && win.ui.Interactive() {
			win.handleUIEvent(event, true)
		} else {
			win.handleInactiveUIPointerEvent(event)
		}
		return
	}

	policy := UIInputOverlay
	if win.sceneCtx != nil {
		policy = win.sceneCtx.UIInputPolicy()
	}
	uiAcceptsInput, uiBlocksScene := win.uiInputRouting(policy)
	if win.ui != nil && !uiAcceptsInput {
		win.reconcileUIInputState()
	}
	uiHandled := false
	if uiAcceptsInput {
		uiHandled = win.handleUIEvent(event, false)
	} else {
		uiHandled = win.handleInactiveUIPointerEvent(event)
	}

	translated, ok := win.sceneEvent(event)
	if !ok {
		return
	}
	win.dispatchSceneEvent(translated, uiHandled, uiBlocksScene)
}

func (win *Window) uiInputRouting(policy UIInputPolicy) (acceptsInput, blocksScene bool) {
	visible := win.ui != nil && win.ui.Visible()
	return visible && win.ui.Interactive() && policy != UIInputPassthrough,
		visible && policy == UIInputModal
}

func (win *Window) dispatchSceneEvent(
	event Event,
	uiHandled bool,
	uiBlocksScene bool,
) {
	if event.Type != EventQuit && (uiHandled || uiBlocksScene) {
		return
	}
	handled := false
	if handler, exists := win.scene.(SceneEventHandler); exists {
		handled = handler.HandleEvent(event)
	}
	if event.Type == EventQuit && !handled {
		win.running = false
	}
}

func (win *Window) handleUIEvent(event *sdl.Event, allowWindowShortcuts bool) bool {
	if win.ui == nil || !win.ui.Visible() || !win.ui.Interactive() {
		return false
	}
	switch event.Type {
	case sdl.EVENT_KEY_DOWN:
		key := event.KeyboardEvent()
		if key == nil {
			return false
		}
		binding := keyChord(key.Key, key.Mod)
		if win.handlePopupKey(binding) {
			return true
		}
		if win.focusedInput != nil && win.focusedInput.HandleKey(key.Key, key.Mod) {
			return true
		}
		if win.focusedArea != nil && win.focusedArea.HandleKey(key.Key, key.Mod) {
			return true
		}
		if key.Repeat || !allowWindowShortcuts {
			return false
		}
		switch key.Key {
		case sdl.K_F12:
			win.debugMode = !win.debugMode
			return true
		case sdl.K_EQUALS, sdl.K_PLUS, sdl.K_KP_PLUS:
			if win.scaleHotkeys && key.Mod&sdl.KMOD_CTRL != 0 {
				win.changeWindowScale(0.1)
				return true
			}
		case sdl.K_MINUS, sdl.K_KP_MINUS:
			if win.scaleHotkeys && key.Mod&sdl.KMOD_CTRL != 0 {
				win.changeWindowScale(-0.1)
				return true
			}
		}
	case sdl.EVENT_TEXT_INPUT:
		text := event.TextInputEvent()
		if text == nil {
			return false
		}
		if win.uiHasOpenPopup() {
			return true
		}
		if win.focusedInput != nil {
			win.focusedInput.HandleTextInput(text.Text)
			return true
		} else if win.focusedArea != nil {
			win.focusedArea.HandleTextInput(text.Text)
			return true
		}
	case sdl.EVENT_MOUSE_MOTION:
		mouse := event.MouseMotionEvent()
		if mouse != nil {
			win.mouseX = float64(mouse.X)
			win.mouseY = float64(mouse.Y)
		}
		return win.uiMouseCapture || win.uiPointerPress != 0 || win.uiHasOpenPopup()
	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		mouse := event.MouseButtonEvent()
		if mouse == nil {
			return false
		}
		win.mouseX = float64(mouse.X)
		win.mouseY = float64(mouse.Y)
		x, y := win.logicalMousePosition()
		handled := win.mouseDown(mouseButton(mouse.Button), x, y)
		win.setUIPointerPress(mouse.Button, handled)
		return handled
	case sdl.EVENT_MOUSE_BUTTON_UP:
		mouse := event.MouseButtonEvent()
		if mouse == nil {
			return false
		}
		win.mouseX = float64(mouse.X)
		win.mouseY = float64(mouse.Y)
		return win.finishUIPointerRelease(mouse.Button)
	case sdl.EVENT_MOUSE_WHEEL:
		wheel := event.MouseWheelEvent()
		if wheel == nil || win.ui == nil {
			return false
		}
		win.mouseX = float64(wheel.MouseX)
		win.mouseY = float64(wheel.MouseY)
		if win.handlePopupWheel(float64(wheel.Y)) {
			return true
		}
		x, y := win.logicalMousePosition()
		for index := win.ui.ComponentCount() - 1; index >= 0; index-- {
			area, ok := win.ui.Component(index).(*components.TextArea)
			if !ok || !win.ui.ComponentEnabled(area) ||
				!rendering.PointWithinBounds(x, y, area.Bounds()) {
				continue
			}
			wheelX := float64(wheel.X)
			wheelY := float64(wheel.Y)
			if sdl.GetModState()&sdl.KMOD_SHIFT != 0 && wheelX == 0 {
				wheelX, wheelY = wheelY, 0
			}
			area.ScrollWheel(wheelX, wheelY, win.font)
			return true
		}
		return false
	}
	return false
}

func (win *Window) handleInactiveUIPointerEvent(event *sdl.Event) bool {
	switch event.Type {
	case sdl.EVENT_MOUSE_MOTION:
		return win.uiPointerPress != 0
	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		if mouse := event.MouseButtonEvent(); mouse != nil {
			win.setUIPointerPress(mouse.Button, false)
		}
	case sdl.EVENT_MOUSE_BUTTON_UP:
		if mouse := event.MouseButtonEvent(); mouse != nil {
			return win.finishUIPointerRelease(mouse.Button)
		}
	}
	return false
}

func pointerButtonBit(button uint8) uint64 {
	if button >= 64 {
		return 0
	}
	return uint64(1) << button
}

func (win *Window) setUIPointerPress(button uint8, handled bool) {
	bit := pointerButtonBit(button)
	if handled {
		win.uiPointerPress |= bit
		return
	}
	win.uiPointerPress &^= bit
}

func (win *Window) finishUIPointerRelease(button uint8) bool {
	bit := pointerButtonBit(button)
	handled := win.uiPointerPress&bit != 0
	win.uiPointerPress &^= bit
	if button != uint8(sdl.BUTTON_LEFT) || !win.uiMouseCapture {
		return handled
	}
	win.releaseUIMouseCapture()
	return true
}

func (win *Window) releaseUIMouseCapture() {
	if !win.uiMouseCapture {
		return
	}
	win.uiMouseCapture = false
	_ = win.syncMouseCapture()
	ui := win.ui
	if ui != nil {
		for index := 0; index < ui.ComponentCount(); index++ {
			if slider, ok := ui.Component(index).(*components.Slider); ok {
				slider.StopDrag()
			}
		}
	}
}

func (win *Window) setSceneMouseCapture(capture bool) error {
	if win == nil || win.handle == nil {
		return ErrWindowNotInitialized
	}
	win.sceneMouseCapture = capture
	return win.syncMouseCapture()
}

func (win *Window) setUIMouseCapture(capture bool) error {
	if win == nil || win.handle == nil {
		return ErrWindowNotInitialized
	}
	win.uiMouseCapture = capture
	return win.syncMouseCapture()
}

func (win *Window) syncMouseCapture() error {
	if win == nil {
		return ErrWindowNotInitialized
	}
	desired := win.sceneMouseCapture || win.uiMouseCapture
	if desired == win.mouseCaptureActive {
		return nil
	}
	capture := win.captureMouse
	if capture == nil {
		capture = sdl.CaptureMouse
	}
	if err := capture(desired); err != nil {
		return err
	}
	win.mouseCaptureActive = desired
	return nil
}

func (win *Window) sceneEvent(event *sdl.Event) (Event, bool) {
	mods := modifiers(sdl.GetModState())
	density := float32(win.pixelDensity())
	switch event.Type {
	case sdl.EVENT_QUIT:
		return Event{Type: EventQuit, Modifiers: mods}, true
	case sdl.EVENT_KEY_DOWN:
		key := event.KeyboardEvent()
		if key == nil {
			return Event{}, false
		}
		return Event{
			Type: EventKeyDown, Key: keyChord(key.Key, key.Mod), Repeat: key.Repeat,
			Modifiers: modifiers(key.Mod),
		}, true
	case sdl.EVENT_TEXT_INPUT:
		text := event.TextInputEvent()
		if text == nil {
			return Event{}, false
		}
		return Event{Type: EventTextInput, Text: text.Text, Modifiers: mods}, true
	case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
		mouse := event.MouseButtonEvent()
		if mouse == nil {
			return Event{}, false
		}
		typeOfEvent := EventMouseDown
		if event.Type == sdl.EVENT_MOUSE_BUTTON_UP {
			typeOfEvent = EventMouseUp
		}
		x, y := rendererPoint(mouse.X, mouse.Y, density)
		return Event{
			Type: typeOfEvent, X: x, Y: y,
			Button: mouseButton(mouse.Button), Clicks: mouse.Clicks, Modifiers: mods,
		}, true
	case sdl.EVENT_MOUSE_MOTION:
		mouse := event.MouseMotionEvent()
		if mouse == nil {
			return Event{}, false
		}
		x, y := rendererPoint(mouse.X, mouse.Y, density)
		return Event{
			Type: EventMouseMove, X: x, Y: y,
			Modifiers: mods,
		}, true
	case sdl.EVENT_MOUSE_WHEEL:
		wheel := event.MouseWheelEvent()
		if wheel == nil {
			return Event{}, false
		}
		x, y := rendererPoint(wheel.MouseX, wheel.MouseY, density)
		return Event{
			Type: EventMouseWheel, X: x, Y: y,
			WheelX: wheel.X, WheelY: wheel.Y, Modifiers: mods,
		}, true
	default:
		return Event{}, false
	}
}

func rendererPoint(x, y, density float32) (float32, float32) {
	return x * density, y * density
}

func mouseButton(button uint8) MouseButton {
	switch button {
	case uint8(sdl.BUTTON_LEFT):
		return MouseLeft
	case uint8(sdl.BUTTON_MIDDLE):
		return MouseMiddle
	case uint8(sdl.BUTTON_RIGHT):
		return MouseRight
	default:
		return MouseButton(button)
	}
}

func (win *Window) updateUI() {
	if win.ui == nil {
		return
	}
	outputW, outputH := win.outputSize()
	scale := win.renderScale()
	_ = win.ensureFont(scale, win.ui)
	for index := 0; index < win.ui.ComponentCount(); index++ {
		if menu, ok := win.ui.Component(index).(*components.MenuBar); ok {
			menu.SyncWidth(win.font)
		}
	}
	viewport := layout.Rect{W: float64(outputW) / scale, H: float64(outputH) / scale}
	win.ui.Layout(viewport.W, viewport.H)
	for index := 0; index < win.ui.ComponentCount(); index++ {
		switch component := win.ui.Component(index).(type) {
		case *components.MenuBar:
			component.Place(viewport)
		case *components.Dropdown:
			component.Place(viewport)
		}
	}
	if !win.ui.Visible() || !win.ui.Interactive() ||
		win.sceneCtx != nil && win.sceneCtx.UIInputPolicy() == UIInputPassthrough {
		win.reconcileUIInputState()
		return
	}

	ui := win.ui
	if win.focusedInput != nil && (!ui.Contains(win.focusedInput) || !ui.ComponentEnabled(win.focusedInput)) ||
		win.focusedArea != nil && (!ui.Contains(win.focusedArea) || !ui.ComponentEnabled(win.focusedArea)) {
		win.clearTextFocus()
	}
	x, y := win.logicalMousePosition()
	for index := ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := ui.ContextMenu(index)
		if menu.IsOpen() && ui.ContextMenuEnabled(menu) {
			menu.SetHovered(menu.HitTest(x, y))
		}
	}
	for index := 0; index < ui.ComponentCount(); index++ {
		component := ui.Component(index)
		if !ui.ComponentEnabled(component) {
			continue
		}
		switch component := component.(type) {
		case *components.MenuBar:
			component.OnMouseMove(x, y)
		case *components.Slider:
			if component.IsDragging() {
				component.UpdateValue(x)
				if win.ui != ui {
					return
				}
			}
		case *components.RadioGroup:
			component.SetHovered(component.HitTest(x, y))
		case *components.Dropdown:
			if component.IsOpen() {
				component.SetHovered(component.HitTestList(x, y))
			}
		}
	}
	win.updateHoveredElement(x, y)
}

func (win *Window) mouseDown(button MouseButton, x, y float64) bool {
	if win.ui == nil {
		return false
	}
	ui := win.ui
	if win.popupMouseDown(button, x, y) {
		return true
	}
	if button == MouseRight {
		return win.openContextMenu(x, y)
	}
	if button != MouseLeft {
		return false
	}

	for index := ui.ComponentCount() - 1; index >= 0; index-- {
		component := ui.Component(index)
		if !ui.ComponentEnabled(component) ||
			!rendering.PointWithinBounds(x, y, component.Container().Bounds) {
			continue
		}
		switch component := component.(type) {
		case *components.TextArea:
			win.focusTextArea(component)
			component.SetCursorAt(x, y, win.font)
			return true
		case *components.TextInput:
			win.focusTextInput(component)
			component.SetCursorAt(x, win.font)
			return true
		case *components.MenuBar:
			win.clearTextFocus()
			if component.OnMouseDown(x, y) {
				if win.ui == ui {
					win.closePopupsExceptMenuBar(component)
				}
				return true
			}
		case *components.Dropdown:
			win.clearTextFocus()
			win.closePopups()
			if win.ui == ui {
				component.Open()
			}
			return true
		case *components.Slider:
			win.clearTextFocus()
			component.StartDrag()
			if err := win.setUIMouseCapture(true); err != nil {
				component.StopDrag()
				return true
			}
			component.UpdateValue(x)
			return true
		case *components.RadioGroup:
			win.clearTextFocus()
			if option := component.HitTest(x, y); option >= 0 {
				component.SetSelectedIndex(option)
				return true
			}
		case *components.Checkbox:
			win.clearTextFocus()
			component.Toggle()
			return true
		case *components.Button:
			win.clearTextFocus()
			component.Click()
			return true
		}
	}
	win.clearTextFocus()
	return false
}

// popupMouseDown gives the active overlay stack first refusal and consumes
// every pointer button while a popup is modal.
func (win *Window) popupMouseDown(button MouseButton, x, y float64) bool {
	if win.ui == nil || !win.uiHasOpenPopup() {
		return false
	}
	if button != MouseLeft {
		win.closePopups()
		return true
	}
	ui := win.ui
	for index := ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := ui.ContextMenu(index)
		if !menu.IsOpen() || !ui.ContextMenuEnabled(menu) {
			continue
		}
		if item := menu.HitTest(x, y); item >= 0 {
			menu.Click(item)
		} else {
			menu.Close()
		}
		return true
	}
	for index := ui.ComponentCount() - 1; index >= 0; index-- {
		component := ui.Component(index)
		if !ui.ComponentEnabled(component) {
			continue
		}
		switch component := component.(type) {
		case *components.MenuBar:
			if component.IsOpen() {
				_ = component.OnMouseDown(x, y)
				if win.ui == ui {
					win.closePopupsExceptMenuBar(component)
				}
				return true
			}
		case *components.Dropdown:
			if !component.IsOpen() {
				continue
			}
			if item := component.HitTestList(x, y); item >= 0 {
				component.SetSelectedIndex(item)
			} else {
				component.Close()
			}
			return true
		}
	}
	win.closePopups()
	return true
}

func (win *Window) handlePopupKey(binding string) bool {
	if win.ui == nil {
		return false
	}
	ui := win.ui
	for index := ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := ui.ContextMenu(index)
		if menu.IsOpen() && ui.ContextMenuEnabled(menu) {
			_ = menu.HandleKey(binding)
			return true
		}
	}
	for index := ui.ComponentCount() - 1; index >= 0; index-- {
		component := ui.Component(index)
		if !ui.ComponentEnabled(component) {
			continue
		}
		switch component := component.(type) {
		case *components.MenuBar:
			if component.IsOpen() {
				_ = component.HandleKey(binding)
				return true
			}
		case *components.Dropdown:
			if component.IsOpen() {
				_ = component.HandleKey(binding)
				return true
			}
		}
	}
	return false
}

func (win *Window) handlePopupWheel(y float64) bool {
	if win.ui == nil {
		return false
	}
	ui := win.ui
	for index := ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := ui.ContextMenu(index)
		if menu.IsOpen() && ui.ContextMenuEnabled(menu) {
			menu.ScrollWheel(y)
			return true
		}
	}
	for index := ui.ComponentCount() - 1; index >= 0; index-- {
		component := ui.Component(index)
		if !ui.ComponentEnabled(component) {
			continue
		}
		switch component := component.(type) {
		case *components.MenuBar:
			if component.IsOpen() {
				component.ScrollWheel(y)
				return true
			}
		case *components.Dropdown:
			if component.IsOpen() {
				component.ScrollWheel(y)
				return true
			}
		}
	}
	return false
}

func (win *Window) openContextMenu(x, y float64) bool {
	// Custom scenes own application context. Their right-click first reaches
	// the Scene, which can populate and open a retained menu through
	// SceneContext.OpenContextMenu. Retained-only apps keep automatic opening.
	if win.ui == nil || win.scene != nil {
		return false
	}
	win.closePopups()
	for index := win.ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := win.ui.ContextMenu(index)
		if !menu.AutoOpen() || !win.ui.ContextMenuEnabled(menu) {
			continue
		}
		outputW, outputH := win.outputSize()
		scale := win.renderScale()
		menu.OpenAt(x, y, layout.Rect{
			W: float64(outputW) / scale,
			H: float64(outputH) / scale,
		})
		return true
	}
	return false
}

func (win *Window) closePopups() {
	if win.ui == nil {
		return
	}
	for index := 0; index < win.ui.ComponentCount(); index++ {
		switch component := win.ui.Component(index).(type) {
		case *components.MenuBar:
			component.Close()
		case *components.Dropdown:
			component.Close()
		}
	}
	for index := 0; index < win.ui.ContextMenuCount(); index++ {
		win.ui.ContextMenu(index).Close()
	}
}

func (win *Window) closePopupsExceptMenuBar(keep *components.MenuBar) {
	if win.ui == nil {
		return
	}
	for index := 0; index < win.ui.ComponentCount(); index++ {
		switch component := win.ui.Component(index).(type) {
		case *components.MenuBar:
			if component != keep {
				component.Close()
			}
		case *components.Dropdown:
			component.Close()
		}
	}
	for index := 0; index < win.ui.ContextMenuCount(); index++ {
		win.ui.ContextMenu(index).Close()
	}
}

func (win *Window) reconcileUIInputState() {
	win.clearTextFocus()
	win.closePopups()
	win.releaseUIMouseCapture()
}

func (win *Window) uiHasOpenPopup() bool {
	if win.ui == nil {
		return false
	}
	for index := 0; index < win.ui.ComponentCount(); index++ {
		component := win.ui.Component(index)
		if !win.ui.ComponentVisible(component) {
			continue
		}
		switch component := component.(type) {
		case *components.MenuBar:
			if component.IsOpen() {
				return true
			}
		case *components.Dropdown:
			if component.IsOpen() {
				return true
			}
		}
	}
	for index := 0; index < win.ui.ContextMenuCount(); index++ {
		menu := win.ui.ContextMenu(index)
		if menu.IsOpen() && win.ui.ContextMenuVisible(menu) {
			return true
		}
	}
	return false
}

func (win *Window) drawUI(clearBackground bool) {
	if win.ui == nil || win.renderer == nil {
		return
	}
	scale := win.renderScale()
	_ = win.resetRendererState(float32(scale))
	theme := win.ui.Theme()
	if clearBackground {
		bg := theme.Background
		_ = win.renderer.SetDrawColor(bg.R, bg.G, bg.B, bg.A)
		_ = win.renderer.Clear()
	}
	if !win.ui.Visible() {
		return
	}
	_ = win.ensureFont(scale, win.ui)
	if win.font == nil {
		return
	}

	x, y := win.logicalMousePosition()
	for index := 0; index < win.ui.ComponentCount(); index++ {
		component := win.ui.Component(index)
		if !win.ui.ComponentVisible(component) {
			continue
		}
		switch component := component.(type) {
		case *components.Panel:
			component.Draw(win.renderer, theme.Panel)
		case *components.Image:
			component.Draw(win.renderer)
		case *components.Label:
			component.Draw(win.renderer, win.font, theme.Label)
		case *components.Button:
			component.Draw(win.renderer, win.font, theme.Button)
		case *components.Checkbox:
			component.Draw(win.renderer, win.font, theme.Checkbox,
				win.ui.ComponentEnabled(component) && rendering.PointWithinBounds(x, y, component.Bounds()))
		case *components.RadioGroup:
			component.Draw(win.renderer, win.font, theme.RadioGroup)
		case *components.Slider:
			component.Draw(win.renderer, win.font, theme.Slider)
		case *components.Dropdown:
			component.DrawControl(win.renderer, win.font, theme.Dropdown)
		case *components.TextInput:
			component.Draw(win.renderer, win.font, theme.TextInput)
		case *components.TextArea:
			component.Draw(win.renderer, win.font, theme.TextArea)
		case *components.MenuBar:
			component.DrawBar(win.renderer, win.font, theme.MenuBar)
		}
	}
	// Popups are a separate overlay stack above the retained tree.
	for index := 0; index < win.ui.ComponentCount(); index++ {
		component := win.ui.Component(index)
		if !win.ui.ComponentVisible(component) {
			continue
		}
		switch component := component.(type) {
		case *components.Dropdown:
			component.DrawList(win.renderer, win.font, theme.Dropdown)
		case *components.MenuBar:
			component.DrawDropdown(win.renderer, win.font, theme.MenuBar)
		}
	}
	for index := 0; index < win.ui.ContextMenuCount(); index++ {
		menu := win.ui.ContextMenu(index)
		if win.ui.ContextMenuVisible(menu) {
			menu.Draw(win.renderer, win.font, theme.ContextMenu)
		}
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
}

func (win *Window) ensureFont(scale float64, ui *components.UI) error {
	if ui == nil {
		return nil
	}
	size := float32(ui.FontSize())
	revision := ui.FontRevision()
	if win.renderer == nil {
		return ErrWindowNotInitialized
	}
	if win.fontAttempted && math.Abs(win.fontAttemptScale-scale) < 0.01 &&
		math.Abs(float64(win.fontAttemptSize-size)) < 0.01 &&
		win.fontAttemptRev == revision {
		return win.fontErr
	}
	win.fontAttempted = true
	win.fontAttemptScale = scale
	win.fontAttemptSize = size
	win.fontAttemptRev = revision
	openFont := win.openFont
	if openFont == nil {
		openFont = rendering.NewFontFromBytes
	}
	font, err := openFont(win.renderer, ui.FontData(), size, float32(scale))
	if err != nil {
		win.fontErr = fmt.Errorf("goak: could not open retained UI font: %w", err)
		return win.fontErr
	}
	if win.font != nil {
		win.font.Close()
	}
	win.font = font
	win.fontErr = nil
	return nil
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
		rootScale = win.ui.Root().Scale()
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
	_ = win.SetWindowScale(math.Max(minScale, math.Min(maxScale, win.WindowScale()+delta)))
}

func normalizeScale(value float64) float64 {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 1
	}
	return value
}

func (win *Window) focusTextInput(input *components.TextInput) {
	if win.focusedInput == input {
		return
	}
	win.clearTextFocus()
	win.focusedInput = input
	input.SetFocused(true)
	_ = win.syncTextInput()
}

func (win *Window) focusTextArea(area *components.TextArea) {
	if win.focusedArea == area {
		return
	}
	win.clearTextFocus()
	win.focusedArea = area
	area.SetFocused(true)
	_ = win.syncTextInput()
}

func (win *Window) clearTextFocus() {
	if win.focusedInput != nil {
		win.focusedInput.SetFocused(false)
		win.focusedInput = nil
	}
	if win.focusedArea != nil {
		win.focusedArea.SetFocused(false)
		win.focusedArea = nil
	}
	_ = win.syncTextInput()
}

func (win *Window) syncTextInput() error {
	if win.handle == nil {
		return ErrWindowNotInitialized
	}
	enabled := win.sceneTextInput || win.focusedInput != nil || win.focusedArea != nil
	if enabled == win.textInputActive {
		return nil
	}
	var err error
	if enabled {
		err = win.handle.StartTextInput()
	} else {
		err = win.handle.StopTextInput()
	}
	if err == nil {
		win.textInputActive = enabled
	}
	return err
}

func (win *Window) updateHoveredElement(x, y float64) {
	win.hasHoveredRect = false
	if !win.debugMode || win.ui == nil {
		return
	}

	for index := win.ui.ContextMenuCount() - 1; index >= 0; index-- {
		menu := win.ui.ContextMenu(index)
		if menu.IsOpen() && rendering.PointWithinBounds(x, y, menu.Bounds()) {
			win.hoveredRect = menu.Bounds()
			win.hasHoveredRect = true
			return
		}
	}
	for index := win.ui.ComponentCount() - 1; index >= 0; index-- {
		component := win.ui.Component(index)
		if !win.ui.ComponentVisible(component) {
			continue
		}
		if menu, ok := component.(*components.MenuBar); ok && menu.IsOpen() {
			for _, rect := range menu.OpenSubItemRects() {
				if rendering.PointWithinBounds(x, y, rect) {
					win.hoveredRect = rect
					win.hasHoveredRect = true
					return
				}
			}
		}
		bounds := component.Container().Bounds
		if rendering.PointWithinBounds(x, y, bounds) {
			win.hoveredRect = bounds
			win.hasHoveredRect = true
			return
		}
	}
}
