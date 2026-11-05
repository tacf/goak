// Package sdl3 implements goak.Backend on SDL3 through go-sdl3 (pure Go,
// no cgo). It is the only package that imports SDL.
package sdl3

import (
	"errors"
	"fmt"
	"math"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/tacf/goak"
	"github.com/tacf/goak/colors"
)

// maxDT caps the frame delta so a stall does not produce a huge step.
const maxDT = 0.25

// pxEps absorbs float32 error before floor and ceil of pixel edges.
const pxEps = 1e-3

var errIndexRange = errors.New("sdl3: index out of range")

var _ goak.Backend = (*Backend)(nil)

// Config selects the window and renderer.
type Config struct {
	Title     string
	Width     int
	Height    int
	Renderer  string // "software" (default), "gpu", "auto" or an SDL driver name
	Resizable bool
	// NoHighDPI disables the high pixel density window flag.
	NoHighDPI bool
	// NoVSync disables vertical sync.
	NoVSync bool
}

// Backend owns the SDL window, renderer and textures.
type Backend struct {
	window       *sdl.Window
	renderer     *sdl.Renderer
	rendererName string
	unload       func()
	textures     map[goak.TextureID]*sdl.Texture
	nextID       goak.TextureID
	lastTicks    uint64
	seenScale    float32
	closed       bool
	textInputErr error

	xy  []float32
	uv  []float32
	col []sdl.FColor
	i16 []uint16

	// pending is the list whose vertices are not yet converted to xy, uv, col.
	pending    *goak.DrawList
	psx, psy   float32
	rects      []sdl.FRect
	noFastPath bool

	// counters for tests
	nFill, nCopy, nSlow int
}

// SetFastPath enables or disables the rect and quad fast path. It is on by
// default and exists for tests and benchmarks.
func (b *Backend) SetFastPath(on bool) { b.noFastPath = !on }

var _ goak.Backend = (*Backend)(nil)

// New creates a window and renderer. Close must be called when done.
func New(cfg Config) (*Backend, error) {
	if cfg.Width <= 0 {
		cfg.Width = 800
	}
	if cfg.Height <= 0 {
		cfg.Height = 600
	}
	if cfg.Title == "" {
		cfg.Title = "Goak"
	}
	lib := binsdl.Load()
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		lib.Unload()
		return nil, fmt.Errorf("sdl3: SDL init failed: %w", err)
	}
	var flags sdl.WindowFlags
	if cfg.Resizable {
		flags |= sdl.WINDOW_RESIZABLE
	}
	if !cfg.NoHighDPI {
		flags |= sdl.WINDOW_HIGH_PIXEL_DENSITY
	}
	win, err := sdl.CreateWindow(cfg.Title, cfg.Width, cfg.Height, flags)
	if err != nil {
		sdl.Quit()
		lib.Unload()
		return nil, fmt.Errorf("sdl3: create window: %w", err)
	}
	driver := cfg.Renderer
	if driver == "" {
		driver = "software"
	}
	hint := driver
	if driver == "auto" {
		hint = ""
	}
	r, err := win.CreateRenderer(hint)
	if err != nil {
		win.Destroy()
		sdl.Quit()
		lib.Unload()
		return nil, fmt.Errorf("sdl3: create renderer %q: %w", driver, err)
	}
	if !cfg.NoVSync {
		_ = r.SetVSync(1)
	}
	_ = r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	name, nameErr := r.Name()
	if nameErr != nil {
		name = driver
	}
	b := &Backend{
		window: win, renderer: r, rendererName: name, unload: lib.Unload,
		textures: map[goak.TextureID]*sdl.Texture{}, nextID: 1,
		lastTicks: sdl.TicksNS(),
	}
	b.seenScale = b.Scale()
	// Text input stays on by default. Callers may stop it.
	b.textInputErr = b.StartTextInput()
	return b, nil
}

// TextInputError returns the error from the default StartTextInput in New,
// or nil. Creation does not fail on it.
func (b *Backend) TextInputError() error { return b.textInputErr }

// StartTextInput enables EVENT_TEXT_INPUT delivery for the window.
func (b *Backend) StartTextInput() error { return b.window.StartTextInput() }

// StopTextInput disables text input events for the window.
func (b *Backend) StopTextInput() error { return b.window.StopTextInput() }

// Close releases every native resource. It is safe to call more than once.
func (b *Backend) Close() {
	if b == nil || b.closed {
		return
	}
	b.closed = true
	for id, t := range b.textures {
		t.Destroy()
		delete(b.textures, id)
	}
	b.renderer.Destroy()
	b.window.Destroy()
	sdl.Quit()
	b.unload()
}

