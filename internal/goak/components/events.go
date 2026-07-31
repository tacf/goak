package components

// Action is a reusable semantic command for click-like controls.
// Actions run synchronously on the UI thread and should return quickly.
type Action func()

// Invoke executes the action when it is non-nil.
func (action Action) Invoke() {
	if action != nil {
		action()
	}
}

// ButtonClickEvent describes a button activation.
type ButtonClickEvent struct {
	Button *Button
}

// CheckboxChangedEvent describes a checkbox state change.
type CheckboxChangedEvent struct {
	Checkbox *Checkbox
	Previous bool
	Checked  bool
}

// RadioChangedEvent describes a radio group selection change.
type RadioChangedEvent struct {
	RadioGroup    *RadioGroup
	PreviousIndex int
	Index         int
	Option        RadioOption
}

// SliderChangedEvent describes a slider value change.
type SliderChangedEvent struct {
	Slider   *Slider
	Previous float64
	Value    float64
}

// DropdownChangedEvent describes a dropdown selection change.
type DropdownChangedEvent struct {
	Dropdown      *Dropdown
	PreviousIndex int
	Index         int
	Option        DropdownOption
}

// MenuActionEvent describes activation of a top-level or submenu action.
// SubIndex is -1 for a top-level action.
type MenuActionEvent struct {
	Menu     *MenuBar
	TopIndex int
	SubIndex int
	Label    string
}

// ContextMenuActionEvent describes activation of a context menu action.
type ContextMenuActionEvent struct {
	Menu  *ContextMenu
	Index int
	Item  ContextMenuItem
}

// TextInputChangedEvent describes a single-line input value change.
type TextInputChangedEvent struct {
	TextInput *TextInput
	Previous  string
	Text      string
}

// TextInputSubmittedEvent describes Enter/Return in a single-line input.
type TextInputSubmittedEvent struct {
	TextInput *TextInput
	Text      string
}

// TextAreaChangedEvent describes a multiline text area value change.
type TextAreaChangedEvent struct {
	TextArea *TextArea
	Previous string
	Text     string
}
