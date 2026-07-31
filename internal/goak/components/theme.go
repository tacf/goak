package components

import "goak/internal/goak/colors"

// Theme contains all colors used to render a UI.
type Theme struct {
	Background  colors.Color
	Debug       DebugTheme
	Panel       PanelTheme
	Label       LabelTheme
	Button      ButtonTheme
	Checkbox    CheckboxTheme
	RadioGroup  RadioTheme
	Slider      SliderTheme
	Dropdown    DropdownTheme
	TextInput   TextInputTheme
	TextArea    TextAreaTheme
	MenuBar     MenuTheme
	ContextMenu ContextMenuTheme
}

// DebugTheme controls debug overlay colors.
type DebugTheme struct {
	Outline colors.Color
	Text    colors.Color
}

// PanelTheme controls panel drawing colors.
type PanelTheme struct {
	DefaultFill colors.Color
	Stroke      colors.Color
}

// LabelTheme controls label drawing colors.
type LabelTheme struct {
	Text colors.Color
}

// ButtonTheme controls button drawing colors.
type ButtonTheme struct {
	Fill   colors.Color
	Stroke colors.Color
	Text   colors.Color
}

// CheckboxTheme controls checkbox drawing colors.
type CheckboxTheme struct {
	BoxFill      colors.Color
	BoxStroke    colors.Color
	CheckFill    colors.Color
	Text         colors.Color
	HoverOverlay colors.Color
}

// RadioTheme controls radio group drawing colors.
type RadioTheme struct {
	CircleFill   colors.Color
	CircleStroke colors.Color
	SelectedFill colors.Color
	Text         colors.Color
	HoverOverlay colors.Color
}

// SliderTheme controls slider drawing colors.
type SliderTheme struct {
	TrackFill   colors.Color
	TrackStroke colors.Color
	FillColor   colors.Color
	ThumbFill   colors.Color
	ThumbStroke colors.Color
	Text        colors.Color
}

// DropdownTheme controls dropdown drawing colors.
type DropdownTheme struct {
	Fill      colors.Color
	Stroke    colors.Color
	Hover     colors.Color
	Selected  colors.Color
	Text      colors.Color
	ArrowFill colors.Color
}

// TextInputTheme controls single-line text input drawing colors.
type TextInputTheme struct {
	Fill          colors.Color
	Stroke        colors.Color
	FocusedStroke colors.Color
	Text          colors.Color
	Placeholder   colors.Color
	Selection     colors.Color
	Caret         colors.Color
}

// TextAreaTheme controls multiline text area drawing colors.
type TextAreaTheme struct {
	Fill           colors.Color
	Stroke         colors.Color
	FocusedStroke  colors.Color
	Text           colors.Color
	Placeholder    colors.Color
	Selection      colors.Color
	Caret          colors.Color
	GutterFill     colors.Color
	LineNumber     colors.Color
	ScrollbarTrack colors.Color
	ScrollbarThumb colors.Color
}

// MenuTheme controls menu bar and dropdown colors.
type MenuTheme struct {
	Fill      colors.Color
	Stroke    colors.Color
	Hover     colors.Color
	Active    colors.Color
	Text      colors.Color
	Separator colors.Color
}

// ContextMenuTheme controls context menu drawing colors.
type ContextMenuTheme struct {
	Fill         colors.Color
	Stroke       colors.Color
	Hover        colors.Color
	Text         colors.Color
	DisabledText colors.Color
	Separator    colors.Color
}

// DefaultTheme returns a complete theme matching Goak's default appearance.
func DefaultTheme() Theme {
	return Theme{
		Background: colors.Black,
		Debug: DebugTheme{
			Outline: colors.Yellow,
			Text:    colors.Yellow,
		},
		Panel: PanelTheme{
			DefaultFill: colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			Stroke:      colors.HexOr("#555", colors.RGB(85, 85, 85)),
		},
		Label: LabelTheme{
			Text: colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		},
		Button: ButtonTheme{
			Fill:   colors.HexOr("#404040", colors.RGB(64, 64, 64)),
			Stroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
			Text:   colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		},
		Checkbox: CheckboxTheme{
			BoxFill:      colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			BoxStroke:    colors.HexOr("#666", colors.RGB(102, 102, 102)),
			CheckFill:    colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			HoverOverlay: colors.RGBA(255, 255, 255, 20),
		},
		RadioGroup: RadioTheme{
			CircleFill:   colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			CircleStroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
			SelectedFill: colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			HoverOverlay: colors.RGBA(255, 255, 255, 20),
		},
		Slider: SliderTheme{
			TrackFill:   colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			TrackStroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
			FillColor:   colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			ThumbFill:   colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			ThumbStroke: colors.HexOr("#666", colors.RGB(102, 102, 102)),
			Text:        colors.HexOr("#eee", colors.RGB(238, 238, 238)),
		},
		Dropdown: DropdownTheme{
			Fill:      colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			Stroke:    colors.HexOr("#666", colors.RGB(102, 102, 102)),
			Hover:     colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
			Selected:  colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			Text:      colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			ArrowFill: colors.HexOr("#aaa", colors.RGB(170, 170, 170)),
		},
		TextInput: TextInputTheme{
			Fill:          colors.HexOr("#252525", colors.RGB(37, 37, 37)),
			Stroke:        colors.HexOr("#666", colors.RGB(102, 102, 102)),
			FocusedStroke: colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			Text:          colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			Placeholder:   colors.HexOr("#888", colors.RGB(136, 136, 136)),
			Selection:     colors.RGBA(74, 158, 255, 110),
			Caret:         colors.White,
		},
		TextArea: TextAreaTheme{
			Fill:           colors.HexOr("#202020", colors.RGB(32, 32, 32)),
			Stroke:         colors.HexOr("#666", colors.RGB(102, 102, 102)),
			FocusedStroke:  colors.HexOr("#4a9eff", colors.RGB(74, 158, 255)),
			Text:           colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			Placeholder:    colors.HexOr("#888", colors.RGB(136, 136, 136)),
			Selection:      colors.RGBA(74, 158, 255, 110),
			Caret:          colors.White,
			GutterFill:     colors.HexOr("#191919", colors.RGB(25, 25, 25)),
			LineNumber:     colors.HexOr("#888", colors.RGB(136, 136, 136)),
			ScrollbarTrack: colors.HexOr("#181818", colors.RGB(24, 24, 24)),
			ScrollbarThumb: colors.HexOr("#666", colors.RGB(102, 102, 102)),
		},
		MenuBar: MenuTheme{
			Fill:      colors.HexOr("#202020", colors.RGB(32, 32, 32)),
			Stroke:    colors.HexOr("#525252", colors.RGB(82, 82, 82)),
			Hover:     colors.HexOr("#2f2f2f", colors.RGB(47, 47, 47)),
			Active:    colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
			Text:      colors.HexOr("#f0f0f0", colors.RGB(240, 240, 240)),
			Separator: colors.HexOr("#606060", colors.RGB(96, 96, 96)),
		},
		ContextMenu: ContextMenuTheme{
			Fill:         colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
			Stroke:       colors.HexOr("#666", colors.RGB(102, 102, 102)),
			Hover:        colors.HexOr("#3a3a3a", colors.RGB(58, 58, 58)),
			Text:         colors.HexOr("#eee", colors.RGB(238, 238, 238)),
			DisabledText: colors.HexOr("#777", colors.RGB(119, 119, 119)),
			Separator:    colors.HexOr("#555", colors.RGB(85, 85, 85)),
		},
	}
}
