package goak

import (
	"errors"
	"image"
	"runtime"
	"strconv"
	"strings"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"

	"github.com/Zyko0/go-sdl3/sdl"
)

var (
	ErrNilScene             = errors.New("goak: scene is nil")
	ErrInvalidUIInputPolicy = errors.New("goak: invalid retained UI input policy")
)

// Scene is the minimal interface for a custom interface hosted by Goak.
// Drawing runs on the SDL thread and the framework presents the completed
// frame. Applications can implement the optional lifecycle interfaces below.
type Scene interface {
	Draw(*SceneContext)
}

type SceneInitializer interface {
	Init(*SceneContext) error
}

type SceneEventHandler interface {
	HandleEvent(Event) bool
}

type SceneUpdater interface {
	Update()
}

type SceneCloser interface {
	Close()
}

type EventType uint8

const (
	EventQuit EventType = iota + 1
	EventKeyDown
	EventTextInput
	EventMouseDown
	EventMouseUp
	EventMouseMove
	EventMouseWheel
)

type MouseButton uint8

const (
	MouseLeft MouseButton = iota + 1
	MouseMiddle
	MouseRight
)

// Modifiers is a platform-normalized keyboard modifier snapshot. Primary is
// Control on Linux/Windows and Command on macOS.
type Modifiers struct {
	Primary bool
	Control bool
	Alt     bool
	Shift   bool
	Super   bool
}

// Event is the renderer-pixel input delivered to custom scenes. Key contains
// a stable chord such as "ctrl+shift+p", "left", or "f12". Printable text is
// delivered separately through EventTextInput.
type Event struct {
	Type      EventType
	Key       string
	Text      string
	X, Y      float32
	WheelX    float32
	WheelY    float32
	Button    MouseButton
	Clicks    uint8
	Repeat    bool
	Modifiers Modifiers
}

type Cursor uint8

const (
	CursorDefault Cursor = iota
	CursorText
	CursorPointer
	CursorResizeEW
	CursorResizeNS
)

// UIInputPolicy controls how a retained UI attached to a scene participates
// in input routing. The retained UI is always drawn after the scene.
type UIInputPolicy uint8

const (
	// UIInputOverlay gives retained controls the first opportunity to handle an
	// event, then forwards unhandled events to the scene.
	UIInputOverlay UIInputPolicy = iota
	// UIInputModal routes non-quit input exclusively to the retained UI.
	UIInputModal
	// UIInputPassthrough draws the retained UI without routing input to it.
	UIInputPassthrough
)

// SceneContext exposes window services that custom interfaces need without
// making them own SDL initialization, shutdown, event polling, or presenting.
// Renderer is an intentional escape hatch for advanced drawing layers.
type SceneContext struct {
	window        *Window
	cursors       map[Cursor]*sdl.Cursor
	currentCursor Cursor
	uiInputPolicy UIInputPolicy
}

func (ctx *SceneContext) Renderer() *sdl.Renderer {
	if ctx == nil || ctx.window == nil {
		return nil
	}
	return ctx.window.renderer
}

func (ctx *SceneContext) RendererName() string {
	if ctx == nil || ctx.window == nil {
		return ""
	}
	return ctx.window.RendererName()
}

func (ctx *SceneContext) OutputSize() (float32, float32) {
	if ctx == nil || ctx.window == nil {
		return 0, 0
	}
	w, h := ctx.window.outputSize()
	return float32(w), float32(h)
}

func (ctx *SceneContext) PixelDensity() float32 {
	if ctx == nil || ctx.window == nil {
		return 1
	}
	return float32(ctx.window.pixelDensity())
}

func (ctx *SceneContext) DisplayScale() float32 {
	if ctx == nil || ctx.window == nil || ctx.window.handle == nil {
		return 1
	}
	if scale, err := ctx.window.handle.DisplayScale(); err == nil && scale > 0 {
		return scale
	}
	return 1
}

func (ctx *SceneContext) Modifiers() Modifiers {
	return modifiers(sdl.GetModState())
}

func (ctx *SceneContext) SetTitle(title string) {
	if ctx != nil && ctx.window != nil {
		ctx.window.SetTitle(title)
	}
}

func (ctx *SceneContext) SetIcon(icon image.Image) error {
	if ctx == nil || ctx.window == nil {
		return ErrWindowNotInitialized
	}
	return ctx.window.SetIcon(icon)
}

func (ctx *SceneContext) SetTextInput(enabled bool) error {
	if ctx == nil || ctx.window == nil || ctx.window.handle == nil {
		return ErrWindowNotInitialized
	}
	ctx.window.sceneTextInput = enabled
	return ctx.window.syncTextInput()
}