// RendererName returns the SDL renderer driver in use.
func (b *Backend) RendererName() string { return b.rendererName }

// Clear fills the whole target with c.
func (b *Backend) Clear(c colors.Color) error {
	if err := b.renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return err
	}
	return b.renderer.Clear()
}

// Present shows the frame.
func (b *Backend) Present() error { return b.renderer.Present() }

// DisplaySize returns the window size in logical units.
func (b *Backend) DisplaySize() goak.Vec2 {
	w, h, err := b.window.Size()
	if err != nil {
		return goak.Vec2{}
	}
	return goak.Vec2{X: float32(w), Y: float32(h)}
}

// Scale returns pixels per logical unit (the HiDPI pixel density).
func (b *Backend) Scale() float32 {
	d, err := b.window.PixelDensity()
	if err != nil || d <= 0 {
		return 1
	}
	return d
}

// DisplayScale returns the OS content scale of the window display.
func (b *Backend) DisplayScale() float32 {
	s, err := b.window.DisplayScale()
	if err != nil || s <= 0 {
		return 1
	}
	return s
}

// ScaleChanged reports the new pixel density when it differs from the one seen
// at the last call (or at New), and remembers it.
func (b *Backend) ScaleChanged() (float32, bool) {
	s := b.Scale()
	if s == b.seenScale {
		return s, false
	}
	b.seenScale = s
	return s, true
}

// RenderDrawData renders dd at the current pixel density of the window, then
// returns the render error. Call Clear before and Present after.
func (b *Backend) RenderDrawData(dd *goak.DrawData) error {
	if dd == nil {
		return nil
	}
	c := *dd
	s := b.Scale()
	c.Scale = goak.Vec2{X: s, Y: s}
	if c.DisplaySize == (goak.Vec2{}) {
		c.DisplaySize = b.DisplaySize()
	}
	return b.Render(&c)
}

// DrawData wraps lists with the current display size and scale.
func (b *Backend) DrawData(lists ...*goak.DrawList) *goak.DrawData {
	s := b.Scale()
	return &goak.DrawData{Lists: lists, DisplaySize: b.DisplaySize(), Scale: goak.Vec2{X: s, Y: s}}
}

// NewTexture uploads RGBA32 pixels and returns a texture id.
func (b *Backend) NewTexture(w, h int, rgba []byte) (goak.TextureID, error) {
	if w <= 0 || h <= 0 {
		return 0, errors.New("sdl3: invalid texture size")
	}
	if len(rgba) < w*h*4 {
		return 0, errors.New("sdl3: pixel data too short")
	}
	t, err := b.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA32, sdl.TEXTUREACCESS_STATIC, w, h)
	if err != nil {
		return 0, fmt.Errorf("sdl3: create texture: %w", err)
	}
	_ = t.SetBlendMode(sdl.BLENDMODE_BLEND)
	_ = t.SetScaleMode(sdl.SCALEMODE_LINEAR)
	if err := t.Update(nil, rgba, int32(w*4)); err != nil {
		t.Destroy()
		return 0, fmt.Errorf("sdl3: upload texture: %w", err)
	}
	id := b.nextID
	b.nextID++
	b.textures[id] = t
	return id, nil
}

// UpdateTexture replaces the pixels of a texture. A size change recreates it
// and keeps the id.
func (b *Backend) UpdateTexture(id goak.TextureID, w, h int, rgba []byte) error {
	old, ok := b.textures[id]
	if !ok {
		return fmt.Errorf("sdl3: unknown texture %d", id)
	}
	if w <= 0 || h <= 0 || len(rgba) < w*h*4 {
		return errors.New("sdl3: invalid texture data")
	}
	if ow, oh := int(old.W), int(old.H); ow == w && oh == h {
		return old.Update(nil, rgba, int32(w*4))
	}
	nid, err := b.NewTexture(w, h, rgba)
	if err != nil {
		return err
	}
	old.Destroy()
	b.textures[id] = b.textures[nid]
	delete(b.textures, nid)
	return nil
}

// DestroyTexture frees a texture. Unknown ids are ignored.
func (b *Backend) DestroyTexture(id goak.TextureID) {
	if t, ok := b.textures[id]; ok {
		t.Destroy()
		delete(b.textures, id)
	}
}

