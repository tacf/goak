package goak

import (
	"errors"
	"fmt"
	"image"
	imagedraw "image/draw"
	"reflect"
	"runtime"

	"github.com/Zyko0/go-sdl3/sdl"
)

// SetIcon sets the native window icon from a standard Go image.
func (win *Window) SetIcon(icon image.Image) error {
	if win == nil || win.handle == nil || win.destroyed {
		return ErrWindowNotInitialized
	}
	rgba, err := iconRGBA(icon)
	if err != nil {
		return err
	}

	surface, err := sdl.CreateSurfaceFrom(
		rgba.Bounds().Dx(),
		rgba.Bounds().Dy(),
		sdl.PIXELFORMAT_RGBA32,
		rgba.Pix,
		rgba.Stride,
	)
	if err != nil {
		return fmt.Errorf("goak: create icon surface: %w", err)
	}
	setErr := win.handle.SetIcon(surface)
	surface.Destroy()
	runtime.KeepAlive(rgba)
	if setErr != nil {
		return fmt.Errorf("goak: set window icon: %w", setErr)
	}
	return nil
}

func iconRGBA(icon image.Image) (*image.NRGBA, error) {
	if icon == nil || isNilImage(icon) {
		return nil, errors.New("goak: window icon is nil")
	}
	bounds := icon.Bounds()
	if bounds.Empty() {
		return nil, errors.New("goak: window icon is empty")
	}
	rgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	imagedraw.Draw(rgba, rgba.Bounds(), icon, bounds.Min, imagedraw.Src)
	return rgba, nil
}

func isNilImage(icon image.Image) bool {
	value := reflect.ValueOf(icon)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
