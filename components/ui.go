package components

import "github.com/tacf/goak/layout"

// UI holds the root and all panels/buttons for layout and drawing.
type UI struct {
	root         *layout.Container
	rootEl       *Root
	theme        Theme
	panels       []*Panel
	labels       []*Label
	buttons      []*Button
	menus        []*MenuBar
	checkboxes   []*Checkbox
	radioGroups  []*RadioGroup
	sliders      []*Slider
	dropdowns    []*Dropdown
	images       []*Image
	textInputs   []*TextInput
	textAreas    []*TextArea
	contextMenus []*ContextMenu
}

// NewUI creates a UI with an empty root. Use Root() to get the root element and build the tree.
func NewUI() *UI {
	root := layout.NewContainer(layout.AutoSize(), layout.AutoSize())
	u := &UI{
		root:    root,
		theme:   DefaultTheme(),
		panels:  nil,
		buttons: nil,
		menus:   nil,
	}
	u.rootEl = &Root{ui: u, c: root, Scale: 1}
	return u
}

// Root returns the root element. Build the tree with root.CreatePanel(...), then panel.CreateButton(...) etc.
// Root.Scale (default 1) scales the whole UI when changed.
func (u *UI) Root() *Root {
	return u.rootEl
}

// Theme returns the UI's live theme. Its fields can be changed at any time.
func (u *UI) Theme() *Theme {
	return &u.theme
}

// SetTheme replaces all theme values for this UI.
func (u *UI) SetTheme(theme Theme) {
	u.theme = theme
}

// Panels returns all panels (for rendering).
func (u *UI) Panels() []*Panel {
	return u.panels
}

// Labels returns all labels for rendering.
func (u *UI) Labels() []*Label {
	return u.labels
}

// Buttons returns all buttons (for rendering and hit-test).
func (u *UI) Buttons() []*Button {
	return u.buttons
}

// MenuBars returns all menu bars (for rendering and hit-test).
func (u *UI) MenuBars() []*MenuBar {
	return u.menus
}

// ButtonClicked activates the button at index.
func (u *UI) ButtonClicked(index int) {
	if index < 0 || index >= len(u.buttons) {
		return
	}
	u.buttons[index].Click()
}

// Checkboxes returns all checkboxes (for rendering and hit-test).
func (u *UI) Checkboxes() []*Checkbox {
	return u.checkboxes
}

// RadioGroups returns all radio groups (for rendering and hit-test).
func (u *UI) RadioGroups() []*RadioGroup {
	return u.radioGroups
}

// Sliders returns all sliders (for rendering and hit-test).
func (u *UI) Sliders() []*Slider {
	return u.sliders
}

// Dropdowns returns all dropdowns (for rendering and hit-test).
func (u *UI) Dropdowns() []*Dropdown {
	return u.dropdowns
}

// Images returns all image components for rendering.
func (u *UI) Images() []*Image {
	return u.images
}

// TextInputs returns all single-line text inputs.
func (u *UI) TextInputs() []*TextInput {
	return u.textInputs
}

// TextAreas returns all multiline text areas.
func (u *UI) TextAreas() []*TextArea {
	return u.textAreas
}

// ContextMenus returns all context menus (for rendering and hit-test).
func (u *UI) ContextMenus() []*ContextMenu {
	return u.contextMenus
}
