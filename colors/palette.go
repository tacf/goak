package colors

import (
	"image/color"
	"strconv"
	"strings"
)

// Color is the user-facing color type for goak.
type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

// RGBA creates a color from RGBA channels.
func RGBA(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// RGB creates an opaque color from RGB channels.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// RGBA makes Color satisfy color.Color.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}

// NRGBA returns the standard library NRGBA value.
func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// ParseHex parses #RGB or #RRGGBB colors.
func ParseHex(s string) (Color, bool) {
	h := strings.TrimSpace(strings.TrimPrefix(s, "#"))
	switch len(h) {
	case 3:
		r, errR := strconv.ParseUint(strings.Repeat(string(h[0]), 2), 16, 8)
		g, errG := strconv.ParseUint(strings.Repeat(string(h[1]), 2), 16, 8)
		b, errB := strconv.ParseUint(strings.Repeat(string(h[2]), 2), 16, 8)
		if errR == nil && errG == nil && errB == nil {
			return RGB(uint8(r), uint8(g), uint8(b)), true
		}
	case 6:
		v, err := strconv.ParseUint(h, 16, 32)
		if err == nil {
			return RGB(uint8(v>>16), uint8(v>>8), uint8(v)), true
		}
	}
	return Color{}, false
}

// HexOr parses a hex color and falls back if invalid.
func HexOr(s string, fallback Color) Color {
	if c, ok := ParseHex(s); ok {
		return c
	}
	return fallback
}

var (
	Transparent = RGBA(0, 0, 0, 0)
	Black       = RGB(0, 0, 0)
	White       = RGB(255, 255, 255)

	LightGray  = RGB(200, 200, 200)
	Gray       = RGB(130, 130, 130)
	DarkGray   = RGB(80, 80, 80)
	Yellow     = RGB(253, 249, 0)
	Gold       = RGB(255, 203, 0)
	Orange     = RGB(255, 161, 0)
	Pink       = RGB(255, 109, 194)
	Red        = RGB(230, 41, 55)
	Maroon     = RGB(190, 33, 55)
	Green      = RGB(0, 228, 48)
	Lime       = RGB(0, 158, 47)
	DarkGreen  = RGB(0, 117, 44)
	SkyBlue    = RGB(102, 191, 255)
	Blue       = RGB(0, 121, 241)
	DarkBlue   = RGB(0, 82, 172)
	Purple     = RGB(200, 122, 255)
	Violet     = RGB(135, 60, 190)
	DarkPurple = RGB(112, 31, 126)
	Beige      = RGB(211, 176, 131)
	Brown      = RGB(127, 106, 79)
	DarkBrown  = RGB(76, 63, 47)
	Magenta    = RGB(255, 0, 255)
	RayWhite   = RGB(245, 245, 245)
)

var named = map[string]Color{
	"transparent": Transparent,
	"black":       Black,
	"white":       White,
	"lightgray":   LightGray,
	"gray":        Gray,
	"darkgray":    DarkGray,
	"yellow":      Yellow,
	"gold":        Gold,
	"orange":      Orange,
	"pink":        Pink,
	"red":         Red,
	"maroon":      Maroon,
	"green":       Green,
	"lime":        Lime,
	"darkgreen":   DarkGreen,
	"skyblue":     SkyBlue,
	"blue":        Blue,
	"darkblue":    DarkBlue,
	"purple":      Purple,
	"violet":      Violet,
	"darkpurple":  DarkPurple,
	"beige":       Beige,
	"brown":       Brown,
	"darkbrown":   DarkBrown,
	"magenta":     Magenta,
	"raywhite":    RayWhite,
}

// ByName returns a named color using case-insensitive lookup.
func ByName(name string) (Color, bool) {
	c, ok := named[strings.ToLower(strings.TrimSpace(name))]
	return c, ok
}

// NameOr returns a named color or fallback when not found.
func NameOr(name string, fallback Color) Color {
	if c, ok := ByName(name); ok {
		return c
	}
	return fallback
}
