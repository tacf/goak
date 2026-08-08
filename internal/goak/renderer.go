package goak

import (
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
)

// RendererDriver selects the SDL 2D rendering driver used by a window.
//
// The predefined values cover SDL's built-in drivers. A custom SDL driver name
// or comma-separated fallback list can be supplied with RendererDriver("name").
type RendererDriver string

const (
	// RendererAuto lets SDL select the best available renderer.
	RendererAuto     RendererDriver = "auto"
	RendererSoftware RendererDriver = "software"
	// RendererGPU uses SDL's cross-platform GPU renderer; SDL selects the
	// appropriate GPU device backend for the operating system.
	RendererGPU        RendererDriver = "gpu"
	RendererDirect3D   RendererDriver = "direct3d"
	RendererDirect3D11 RendererDriver = "direct3d11"
	RendererDirect3D12 RendererDriver = "direct3d12"
	RendererOpenGL     RendererDriver = "opengl"
	RendererOpenGLES   RendererDriver = "opengles"
	RendererOpenGLES2  RendererDriver = "opengles2"
	RendererMetal      RendererDriver = "metal"
	RendererVulkan     RendererDriver = "vulkan"
)

func normalizeRendererDriver(driver RendererDriver) RendererDriver {
	if driver == "" {
		return RendererSoftware
	}
	return driver
}

// RendererDrivers returns the SDL renderer names available on this system.
// Call it during startup, before creating a window.
func RendererDrivers() ([]string, error) {
	library := binsdl.Load()
	defer library.Unload()
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}
	defer sdl.Quit()
	drivers := make([]string, 0, sdl.GetNumRenderDrivers())
	for i := 0; i < sdl.GetNumRenderDrivers(); i++ {
		drivers = append(drivers, sdl.GetRenderDriver(i))
	}
	return drivers, nil
}
