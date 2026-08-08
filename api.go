// Package goak provides SDL3 application lifecycle, retained widgets, and a
// low-level scene host for custom interfaces.
package goak

import core "goak/internal/goak"

type (
	App               = core.App
	Config            = core.Config
	Cursor            = core.Cursor
	Event             = core.Event
	EventType         = core.EventType
	Modifiers         = core.Modifiers
	MouseButton       = core.MouseButton
	RendererDriver    = core.RendererDriver
	Scene             = core.Scene
	SceneCloser       = core.SceneCloser
	SceneContext      = core.SceneContext
	SceneEventHandler = core.SceneEventHandler
	SceneInitializer  = core.SceneInitializer
	SceneUpdater      = core.SceneUpdater
	Window            = core.Window
)

var (
	ErrAppStopped                 = core.ErrAppStopped
	ErrNilDispatch                = core.ErrNilDispatch
	ErrNilScene                   = core.ErrNilScene
	ErrRendererAlreadyInitialized = core.ErrRendererAlreadyInitialized
	ErrWindowNotInitialized       = core.ErrWindowNotInitialized
)

func NewApp() *App { return core.NewApp() }

func InitWindow(title string, width, height int) *Window {
	return core.InitWindow(title, width, height)
}

func RendererDrivers() ([]string, error) { return core.RendererDrivers() }

const (
	RendererAuto       = core.RendererAuto
	RendererSoftware   = core.RendererSoftware
	RendererGPU        = core.RendererGPU
	RendererDirect3D   = core.RendererDirect3D
	RendererDirect3D11 = core.RendererDirect3D11
	RendererDirect3D12 = core.RendererDirect3D12
	RendererOpenGL     = core.RendererOpenGL
	RendererOpenGLES   = core.RendererOpenGLES
	RendererOpenGLES2  = core.RendererOpenGLES2
	RendererMetal      = core.RendererMetal
	RendererVulkan     = core.RendererVulkan

	EventQuit       = core.EventQuit
	EventKeyDown    = core.EventKeyDown
	EventTextInput  = core.EventTextInput
	EventMouseDown  = core.EventMouseDown
	EventMouseUp    = core.EventMouseUp
	EventMouseMove  = core.EventMouseMove
	EventMouseWheel = core.EventMouseWheel

	MouseLeft   = core.MouseLeft
	MouseMiddle = core.MouseMiddle
	MouseRight  = core.MouseRight

	CursorDefault  = core.CursorDefault
	CursorText     = core.CursorText
	CursorPointer  = core.CursorPointer
	CursorResizeEW = core.CursorResizeEW
	CursorResizeNS = core.CursorResizeNS
)
