package goak

// RendererDriver selects the SDL 2D rendering driver used by a window.
//
// The predefined values cover SDL's built-in drivers. A custom SDL driver name
// or comma-separated fallback list can be supplied with RendererDriver("name").
type RendererDriver string

const (
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
