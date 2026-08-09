package components

import (
	"errors"
	"image"
	"math"
	"reflect"
	"sort"

	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/layout"
	"github.com/tacf/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

var (
	// ErrNilComponent is returned when a nil component is added to a tree.
	ErrNilComponent = errors.New("goak: component is nil")
	// ErrComponentAttached is returned when a component already belongs to a tree.
	ErrComponentAttached = errors.New("goak: component is already attached")
	// ErrComponentCycle is returned when an attachment would create a cycle.
	ErrComponentCycle = errors.New("goak: component attachment would create a cycle")
	// ErrInvalidScale is returned for non-positive or non-finite scale values.
	ErrInvalidScale = errors.New("goak: scale must be finite and greater than zero")
	// ErrInvalidLength is returned for non-finite or out-of-range geometry.
	ErrInvalidLength = errors.New("goak: length is outside the supported range")
)

// Component is a retained node that can be mounted below a Root or Panel.
// Built-in components retain their layout position while hidden.
type Component interface {
	Container() *layout.Container
	Visible() bool
	SetVisible(bool)
	Enabled() bool
	SetEnabled(bool)
	ZIndex() int
	SetZIndex(int)
	Parent() Component
	Remove() bool
	componentElement() *Element
}

// Element supplies ownership, visibility, enabled-state, and z-order behavior
// to built-in retained components. It is embedded by every component type.
type Element struct {
	owner    Component
	parent   *Element
	children []Component
	menus    []*ContextMenu
	ui       *UI
	hidden   bool
	disabled bool
	zIndex   int
}

func (e *Element) init(owner Component) {
	e.owner = owner
}

func (e *Element) componentElement() *Element { return e }

// Visible reports the component's local visibility. Hidden ancestors can
// still make an otherwise visible component effectively hidden.
func (e *Element) Visible() bool { return e != nil && !e.hidden }

// SetVisible controls drawing and input without removing layout space.
func (e *Element) SetVisible(visible bool) {
	if e == nil || e.hidden == !visible {
		return
	}
	e.hidden = !visible
	if !visible {
		e.closePopups()
	}
}

// Enabled reports the component's local enabled state. Disabled ancestors
// make all descendants effectively disabled.
func (e *Element) Enabled() bool { return e != nil && !e.disabled }

// SetEnabled enables or disables retained input for the component subtree.
func (e *Element) SetEnabled(enabled bool) {
	if e == nil || e.disabled == !enabled {
		return
	}
	e.disabled = !enabled
	if !enabled {
		e.closePopups()
	}
}

// ZIndex returns the sibling draw-order override. Higher values draw and hit
// test above lower values; equal values retain insertion order.
func (e *Element) ZIndex() int {
	if e == nil {
		return 0
	}
	return e.zIndex
}

// SetZIndex changes sibling draw and hit-test ordering without changing layout.
func (e *Element) SetZIndex(z int) {
	if e == nil || e.zIndex == z {
		return
	}
	e.zIndex = z
	if e.ui != nil {
		e.ui.invalidateTree()
	}
}

// Parent returns the retained parent, or nil for a detached component.
func (e *Element) Parent() Component {
	if e == nil || e.parent == nil {
		return nil
	}
	return e.parent.owner
}

// Remove detaches a component subtree. It returns false for detached nodes and
// the root. A removed subtree can be attached elsewhere afterward.
func (e *Element) Remove() bool {
	if e == nil || e.parent == nil || e.owner == nil {
		return false
	}
	parent := e.parent
	ui := e.ui
	removeComponent(&parent.children, e.owner)
	removeContainer(&parent.owner.Container().Children, e.owner.Container())
	e.parent = nil
	if ui != nil {
		releaseTree(e)
		setTreeUI(e, nil)
		ui.invalidateTree()
		ui.invalidateLayout()
	}
	return true
}

// SetSize updates both layout dimensions and invalidates an attached UI.
func (e *Element) SetSize(width, height layout.Size) {
	if e == nil || e.owner == nil {
		return
	}
	c := e.owner.Container()
	c.Width, c.Height = width, height
	if e.ui != nil {
		e.ui.invalidateLayout()
	}
}

func (e *Element) effectivelyVisible() bool {
	for current := e; current != nil; current = current.parent {
		if current.hidden {
			return false
		}
	}
	return true
}

func (e *Element) effectivelyEnabled() bool {
	for current := e; current != nil; current = current.parent {
		if current.disabled || current.hidden {
			return false
		}
	}
	return true
}

func (e *Element) closePopups() {
	for _, menu := range e.menus {
		menu.Close()
	}
	for _, child := range e.children {
		child.componentElement().closePopups()
		switch component := child.(type) {
		case *Dropdown:
			component.Close()
		case *MenuBar:
			component.Close()
		}
	}
}

func attachComponent(parent *Element, child Component) error {
	if parent == nil || parent.owner == nil || isNilComponent(child) {
		return ErrNilComponent
	}
	state := child.componentElement()
	if state == nil || state.owner == nil || child.Container() == nil {
		return ErrNilComponent
	}
	if state.parent != nil || state.ui != nil {
		return ErrComponentAttached
	}
	for current := parent; current != nil; current = current.parent {
		if current == state {
			return ErrComponentCycle
		}
	}
	state.parent = parent
	parent.children = append(parent.children, child)
	parent.owner.Container().Children = append(parent.owner.Container().Children, child.Container())
	setTreeUI(state, parent.ui)
	if parent.ui != nil {
		parent.ui.invalidateTree()
		parent.ui.invalidateLayout()
	}
	return nil
}

func isNilComponent(component Component) bool {
	if component == nil {
		return true
	}
	value := reflect.ValueOf(component)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func attachContextMenu(parent *Element, menu *ContextMenu) error {
	if parent == nil || menu == nil {
		return ErrNilComponent
	}
	if menu.owner != nil || menu.ui != nil {
		return ErrComponentAttached
	}
	menu.owner = parent
	menu.ui = parent.ui
	parent.menus = append(parent.menus, menu)
	if parent.ui != nil {
		parent.ui.invalidateTree()
	}
	return nil
}

func setTreeUI(element *Element, ui *UI) {
	if element == nil {
		return
	}
	element.ui = ui
	for _, menu := range element.menus {
		menu.ui = ui
	}
	for _, child := range element.children {
		setTreeUI(child.componentElement(), ui)
	}
}

func releaseTree(element *Element) {
	if element == nil {
		return
	}
	element.closePopups()
	if imageComponent, ok := element.owner.(*Image); ok {
		imageComponent.Close()
	}
	if slider, ok := element.owner.(*Slider); ok {
		slider.StopDrag()
	}
	for _, child := range element.children {
		releaseTree(child.componentElement())
	}
}

func removeComponent(components *[]Component, target Component) {
	values := *components
	for index, component := range values {
		if component == target {
			copy(values[index:], values[index+1:])
			*components = values[:len(values)-1]
			return
		}
	}
}

func removeContainer(containers *[]*layout.Container, target *layout.Container) {
	values := *containers
	for index, container := range values {
		if container == target {
			copy(values[index:], values[index+1:])
			*containers = values[:len(values)-1]
			return
		}
	}
}

func sortedChildren(element *Element) []Component {
	if len(element.children) < 2 {
		return element.children
	}
	children := append([]Component(nil), element.children...)
	sort.SliceStable(children, func(i, j int) bool {
		return children[i].ZIndex() < children[j].ZIndex()
	})
	return children
}

// Root is the retained tree root. Use UI.Root to obtain it.
type Root struct {
	Element
	c     *layout.Container
	scale float64
}

func newRoot(ui *UI, container *layout.Container) *Root {
	root := &Root{c: container, scale: 1}
	root.Element.init(root)
	root.Element.ui = ui
	return root
}

// Container returns the root layout container.
func (r *Root) Container() *layout.Container { return r.c }

// Scale returns the logical-to-renderer UI scale multiplier.
func (r *Root) Scale() float64 {
	if r == nil || r.scale <= 0 {
		return 1
	}
	return r.scale
}

// SetScale sets the content scale and rejects non-positive or non-finite values.
func (r *Root) SetScale(scale float64) error {
	if r == nil || scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		return ErrInvalidScale
	}
	if r.scale != scale {
		r.scale = scale
		if r.ui != nil {
			r.ui.invalidateLayout()
		}
	}
	return nil
}

// Add mounts a retained component as a direct root child.
func (r *Root) Add(component Component) error { return attachComponent(&r.Element, component) }

func (r *Root) SetAlignment(horizontal, vertical layout.Alignment) {
	r.c.HorizontalAlign = normalizedAlignment(horizontal)
	r.c.VerticalAlign = normalizedAlignment(vertical)
	r.ui.invalidateLayout()
}

func (r *Root) SetDirection(direction layout.Direction) {
	r.c.Direction = normalizedDirection(direction)
	r.ui.invalidateLayout()
}

func (r *Root) SetGap(gap float64) {
	r.c.Gap = normalizedLength(gap)
	r.ui.invalidateLayout()
}

func (r *Root) SetPadding(padding float64) {
	r.c.Padding = normalizedLength(padding)
	r.c.Insets = layout.Insets{}
	r.ui.invalidateLayout()
}

func (r *Root) SetInsets(insets layout.Insets) {
	r.c.Padding = 0
	r.c.Insets = normalizedInsets(insets)
	r.ui.invalidateLayout()
}

func (r *Root) CreatePanel(width, height layout.Size) *Panel {
	panel := NewPanel(width, height)
	_ = r.Add(panel)
	return panel
}

func (r *Root) AddPanel(panel *Panel) error { return r.Add(panel) }

func (r *Root) CreateLabel(width, height layout.Size, text string) *Label {
	label := NewLabel(width, height, text)
	_ = r.Add(label)
	return label
}

func (r *Root) AddLabel(label *Label) error { return r.Add(label) }

func (r *Root) CreateImage(width, height layout.Size, source image.Image) *Image {
	component := NewImage(width, height, source)
	_ = r.Add(component)
	return component
}

func (r *Root) AddImage(component *Image) error { return r.Add(component) }

func (r *Root) CreateTextInput(width, height layout.Size, text string) *TextInput {
	input := NewTextInput(width, height, text)
	_ = r.Add(input)
	return input
}

func (r *Root) AddTextInput(input *TextInput) error { return r.Add(input) }

func (r *Root) CreateTextArea(width, height layout.Size, text string) *TextArea {
	area := NewTextArea(width, height, text)
	_ = r.Add(area)
	return area
}

func (r *Root) AddTextArea(area *TextArea) error { return r.Add(area) }

func (r *Root) CreateMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	menu := NewMenuBar(height, widthMode)
	_ = r.Add(menu)
	return menu
}

