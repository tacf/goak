package goak

import (
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
)

// Vec2 and Rect are the draw package types.
type (
	Vec2 = draw.Vec2
	Rect = draw.Rect
)

// Col indexes Style.Colors.
type Col int

// Style colors.
const (
	ColText Col = iota
	ColWindowBg
	ColChildBg
	ColBorder
	ColTitleBg
	ColTitleBgActive
	ColFrameBg
	ColFrameBgHovered
	ColFrameBgActive
	ColButton
	ColButtonHovered
	ColButtonActive
	ColCheckMark
	ColScrollbarBg
	ColScrollbarGrab
	ColScrollbarGrabHovered
	ColScrollbarGrabActive
	ColSeparator
	ColResizeGrip
	ColResizeGripHovered
	ColResizeGripActive
	ColCloseHovered
	ColTextDisabled
	ColTextSelectedBg
	ColSliderGrab
	ColSliderGrabActive
	ColHeader
	ColHeaderHovered
	ColHeaderActive
	ColPopupBg
	ColMenuBarBg
	ColTab
	ColTabHovered
	ColTabActive
	ColProgressBar
	ColCaret
	ColPlotLines
	ColPlotLinesHovered
	ColPlotHistogram
	ColPlotHistogramHovered
	ColTableHeaderBg
	ColTableBorder
	ColTableRowBgAlt
	ColCount
)

// Style holds metrics and colors. Values are in logical units.
type Style struct {
	WindowPadding     Vec2
	ChildPadding      Vec2
	FramePadding      Vec2
	ItemSpacing       Vec2
	ItemInnerSpacing  Vec2
	WindowMinSize     Vec2
	IndentSpacing     float32
	ScrollbarSize     float32
	GrabMinSize       float32
	ResizeGripSize    float32
	WindowRounding    float32
	ChildRounding     float32
	FrameRounding     float32
	ScrollbarRounding float32
	BorderSize        float32
	PopupRounding     float32
	TabRounding       float32
	GrabRounding      float32
	// ScrollStep is the distance of one wheel notch in lines.
	ScrollStep float32
	Colors     [ColCount]colors.Color
}

// DefaultStyle returns DarkStyle.
func DefaultStyle() Style { return DarkStyle() }

func baseStyle() Style {
	return Style{
		WindowPadding:     Vec2{X: 8, Y: 8},
		ChildPadding:      Vec2{X: 4, Y: 4},
		FramePadding:      Vec2{X: 8, Y: 4},
		ItemSpacing:       Vec2{X: 8, Y: 4},
		ItemInnerSpacing:  Vec2{X: 4, Y: 4},
		WindowMinSize:     Vec2{X: 64, Y: 32},
		IndentSpacing:     16,
		ScrollbarSize:     10,
		GrabMinSize:       12,
		ResizeGripSize:    14,
		WindowRounding:    4,
		ChildRounding:     2,
		FrameRounding:     3,
		ScrollbarRounding: 3,
		BorderSize:        1,
		PopupRounding:     3,
		TabRounding:       3,
		GrabRounding:      6,
		ScrollStep:        3,
	}
}