// SetUI attaches a retained UI to the scene. It is laid out each frame, drawn
// after the scene, and routed according to UIInputPolicy.
func (ctx *SceneContext) SetUI(ui *components.UI) {
	if ctx == nil || ctx.window == nil {
		return
	}
	ctx.window.setUI(ui)
}

// ClearUI detaches the scene's retained UI and releases renderer resources it
// created. The UI value can be attached again later.
func (ctx *SceneContext) ClearUI() {
	if ctx == nil || ctx.window == nil {
		return
	}
	ctx.window.setUI(nil)
}

// UI returns the retained UI currently attached to the scene, if any.
func (ctx *SceneContext) UI() *components.UI {
	if ctx == nil || ctx.window == nil {
		return nil
	}
	return ctx.window.ui
}

// OpenContextMenu opens a menu registered with the attached retained UI. The
// coordinates use the same renderer-pixel space as Scene mouse events; Goak
// converts them to UI space, clamps the menu to the viewport, and closes any
// other popup first. It reports whether the menu belongs to the attached UI.
func (ctx *SceneContext) OpenContextMenu(menu *components.ContextMenu, x, y float32) bool {
	if ctx == nil || ctx.window == nil || ctx.window.ui == nil || menu == nil {
		return false
	}
	found := false
	for _, candidate := range ctx.window.ui.ContextMenus() {
		if candidate == menu {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	ctx.window.closePopups()
	uiX, uiY := ctx.UIPoint(x, y)
	menu.OpenAt(uiX, uiY, ctx.UIViewport())
	return true
}

// CloseUIPopups closes retained context menus, menu-bar menus, and dropdowns
// attached to the scene.
func (ctx *SceneContext) CloseUIPopups() {
	if ctx != nil && ctx.window != nil {
		ctx.window.closePopups()
	}
}

// SetUIInputPolicy changes how the attached retained UI receives input.
func (ctx *SceneContext) SetUIInputPolicy(policy UIInputPolicy) error {
	if ctx == nil {
		return ErrWindowNotInitialized
	}
	switch policy {
	case UIInputOverlay, UIInputModal, UIInputPassthrough:
		ctx.uiInputPolicy = policy
	default:
		return ErrInvalidUIInputPolicy
	}
	if ctx.window != nil && ctx.uiInputPolicy == UIInputPassthrough {
		ctx.window.reconcileUIInputState()
	}
	return nil
}

// UIInputPolicy returns the retained UI input policy. The default is
// UIInputOverlay.
func (ctx *SceneContext) UIInputPolicy() UIInputPolicy {
	if ctx == nil {
		return UIInputOverlay
	}
	return ctx.uiInputPolicy
}

// UIScale returns the logical-to-renderer scale used by the attached retained
// UI. It includes root, window, and optional automatic DPI scaling.
func (ctx *SceneContext) UIScale() float32 {
	if ctx == nil || ctx.window == nil {
		return 1
	}
	return float32(ctx.window.renderScale())
}

// UIViewport returns the retained UI's logical viewport.
func (ctx *SceneContext) UIViewport() layout.Rect {
	if ctx == nil || ctx.window == nil {
		return layout.Rect{}
	}
	width, height := ctx.window.outputSize()
	scale := ctx.window.renderScale()
	return layout.Rect{W: float64(width) / scale, H: float64(height) / scale}
}

// UIPoint converts renderer-pixel scene coordinates to retained UI logical
// coordinates. This is useful when a scene opens a retained popup in response
// to an event it handled itself.
func (ctx *SceneContext) UIPoint(x, y float32) (float64, float64) {
	if ctx == nil || ctx.window == nil {
		return float64(x), float64(y)
	}
	scale := ctx.window.renderScale()
	return float64(x) / scale, float64(y) / scale
}

func (ctx *SceneContext) CaptureMouse(capture bool) error {
	if ctx == nil || ctx.window == nil {
		return ErrWindowNotInitialized
	}
	return ctx.window.setSceneMouseCapture(capture)
}

// UIFontError reports the current retained font configuration failure, if any.
func (ctx *SceneContext) UIFontError() error {
	if ctx == nil || ctx.window == nil {
		return nil
	}
	return ctx.window.UIFontError()
}

func (ctx *SceneContext) ClipboardText() (string, error) {
	return sdl.GetClipboardText()
}

func (ctx *SceneContext) SetClipboardText(text string) error {
	return sdl.SetClipboardText(text)
}

func (ctx *SceneContext) SetCursor(cursor Cursor) error {
	if ctx == nil || ctx.window == nil {
		return ErrWindowNotInitialized
	}
	if cursor == ctx.currentCursor && len(ctx.cursors) > 0 {
		return nil
	}
	if ctx.cursors == nil {
		ctx.cursors = make(map[Cursor]*sdl.Cursor)
	}
	native := ctx.cursors[cursor]
	if native == nil {
		created, err := sdl.CreateSystemCursor(systemCursor(cursor))
		if err != nil {
			return err
		}
		ctx.cursors[cursor] = created
		native = created
	}
	if err := sdl.SetCursor(native); err != nil {
		return err
	}
	ctx.currentCursor = cursor
	return nil
}

func (ctx *SceneContext) Quit() {
	if ctx != nil && ctx.window != nil {
		ctx.window.running = false
	}
}

func (ctx *SceneContext) close() {
	if ctx.window != nil {
		ctx.window.sceneTextInput = false
		ctx.window.sceneMouseCapture = false
		ctx.window.setUI(nil)
		_ = ctx.window.syncTextInput()
		_ = ctx.window.syncMouseCapture()
	}
	for _, cursor := range ctx.cursors {
		cursor.Destroy()
	}
	ctx.cursors = nil
}

func systemCursor(cursor Cursor) sdl.SystemCursor {
	switch cursor {
	case CursorText:
		return sdl.SYSTEM_CURSOR_TEXT
	case CursorPointer:
		return sdl.SYSTEM_CURSOR_POINTER
	case CursorResizeEW:
		return sdl.SYSTEM_CURSOR_EW_RESIZE
	case CursorResizeNS:
		return sdl.SYSTEM_CURSOR_NS_RESIZE
	default:
		return sdl.SYSTEM_CURSOR_DEFAULT
	}
}

func modifiers(mod sdl.Keymod) Modifiers {
	result := Modifiers{
		Control: mod&sdl.KMOD_CTRL != 0,
		Alt:     mod&sdl.KMOD_ALT != 0,
		Shift:   mod&sdl.KMOD_SHIFT != 0,
		Super:   mod&sdl.KMOD_GUI != 0,
	}
	result.Primary = result.Control || (runtime.GOOS == "darwin" && result.Super)
	return result
}

func keyChord(key sdl.Keycode, mod sdl.Keymod) string {
	name := keyName(key)
	if name == "" {
		return ""
	}
	modifiers := modifiers(mod)
	var chord strings.Builder
	if modifiers.Primary {
		chord.WriteString("ctrl+")
	}
	if modifiers.Alt {
		chord.WriteString("alt+")
	}
	if modifiers.Shift {
		chord.WriteString("shift+")
	}
	chord.WriteString(name)
	return chord.String()
}

func keyName(key sdl.Keycode) string {
	switch key {
	case sdl.K_RETURN, sdl.K_KP_ENTER:
		return "return"
	case sdl.K_ESCAPE:
		return "escape"
	case sdl.K_BACKSPACE:
		return "backspace"
	case sdl.K_DELETE:
		return "delete"
	case sdl.K_TAB:
		return "tab"
	case sdl.K_SPACE:
		return "space"
	case sdl.K_LEFT:
		return "left"
	case sdl.K_RIGHT:
		return "right"
	case sdl.K_UP:
		return "up"
	case sdl.K_DOWN:
		return "down"
	case sdl.K_HOME:
		return "home"
	case sdl.K_END:
		return "end"
	case sdl.K_PAGEUP:
		return "pageup"
	case sdl.K_PAGEDOWN:
		return "pagedown"
	case sdl.K_INSERT:
		return "insert"
	case sdl.K_F1, sdl.K_F2, sdl.K_F3, sdl.K_F4, sdl.K_F5, sdl.K_F6,
		sdl.K_F7, sdl.K_F8, sdl.K_F9, sdl.K_F10, sdl.K_F11, sdl.K_F12:
		return "f" + strconv.Itoa(int(key-sdl.K_F1)+1)
	}
	if key < 0x20 || key >= 0x7f {
		return ""
	}
	switch rune(key) {
	case '/':
		return "slash"
	case '=', '+':
		return "equals"
	case '-':
		return "minus"
	case '\\':
		return "backslash"
	case '\'':
		return "quote"
	case '`':
		return "backquote"
	case ',':
		return "comma"
	case '.':
		return "period"
	case ';':
		return "semicolon"
	case '[':
		return "leftbracket"
	case ']':
		return "rightbracket"
	default:
		return string(rune(key))
	}
}