func (r *Root) AddMenuBar(menu *MenuBar) error { return r.Add(menu) }

func (r *Root) AddContextMenu(menu *ContextMenu) error {
	return attachContextMenu(&r.Element, menu)
}

// Panel is a retained container that can be composed while detached and later
// mounted as a complete subtree.
type Panel struct {
	Element
	c          *layout.Container
	background *colors.Color
}

func NewPanel(width, height layout.Size) *Panel {
	panel := &Panel{c: layout.NewContainer(width, height)}
	panel.Element.init(panel)
	return panel
}

func (p *Panel) Container() *layout.Container { return p.c }
func (p *Panel) Bounds() layout.Rect          { return p.c.Bounds }

// Add mounts a component below the panel. It is safe for detached panels.
func (p *Panel) Add(component Component) error { return attachComponent(&p.Element, component) }

func (p *Panel) SetAlignment(horizontal, vertical layout.Alignment) {
	p.c.HorizontalAlign = normalizedAlignment(horizontal)
	p.c.VerticalAlign = normalizedAlignment(vertical)
	p.invalidateLayout()
}

func (p *Panel) SetDirection(direction layout.Direction) {
	p.c.Direction = normalizedDirection(direction)
	p.invalidateLayout()
}

func (p *Panel) SetGap(gap float64) {
	p.c.Gap = normalizedLength(gap)
	p.invalidateLayout()
}