// Render draws every list of dd. Positions are scaled by dd.Scale into pixels.
func (b *Backend) Render(dd *goak.DrawData) error {
	if dd == nil {
		return nil
	}
	sx, sy := dd.Scale.X, dd.Scale.Y
	if sx <= 0 {
		sx = 1
	}
	if sy <= 0 {
		sy = 1
	}
	ow, oh, err := b.renderer.CurrentOutputSize()
	if err != nil {
		return err
	}
	defer b.renderer.SetClipRect(nil)
	defer func() { b.pending = nil }()
	var firstErr error
	for _, dl := range dd.Lists {
		if dl == nil || len(dl.Cmds) == 0 {
			continue
		}
		b.fillVertices(dl.Vertices, sx, sy)
		for i := range dl.Cmds {
			if err := b.renderCmd(dl, &dl.Cmds[i], sx, sy, int(ow), int(oh)); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// convert fills xy, uv and col for the pending list once, on first use.
func (b *Backend) convert() {
	if b.pending != nil {
		b.fillVertices(b.pending.Vertices, b.psx, b.psy)
		b.pending = nil
	}
}

func (b *Backend) fillVertices(vs []goak.Vertex, sx, sy float32) {
	n := len(vs)
	b.xy = growF(b.xy, n*2)
	b.uv = growF(b.uv, n*2)
	if cap(b.col) < n {
		b.col = make([]sdl.FColor, n)
	}
	b.col = b.col[:n]
	const inv = 1.0 / 255.0
	for i, v := range vs {
		b.xy[i*2], b.xy[i*2+1] = v.Pos.X*sx, v.Pos.Y*sy
		b.uv[i*2], b.uv[i*2+1] = v.UV[0], v.UV[1]
		b.col[i] = sdl.FColor{
			R: float32(v.Color&0xff) * inv,
			G: float32(v.Color>>8&0xff) * inv,
			B: float32(v.Color>>16&0xff) * inv,
			A: float32(v.Color>>24&0xff) * inv,
		}
	}
}

func growF(s []float32, n int) []float32 {
	if cap(s) < n {
		return make([]float32, n)
	}
	return s[:n]
}

func (b *Backend) renderCmd(dl *goak.DrawList, c *goak.Cmd, sx, sy float32, ow, oh int) error {
	if c.IdxCount < 3 || c.IdxOffset < 0 || c.IdxOffset+c.IdxCount > len(dl.Indices) {
		return nil
	}
	clip, ok := pixelClip(c.ClipRect, sx, sy, ow, oh)
	if !ok {
		return nil
	}
	var tex *sdl.Texture
	if c.TextureID != 0 {
		tex = b.textures[c.TextureID]
		if tex == nil {
			return fmt.Errorf("sdl3: unknown texture %d", c.TextureID)
		}
	}
	if err := b.renderer.SetClipRect(&clip); err != nil {
		return err
	}
	idx := dl.Indices[c.IdxOffset : c.IdxOffset+c.IdxCount]
	if c.Quads && !b.noFastPath && c.VtxOffset == 0 && len(idx)%6 == 0 {
		return b.renderQuads(dl, idx, tex, sx, sy)
	}
	if c.VtxOffset != 0 {
		shifted := make([]uint32, len(idx))
		for i, v := range idx {
			shifted[i] = v + uint32(c.VtxOffset)
		}
		idx = shifted
	}
	return b.drawGeometry(tex, idx)
}

// drawGeometry submits triangles through RenderGeometryRaw.
func (b *Backend) drawGeometry(tex *sdl.Texture, idx []uint32) error {
	b.convert()
	nv := len(b.col)
	if len(idx) < 3 {
		return nil
	}
	// Common case is one chunk, rebased into a reused buffer.
	lo, hi := int(idx[0]), int(idx[0])
	for _, v := range idx {
		lo, hi = min(lo, int(v)), max(hi, int(v))
	}
	if hi >= nv {
		return errors.New("sdl3: index out of range")
	}
	if hi-lo+1 <= maxChunkVerts {
		n := len(idx) / 3 * 3
		if cap(b.i16) < n {
			b.i16 = make([]uint16, n)
		}
		rel := b.i16[:n]
		for i := range rel {
			rel[i] = uint16(int(idx[i]) - lo)
		}
		return b.renderer.RenderGeometryRaw(tex,
			b.xy[lo*2:(hi+1)*2], 8,
			b.col[lo:hi+1], 16,
			b.uv[lo*2:(hi+1)*2], 8,
			rel)
	}
	for _, ch := range splitIndices(idx, maxChunkVerts) {
		if ch.VtxMax >= nv {
			return errors.New("sdl3: index out of range")
		}
		lo, hi := ch.VtxMin, ch.VtxMax+1
		err := b.renderer.RenderGeometryRaw(tex,
			b.xy[lo*2:hi*2], 8,
			b.col[lo:hi], 16,
			b.uv[lo*2:hi*2], 8,
			ch.Indices)
		if err != nil {
			return err
		}
	}
	return nil
}

// pixelClip converts a logical clip rect to a pixel rect inside the output.
// It reports false when the result is empty.
func pixelClip(r goak.Rect, sx, sy float32, ow, oh int) (sdl.Rect, bool) {
	if r.Empty() {
		return sdl.Rect{}, false
	}
	x0 := clampPx(floorEps(float64(r.Min.X*sx)), ow)
	y0 := clampPx(floorEps(float64(r.Min.Y*sy)), oh)
	x1 := clampPx(ceilEps(float64(r.Max.X*sx)), ow)
	y1 := clampPx(ceilEps(float64(r.Max.Y*sy)), oh)
	if x1 <= x0 || y1 <= y0 {
		return sdl.Rect{}, false
	}
	return sdl.Rect{X: int32(x0), Y: int32(y0), W: int32(x1 - x0), H: int32(y1 - y0)}, true
}

// floorEps and ceilEps snap values within pxEps of an integer first, so
// float32 error such as 99.99999 does not shift an edge by a pixel.
func floorEps(v float64) float64 {
	if n := math.Round(v); math.Abs(v-n) < pxEps {
		return n
	}
	return math.Floor(v)
}

func ceilEps(v float64) float64 {
	if n := math.Round(v); math.Abs(v-n) < pxEps {
		return n
	}
	return math.Ceil(v)
}

func clampPx(v float64, limit int) int {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > float64(limit) {
		return limit
	}
	return int(v)
}

// ReadPixels returns the current target as tightly packed RGBA bytes.
func (b *Backend) ReadPixels() (w, h int, rgba []byte, err error) {
	s, err := b.renderer.ReadPixels(nil)
	if err != nil {
		return 0, 0, nil, err
	}
	defer s.Destroy()
	if s.Format != sdl.PIXELFORMAT_RGBA32 {
		c, cerr := s.Convert(sdl.PIXELFORMAT_RGBA32)
		if cerr != nil {
			return 0, 0, nil, cerr
		}
		defer c.Destroy()
		s = c
	}
	w, h = int(s.W), int(s.H)
	src := s.Pixels()
	rgba = make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		copy(rgba[y*w*4:(y+1)*w*4], src[y*int(s.Pitch):])
	}
	return w, h, rgba, nil
}

// NewFrame starts an input frame, It calls
// in.BeginFrame, sets in.DT from the time since the previous call, drains the
// SDL queue into in and reports whether the app should keep running. It
// returns false once the window was closed or a quit was requested, and in.Quit
// is then true. Escape and focus policy is left to the caller.
func (b *Backend) NewFrame(in *goak.Input) bool {
	in.BeginFrame()
	now := sdl.TicksNS()
	in.DT = float32(min(float64(now-b.lastTicks)/1e9, maxDT))
	b.lastTicks = now
	var ev sdl.Event
	lostFocus := false
	for sdl.PollEvent(&ev) {
		switch ev.Type {
		case sdl.EVENT_WINDOW_FOCUS_LOST:
			// Release held state so keys and buttons do not stick.
			in.KeyDown = [goak.KeyCount]bool{}
			in.MouseDown = [goak.MouseButtonCount]bool{}
			lostFocus = true
		case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_CLOSE_REQUESTED:
			in.Quit = true
		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			if k := ev.KeyboardEvent(); k != nil {
				in.SetKey(translateKey(k.Key), ev.Type == sdl.EVENT_KEY_DOWN)
				in.Mods = translateMods(k.Mod)
			}
		case sdl.EVENT_TEXT_INPUT:
			if t := ev.TextInputEvent(); t != nil {
				in.Text += t.Text
			}
		case sdl.EVENT_MOUSE_MOTION:
			if m := ev.MouseMotionEvent(); m != nil {
				in.MousePos = [2]float32{m.X, m.Y}
			}
		case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
			if m := ev.MouseButtonEvent(); m != nil {
				in.MousePos = [2]float32{m.X, m.Y}
				if btn, ok := translateButton(m.Button); ok {
					in.SetMouseButton(btn, ev.Type == sdl.EVENT_MOUSE_BUTTON_DOWN, m.Clicks)
				}
			}
		case sdl.EVENT_MOUSE_WHEEL:
			if m := ev.MouseWheelEvent(); m != nil {
				in.MousePos = [2]float32{m.MouseX, m.MouseY}
				in.Wheel[0] += m.X
				in.Wheel[1] += m.Y
			}
		}
	}
	in.Mods = translateMods(sdl.GetModState())
	if lostFocus {
		in.Mods = 0
	}
	return !in.Quit
}
