package components

import (
	"bytes"
	"errors"
	"math"

	"github.com/tacf/goak/layout"
)

const defaultUIFontSize = 20.0

var ErrInvalidFontSize = errors.New("goak: font size must be finite and greater than zero")

// UI owns one retained component tree and its separate context-menu overlays.
type UI struct {
	root         *layout.Container
	rootEl       *Root
	theme        Theme
	visible      bool
	interactive  bool
	fontSize     float64
	fontData     []byte
	fontRevision uint64

	components   []Component
	menus        []*ContextMenu
	componentSet map[Component]struct{}
	treeDirty    bool
	layoutDirty  bool
	layoutWidth  float64
	layoutHeight float64
}

func NewUI() *UI {
	root := layout.NewContainer(layout.AutoSize(), layout.AutoSize())
	ui := &UI{
		root:         root,
		theme:        DefaultTheme(),
		visible:      true,
		interactive:  true,
		fontSize:     defaultUIFontSize,
		treeDirty:    true,
		layoutDirty:  true,
		componentSet: make(map[Component]struct{}),
	}
	ui.rootEl = newRoot(ui, root)
	return ui
}

func (u *UI) Visible() bool { return u != nil && u.visible }

func (u *UI) SetVisible(visible bool) {
	if u == nil || u.visible == visible {
		return
	}
	u.visible = visible
	if !visible {
		u.rootEl.closePopups()
	}
}

func (u *UI) Interactive() bool { return u != nil && u.interactive }

func (u *UI) SetInteractive(interactive bool) {
	if u == nil || u.interactive == interactive {
		return
	}
	u.interactive = interactive
	if !interactive {
		u.rootEl.closePopups()
	}
}

func (u *UI) FontSize() float64 {
	if u == nil || u.fontSize <= 0 {
		return defaultUIFontSize
	}
	return u.fontSize
}

// SetFontSize changes the retained font size and reports invalid input.
func (u *UI) SetFontSize(size float64) error {
	if u == nil || size <= 0 || math.IsNaN(size) || math.IsInf(size, 0) {
		return ErrInvalidFontSize
	}
	if u.fontSize != size {
		u.fontSize = size
		u.fontRevision++
		u.invalidateLayout()
	}
	return nil
}

func (u *UI) FontData() []byte {
	if u == nil {
		return nil
	}
	return bytes.Clone(u.fontData)
}

func (u *UI) SetFontData(data []byte) {
	if u == nil {
		return
	}
	if len(data) == 0 {
		data = nil
	}
	if bytes.Equal(u.fontData, data) {
		return
	}
	u.fontData = bytes.Clone(data)
	u.fontRevision++
	u.invalidateLayout()
}

func (u *UI) FontRevision() uint64 {
	if u == nil {
		return 0
	}
	return u.fontRevision
}

func (u *UI) Root() *Root {
	if u == nil {
		return nil
	}
	return u.rootEl
}

func (u *UI) Theme() *Theme {
	if u == nil {
		return nil
	}
	return &u.theme
}

func (u *UI) SetTheme(theme Theme) {
	if u != nil {
		u.theme = theme
	}
}

// InvalidateLayout requests layout on the next host update. Call it after
// directly mutating an exposed layout.Container field.
func (u *UI) InvalidateLayout() {
	if u != nil {
		u.invalidateLayout()
	}
}

func (u *UI) invalidateLayout() { u.layoutDirty = true }

func (u *UI) invalidateTree() {
	u.treeDirty = true
	u.layoutDirty = true
}

// Layout resolves the retained tree only when its viewport or layout state
// changed. It returns whether a layout pass ran.
func (u *UI) Layout(width, height float64) bool {
	if u == nil {
		return false
	}
	if !u.layoutDirty && u.layoutWidth == width && u.layoutHeight == height {
		return false
	}
	layout.Layout(u.root, width, height)
	u.layoutWidth, u.layoutHeight = width, height
	u.layoutDirty = false
	return true
}