func (p *Panel) SetPadding(padding float64) {
	p.c.Padding = normalizedLength(padding)
	p.c.Insets = layout.Insets{}
	p.invalidateLayout()
}

func (p *Panel) SetInsets(insets layout.Insets) {
	p.c.Padding = 0
	p.c.Insets = normalizedInsets(insets)
	p.invalidateLayout()
}

func (p *Panel) invalidateLayout() {
	if p.ui != nil {
		p.ui.invalidateLayout()
	}
}

func normalizedDirection(direction layout.Direction) layout.Direction {
	if direction == layout.Row {
		return layout.Row
	}
	return layout.Column
}

func normalizedAlignment(alignment layout.Alignment) layout.Alignment {
	switch alignment {
	case layout.AlignCenter, layout.AlignEnd:
		return alignment
	default:
		return layout.AlignStart
	}
}

func normalizedLength(value float64) float64 {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func normalizedInsets(insets layout.Insets) layout.Insets {
	return layout.Insets{
		Top:    normalizedLength(insets.Top),
		Right:  normalizedLength(insets.Right),
		Bottom: normalizedLength(insets.Bottom),
		Left:   normalizedLength(insets.Left),
	}
}

func (p *Panel) SetBackground(color colors.Color) { p.background = &color }

func (p *Panel) ClearBackground() { p.background = nil }

func (p *Panel) Background() (colors.Color, bool) {
	if p == nil || p.background == nil {
		return colors.Color{}, false
	}
	return *p.background, true
}

func (p *Panel) SetBackgroundHex(hex string) bool {
	color, ok := colors.ParseHex(hex)
	if !ok {
		return false
	}
	p.background = &color
	return true
}

func (p *Panel) CreatePanel(width, height layout.Size) *Panel {
	child := NewPanel(width, height)
	_ = p.Add(child)
	return child
}

func (p *Panel) AddPanel(child *Panel) error { return p.Add(child) }

func (p *Panel) CreateButton(width, height layout.Size, label string) *Button {
	button := NewButton(width, height, label)
	_ = p.Add(button)
	return button
}

func (p *Panel) AddButton(button *Button) error { return p.Add(button) }

func (p *Panel) CreateLabel(width, height layout.Size, text string) *Label {
	label := NewLabel(width, height, text)
	_ = p.Add(label)
	return label
}

func (p *Panel) AddLabel(label *Label) error { return p.Add(label) }

func (p *Panel) CreateImage(width, height layout.Size, source image.Image) *Image {
	component := NewImage(width, height, source)
	_ = p.Add(component)
	return component
}

func (p *Panel) AddImage(component *Image) error { return p.Add(component) }

func (p *Panel) CreateTextInput(width, height layout.Size, text string) *TextInput {
	input := NewTextInput(width, height, text)
	_ = p.Add(input)
	return input
}

func (p *Panel) AddTextInput(input *TextInput) error { return p.Add(input) }

func (p *Panel) CreateTextArea(width, height layout.Size, text string) *TextArea {
	area := NewTextArea(width, height, text)
	_ = p.Add(area)
	return area
}

func (p *Panel) AddTextArea(area *TextArea) error { return p.Add(area) }

func (p *Panel) CreateMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	menu := NewMenuBar(height, widthMode)
	_ = p.Add(menu)
	return menu
}

