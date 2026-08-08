// Package colors exposes Goak's color type, constructors, and named palette.
package colors

import internal "goak/internal/goak/colors"

type Color = internal.Color

var (
	RGBA     = internal.RGBA
	RGB      = internal.RGB
	ParseHex = internal.ParseHex
	HexOr    = internal.HexOr
	ByName   = internal.ByName
	NameOr   = internal.NameOr

	Transparent = internal.Transparent
	Black       = internal.Black
	White       = internal.White
	LightGray   = internal.LightGray
	Gray        = internal.Gray
	DarkGray    = internal.DarkGray
	Yellow      = internal.Yellow
	Gold        = internal.Gold
	Orange      = internal.Orange
	Pink        = internal.Pink
	Red         = internal.Red
	Maroon      = internal.Maroon
	Green       = internal.Green
	Lime        = internal.Lime
	DarkGreen   = internal.DarkGreen
	SkyBlue     = internal.SkyBlue
	Blue        = internal.Blue
	DarkBlue    = internal.DarkBlue
	Purple      = internal.Purple
	Violet      = internal.Violet
	DarkPurple  = internal.DarkPurple
	Beige       = internal.Beige
	Brown       = internal.Brown
	DarkBrown   = internal.DarkBrown
	Magenta     = internal.Magenta
	RayWhite    = internal.RayWhite
)