func (u *UI) rebuildComponents() {
	if u == nil || !u.treeDirty {
		return
	}
	u.components = u.components[:0]
	u.menus = u.menus[:0]
	clear(u.componentSet)
	var walk func(*Element)
	walk = func(parent *Element) {
		u.menus = append(u.menus, parent.menus...)
		for _, child := range sortedChildren(parent) {
			u.components = append(u.components, child)
			u.componentSet[child] = struct{}{}
			walk(child.componentElement())
		}
	}
	walk(&u.rootEl.Element)
	u.treeDirty = false
}

// ComponentCount returns the number of mounted components in draw order.
func (u *UI) ComponentCount() int {
	if u == nil {
		return 0
	}
	u.rebuildComponents()
	return len(u.components)
}

// Component returns the component at a draw-order index, or nil.
func (u *UI) Component(index int) Component {
	if u == nil {
		return nil
	}
	u.rebuildComponents()
	if index < 0 || index >= len(u.components) {
		return nil
	}
	return u.components[index]
}

// Components returns a safe snapshot in draw order.
func (u *UI) Components() []Component {
	if u == nil {
		return nil
	}
	u.rebuildComponents()
	return append([]Component(nil), u.components...)
}

// Contains reports whether a component is currently mounted in this UI.
func (u *UI) Contains(component Component) bool {
	if u == nil || component == nil {
		return false
	}
	u.rebuildComponents()
	_, ok := u.componentSet[component]
	return ok
}

// ComponentVisible reports effective visibility, including ancestors.
func (u *UI) ComponentVisible(component Component) bool {
	return u != nil && component != nil && u.Contains(component) &&
		component.componentElement().effectivelyVisible()
}

// ComponentEnabled reports effective input state, including ancestors.
func (u *UI) ComponentEnabled(component Component) bool {
	return u != nil && component != nil && u.Contains(component) &&
		component.componentElement().effectivelyEnabled()
}

func collectComponents[T Component](u *UI) []T {
	if u == nil {
		return nil
	}
	u.rebuildComponents()
	result := make([]T, 0)
	for _, component := range u.components {
		if typed, ok := component.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

// Typed accessors return copies and cannot mutate the UI's ownership state.
func (u *UI) Panels() []*Panel           { return collectComponents[*Panel](u) }
func (u *UI) Labels() []*Label           { return collectComponents[*Label](u) }
func (u *UI) Buttons() []*Button         { return collectComponents[*Button](u) }
func (u *UI) MenuBars() []*MenuBar       { return collectComponents[*MenuBar](u) }
func (u *UI) Checkboxes() []*Checkbox    { return collectComponents[*Checkbox](u) }
func (u *UI) RadioGroups() []*RadioGroup { return collectComponents[*RadioGroup](u) }
func (u *UI) Sliders() []*Slider         { return collectComponents[*Slider](u) }
func (u *UI) Dropdowns() []*Dropdown     { return collectComponents[*Dropdown](u) }
func (u *UI) Images() []*Image           { return collectComponents[*Image](u) }
func (u *UI) TextInputs() []*TextInput   { return collectComponents[*TextInput](u) }
func (u *UI) TextAreas() []*TextArea     { return collectComponents[*TextArea](u) }

func (u *UI) ButtonClicked(index int) {
	buttons := u.Buttons()
	if index >= 0 && index < len(buttons) {
		buttons[index].Click()
	}
}

func (u *UI) contextMenus() []*ContextMenu {
	if u == nil {
		return nil
	}
	u.rebuildComponents()
	return u.menus
}

func (u *UI) ContextMenus() []*ContextMenu {
	return append([]*ContextMenu(nil), u.contextMenus()...)
}

func (u *UI) ContextMenuCount() int { return len(u.contextMenus()) }

func (u *UI) ContextMenu(index int) *ContextMenu {
	menus := u.contextMenus()
	if index < 0 || index >= len(menus) {
		return nil
	}
	return menus[index]
}

func (u *UI) ContextMenuVisible(menu *ContextMenu) bool {
	return u != nil && menu != nil && menu.ui == u && menu.owner != nil &&
		menu.owner.effectivelyVisible()
}

func (u *UI) ContextMenuEnabled(menu *ContextMenu) bool {
	return u != nil && menu != nil && menu.ui == u && menu.owner != nil &&
		menu.owner.effectivelyEnabled()
}
