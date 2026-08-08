// Package components exposes Goak's retained-mode widgets. The aliases keep
// one canonical implementation while allowing applications outside the Goak
// module to build and compose widget trees.
package components

import internal "github.com/tacf/goak/internal/goak/components"

type (
	Action                  = internal.Action
	Button                  = internal.Button
	ButtonClickEvent        = internal.ButtonClickEvent
	ButtonTheme             = internal.ButtonTheme
	Checkbox                = internal.Checkbox
	CheckboxChangedEvent    = internal.CheckboxChangedEvent
	CheckboxTheme           = internal.CheckboxTheme
	ContextMenu             = internal.ContextMenu
	ContextMenuActionEvent  = internal.ContextMenuActionEvent
	ContextMenuItem         = internal.ContextMenuItem
	ContextMenuItemKind     = internal.ContextMenuItemKind
	ContextMenuTheme        = internal.ContextMenuTheme
	DebugTheme              = internal.DebugTheme
	Dropdown                = internal.Dropdown
	DropdownChangedEvent    = internal.DropdownChangedEvent
	DropdownOption          = internal.DropdownOption
	DropdownTheme           = internal.DropdownTheme
	Image                   = internal.Image
	ImageFit                = internal.ImageFit
	Label                   = internal.Label
	LabelTheme              = internal.LabelTheme
	MenuActionEvent         = internal.MenuActionEvent
	MenuBar                 = internal.MenuBar
	MenuBarWidthMode        = internal.MenuBarWidthMode
	MenuEntry               = internal.MenuEntry
	MenuEntryKind           = internal.MenuEntryKind
	MenuItem                = internal.MenuItem
	MenuTheme               = internal.MenuTheme
	Panel                   = internal.Panel
	PanelTheme              = internal.PanelTheme
	RadioChangedEvent       = internal.RadioChangedEvent
	RadioGroup              = internal.RadioGroup
	RadioOption             = internal.RadioOption
	RadioTheme              = internal.RadioTheme
	Root                    = internal.Root
	Slider                  = internal.Slider
	SliderChangedEvent      = internal.SliderChangedEvent
	SliderTheme             = internal.SliderTheme
	TextArea                = internal.TextArea
	TextAreaChangedEvent    = internal.TextAreaChangedEvent
	TextAreaTheme           = internal.TextAreaTheme
	TextInput               = internal.TextInput
	TextInputChangedEvent   = internal.TextInputChangedEvent
	TextInputSubmittedEvent = internal.TextInputSubmittedEvent
	TextInputTheme          = internal.TextInputTheme
	Theme                   = internal.Theme
	UI                      = internal.UI
)

const (
	MenuBarWidthAuto = internal.MenuBarWidthAuto
	MenuBarWidthFull = internal.MenuBarWidthFull

	MenuEntryItem      = internal.MenuEntryItem
	MenuEntrySeparator = internal.MenuEntrySeparator

	ContextMenuItemAction    = internal.ContextMenuItemAction
	ContextMenuItemSeparator = internal.ContextMenuItemSeparator

	ImageFitContain = internal.ImageFitContain
	ImageFitCover   = internal.ImageFitCover
	ImageFitStretch = internal.ImageFitStretch
	ImageFitNone    = internal.ImageFitNone
)

var (
	DefaultTheme                 = internal.DefaultTheme
	NewButton                    = internal.NewButton
	NewCheckbox                  = internal.NewCheckbox
	NewContextMenu               = internal.NewContextMenu
	NewContextMenuAction         = internal.NewContextMenuAction
	NewContextMenuSeparator      = internal.NewContextMenuSeparator
	NewDisabledContextMenuAction = internal.NewDisabledContextMenuAction
	NewDropdown                  = internal.NewDropdown
	NewImage                     = internal.NewImage
	NewLabel                     = internal.NewLabel
	NewMenuBar                   = internal.NewMenuBar
	NewPanel                     = internal.NewPanel
	NewRadioGroup                = internal.NewRadioGroup
	NewSlider                    = internal.NewSlider
	NewTextArea                  = internal.NewTextArea
	NewTextInput                 = internal.NewTextInput
	NewUI                        = internal.NewUI
)
