package components

import (
	"errors"
	"image"
	imagedraw "image/draw"
	"reflect"
	"runtime"

	"github.com/tacf/goak/internal/goak/layout"

	"github.com/Zyko0/go-sdl3/sdl"
)

// ImageFit controls how an image is placed inside its layout bounds.
type ImageFit int

const (
	// ImageFitContain preserves aspect ratio and keeps the entire image visible.
	ImageFitContain ImageFit = iota
	// ImageFitCover preserves aspect ratio and fills the bounds, cropping as needed.
	ImageFitCover
	// ImageFitStretch fills the bounds without preserving aspect ratio.
	ImageFitStretch
	// ImageFitNone draws the image at its native size, centered in the bounds.
	ImageFitNone
)

// Image displays a standard Go image.Image.
type Image struct {
	c       *layout.Container
	source  image.Image
	fit     ImageFit
	texture *sdl.Texture
	render  *sdl.Renderer
	width   int
	height  int
}

// NewImage creates a standalone image component.
func NewImage(width, height layout.Size, source image.Image) *Image {
	return &Image{
		c:      layout.NewContainer(width, height),
		source: source,
		fit:    ImageFitContain,
	}
}

// Container returns the image's underlying layout node.
func (i *Image) Container() *layout.Container { return i.c }

// Bounds returns the computed image bounds.
func (i *Image) Bounds() layout.Rect { return i.c.Bounds }

// Source returns the displayed source image.
func (i *Image) Source() image.Image { return i.source }

// SetSource changes the displayed image. A nil image leaves the component blank.
func (i *Image) SetSource(source image.Image) {
	if i == nil {
		return
	}
	i.releaseTexture()
	i.source = source
}

// Fit returns the image placement mode.
func (i *Image) Fit() ImageFit { return i.fit }

// SetFit changes how the image is placed inside its bounds.
func (i *Image) SetFit(fit ImageFit) {
	switch fit {
	case ImageFitContain, ImageFitCover, ImageFitStretch, ImageFitNone:
		i.fit = fit
	default:
		i.fit = ImageFitContain
	}
}

// Close releases renderer resources held by the image. The source remains set
// and will be uploaded again if the component is drawn later.
func (i *Image) Close() {
	if i != nil {
		i.releaseTexture()
	}
}

func (i *Image) releaseTexture() {
	if i.texture != nil {
		i.texture.Destroy()
	}
	i.texture = nil
	i.render = nil
	i.width = 0
	i.height = 0
}

func (i *Image) ensureTexture(renderer *sdl.Renderer) error {
	if i.texture != nil && i.render == renderer {
		return nil
	}
	i.releaseTexture()
	if renderer == nil || i.source == nil || isNilStandardImage(i.source) {
		return nil
	}
	bounds := i.source.Bounds()
	if bounds.Empty() {
		return errors.New("goak: component image is empty")
	}
	rgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	imagedraw.Draw(rgba, rgba.Bounds(), i.source, bounds.Min, imagedraw.Src)
	surface, err := sdl.CreateSurfaceFrom(
		rgba.Bounds().Dx(),
		rgba.Bounds().Dy(),
		sdl.PIXELFORMAT_RGBA32,
		rgba.Pix,
		rgba.Stride,
	)
	if err != nil {
		return err
	}
	texture, err := renderer.CreateTextureFromSurface(surface)
	surface.Destroy()
	runtime.KeepAlive(rgba)
	if err != nil {
		return err
	}
	_ = texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	_ = texture.SetScaleMode(sdl.SCALEMODE_LINEAR)
	i.texture = texture
	i.render = renderer
	i.width = bounds.Dx()
	i.height = bounds.Dy()
	return nil
}

// Draw renders the image.
func (i *Image) Draw(renderer *sdl.Renderer) {
	if i == nil || renderer == nil || i.source == nil || i.Bounds().W <= 0 || i.Bounds().H <= 0 {
		return
	}
	if err := i.ensureTexture(renderer); err != nil || i.texture == nil {
		return
	}
	source, destination := imageRects(i.Bounds(), float64(i.width), float64(i.height), i.fit)
	_ = renderer.RenderTexture(i.texture, source, &destination)
}

func imageRects(bounds layout.Rect, imageW, imageH float64, fit ImageFit) (*sdl.FRect, sdl.FRect) {
	destination := sdl.FRect{
		X: float32(bounds.X),
		Y: float32(bounds.Y),
		W: float32(bounds.W),
		H: float32(bounds.H),
	}
	if imageW <= 0 || imageH <= 0 || bounds.W <= 0 || bounds.H <= 0 {
		return nil, destination
	}

	switch fit {
	case ImageFitStretch:
		return nil, destination
	case ImageFitNone:
		destination.W = float32(imageW)
		destination.H = float32(imageH)
		destination.X = float32(bounds.X + (bounds.W-imageW)/2)
		destination.Y = float32(bounds.Y + (bounds.H-imageH)/2)
		return nil, destination
	case ImageFitCover:
		scale := max(bounds.W/imageW, bounds.H/imageH)
		sourceW := bounds.W / scale
		sourceH := bounds.H / scale
		source := &sdl.FRect{
			X: float32((imageW - sourceW) / 2),
			Y: float32((imageH - sourceH) / 2),
			W: float32(sourceW),
			H: float32(sourceH),
		}
		return source, destination
	default:
		scale := min(bounds.W/imageW, bounds.H/imageH)
		drawW := imageW * scale
		drawH := imageH * scale
		destination.X = float32(bounds.X + (bounds.W-drawW)/2)
		destination.Y = float32(bounds.Y + (bounds.H-drawH)/2)
		destination.W = float32(drawW)
		destination.H = float32(drawH)
		return nil, destination
	}
}

func isNilStandardImage(source image.Image) bool {
	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