func (p *Panel) AddMenuBar(menu *MenuBar) error { return p.Add(menu) }

func (p *Panel) CreateCheckbox(width, height layout.Size, label string) *Checkbox {
	checkbox := NewCheckbox(width, height, label)
	_ = p.Add(checkbox)
	return checkbox
}

func (p *Panel) AddCheckbox(checkbox *Checkbox) error { return p.Add(checkbox) }

func (p *Panel) CreateRadioGroup(width, height layout.Size, options []RadioOption) *RadioGroup {
	group := NewRadioGroup(width, height, options)
	_ = p.Add(group)
	return group
}

func (p *Panel) AddRadioGroup(group *RadioGroup) error { return p.Add(group) }

func (p *Panel) CreateSlider(width, height layout.Size, label string, min, max, initial float64) *Slider {
	slider := NewSlider(width, height, label, min, max, initial)
	_ = p.Add(slider)
	return slider
}

func (p *Panel) AddSlider(slider *Slider) error { return p.Add(slider) }

func (p *Panel) CreateDropdown(width, height layout.Size, label string, options []DropdownOption) *Dropdown {
	dropdown := NewDropdown(width, height, label, options)
	_ = p.Add(dropdown)
	return dropdown
}

func (p *Panel) AddDropdown(dropdown *Dropdown) error { return p.Add(dropdown) }

func (p *Panel) AddContextMenu(menu *ContextMenu) error {
	return attachContextMenu(&p.Element, menu)
}

func (p *Panel) Draw(renderer *sdl.Renderer, theme PanelTheme) {
	bounds := p.Bounds()
	fill := theme.DefaultFill
	if p.background != nil {
		fill = *p.background
	}
	rendering.FillRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, fill)
	rendering.DrawStrokeRect(renderer, bounds.X, bounds.Y, bounds.W, bounds.H, 1, theme.Stroke)
}