// DarkStyle is seeded from components/theme.go and colors/palette.go.
func DarkStyle() Style {
	s := baseStyle()
	text := colors.RGB(238, 238, 238)
	accent := colors.RGB(74, 158, 255)
	stroke := colors.RGB(100, 120, 145)
	c := &s.Colors
	c[ColText] = text
	c[ColWindowBg] = colors.RGB(45, 45, 45)
	c[ColChildBg] = colors.RGB(35, 35, 35)
	c[ColBorder] = stroke
	c[ColTitleBg] = colors.RGB(45, 52, 64)
	c[ColTitleBgActive] = colors.RGB(55, 65, 80)
	c[ColFrameBg] = colors.RGB(45, 45, 45)
	c[ColFrameBgHovered] = colors.RGB(64, 64, 64)
	c[ColFrameBgActive] = colors.RGB(85, 85, 85)
	c[ColButton] = colors.RGB(64, 64, 64)
	c[ColButtonHovered] = colors.RGB(85, 85, 85)
	c[ColButtonActive] = colors.RGB(55, 80, 115)
	c[ColCheckMark] = accent
	c[ColScrollbarBg] = colors.RGBA(0, 0, 0, 60)
	c[ColScrollbarGrab] = colors.RGB(110, 120, 135)
	c[ColScrollbarGrabHovered] = colors.RGB(140, 150, 165)
	c[ColScrollbarGrabActive] = accent
	c[ColSeparator] = colors.RGB(102, 102, 102)
	c[ColResizeGrip] = colors.RGBA(110, 120, 135, 120)
	c[ColResizeGripHovered] = colors.RGB(140, 150, 165)
	c[ColResizeGripActive] = accent
	c[ColCloseHovered] = colors.Maroon
	c[ColTextDisabled] = colors.RGB(140, 140, 140)
	c[ColTextSelectedBg] = colors.RGBA(74, 158, 255, 110)
	c[ColSliderGrab] = text
	c[ColSliderGrabActive] = accent
	c[ColHeader] = colors.RGB(55, 80, 115)
	c[ColHeaderHovered] = colors.RGB(58, 58, 58)
	c[ColHeaderActive] = colors.RGB(74, 110, 160)
	c[ColPopupBg] = colors.RGB(37, 37, 37)
	c[ColMenuBarBg] = colors.RGB(45, 52, 64)
	c[ColTab] = colors.RGB(45, 52, 64)
	c[ColTabHovered] = colors.RGB(74, 110, 160)
	c[ColTabActive] = colors.RGB(55, 80, 115)
	c[ColProgressBar] = accent
	c[ColCaret] = text
	c[ColPlotLines] = colors.RGB(156, 170, 190)
	c[ColPlotLinesHovered] = colors.RGB(255, 110, 90)
	c[ColPlotHistogram] = colors.RGB(230, 180, 0)
	c[ColPlotHistogramHovered] = colors.RGB(255, 150, 0)
	c[ColTableHeaderBg] = colors.RGB(55, 65, 80)
	c[ColTableBorder] = colors.RGB(80, 84, 92)
	c[ColTableRowBgAlt] = colors.RGBA(255, 255, 255, 12)
	return s
}

// LightStyle is a light variant built on colors/palette.go values.
func LightStyle() Style {
	s := baseStyle()
	text := colors.RGB(30, 30, 30)
	accent := colors.RGB(0, 121, 241)
	c := &s.Colors
	c[ColText] = text
	c[ColWindowBg] = colors.RayWhite
	c[ColChildBg] = colors.RGB(235, 235, 235)
	c[ColBorder] = colors.RGB(160, 165, 175)
	c[ColTitleBg] = colors.RGB(205, 210, 220)
	c[ColTitleBgActive] = colors.RGB(185, 195, 212)
	c[ColFrameBg] = colors.White
	c[ColFrameBgHovered] = colors.RGB(228, 232, 240)
	c[ColFrameBgActive] = colors.RGB(205, 212, 226)
	c[ColButton] = colors.RGB(215, 218, 225)
	c[ColButtonHovered] = colors.RGB(195, 202, 216)
	c[ColButtonActive] = colors.RGB(160, 185, 225)
	c[ColCheckMark] = accent
	c[ColScrollbarBg] = colors.RGBA(0, 0, 0, 25)
	c[ColScrollbarGrab] = colors.RGB(150, 158, 172)
	c[ColScrollbarGrabHovered] = colors.RGB(120, 130, 148)
	c[ColScrollbarGrabActive] = accent
	c[ColSeparator] = colors.RGB(170, 170, 170)
	c[ColResizeGrip] = colors.RGBA(120, 130, 148, 120)
	c[ColResizeGripHovered] = colors.RGB(120, 130, 148)
	c[ColResizeGripActive] = accent
	c[ColCloseHovered] = colors.Red
	c[ColTextDisabled] = colors.RGB(130, 130, 130)
	c[ColTextSelectedBg] = colors.RGBA(0, 121, 241, 90)
	c[ColSliderGrab] = colors.RGB(110, 120, 138)
	c[ColSliderGrabActive] = accent
	c[ColHeader] = colors.RGB(170, 200, 240)
	c[ColHeaderHovered] = colors.RGB(215, 224, 240)
	c[ColHeaderActive] = colors.RGB(140, 180, 235)
	c[ColPopupBg] = colors.White
	c[ColMenuBarBg] = colors.RGB(205, 210, 220)
	c[ColTab] = colors.RGB(205, 210, 220)
	c[ColTabHovered] = colors.RGB(160, 190, 235)
	c[ColTabActive] = colors.RGB(170, 200, 240)
	c[ColProgressBar] = accent
	c[ColCaret] = text
	c[ColPlotLines] = colors.RGB(90, 100, 120)
	c[ColPlotLinesHovered] = colors.RGB(220, 70, 50)
	c[ColPlotHistogram] = colors.RGB(200, 140, 0)
	c[ColPlotHistogramHovered] = colors.RGB(230, 110, 0)
	c[ColTableHeaderBg] = colors.RGB(205, 210, 220)
	c[ColTableBorder] = colors.RGB(170, 175, 185)
	c[ColTableRowBgAlt] = colors.RGBA(0, 0, 0, 14)
	return s
}
