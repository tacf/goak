package components

import "goak/internal/goak/layout"

// UI holds the root and all panels/buttons for layout and drawing.
type UI struct {
	root         *layout.Container
	rootEl       *Root
	panels       []*Panel
	buttons      []*Button
	menus        []*MenuBar
	checkboxes   []*Checkbox
	radioGroups  []*RadioGroup
	sliders      []*Slider
	dropdowns    []*Dropdown
	contextMenus []*ContextMenu
}

// NewUI creates a UI with an empty root. Use Root() to get the root element and build the tree.
func NewUI() *UI {
	root := layout.NewContainer(layout.AutoSize(), layout.AutoSize())
	u := &UI{
		root:    root,
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

// Panels returns all panels (for rendering).
func (u *UI) Panels() []*Panel {
	return u.panels
}

// Buttons returns all buttons (for rendering and hit-test).
func (u *UI) Buttons() []*Button {
	return u.buttons
}

// MenuBars returns all menu bars (for rendering and hit-test).
func (u *UI) MenuBars() []*MenuBar {
	return u.menus
}

// ButtonClicked runs the OnClick callback for the button at index. No-op if index is out of range or OnClick is nil.
func (u *UI) ButtonClicked(index int) {
	if index < 0 || index >= len(u.buttons) {
		return
	}
	if u.buttons[index].OnClick != nil {
		u.buttons[index].OnClick()
	}
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

// ContextMenus returns all context menus (for rendering and hit-test).
func (u *UI) ContextMenus() []*ContextMenu {
	return u.contextMenus
}
