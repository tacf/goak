package components

import (
	"image"

	"github.com/tacf/goak/internal/goak/colors"
	"github.com/tacf/goak/internal/goak/layout"
	"github.com/tacf/goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Root is the root element. Use ui.Root() to get it, then root.CreatePanel(...) or root.AddPanel(panel) to build the tree.
// Scale is the content scale (1 = 1:1). Change it to scale the whole UI (e.g. 2 = 2x bigger).
type Root struct {
	ui    *UI
	c     *layout.Container
	Scale float64 // default 1
}

// Container returns the underlying layout container (for layout.Layout). Internal use.
func (r *Root) Container() *layout.Container { return r.c }

// SetAlignment sets how direct children are positioned inside the root.
func (r *Root) SetAlignment(horizontal, vertical layout.Alignment) {
	r.c.HorizontalAlign = horizontal
	r.c.VerticalAlign = vertical
}

// SetPadding sets uniform padding between the root bounds and its children.
func (r *Root) SetPadding(padding float64) {
	r.c.Padding = max(0, padding)
}

// CreatePanel creates a new panel and adds it as a direct child of the root. Returns the panel.
func (r *Root) CreatePanel(width, height layout.Size) *Panel {
	p := NewPanel(width, height)
	r.AddPanel(p)
	return p
}

// AddPanel adds an existing panel (e.g. from NewPanel) as a direct child of the root. Reusable panels.
func (r *Root) AddPanel(p *Panel) {
	p.ui = r.ui
	r.c.Children = append(r.c.Children, p.c)
	r.ui.panels = append(r.ui.panels, p)
}

// CreateLabel creates a label and adds it as a direct child of the root.
func (r *Root) CreateLabel(width, height layout.Size, text string) *Label {
	label := NewLabel(width, height, text)
	r.AddLabel(label)
	return label
}

// AddLabel adds an existing label as a direct child of the root.
func (r *Root) AddLabel(label *Label) {
	r.c.Children = append(r.c.Children, label.Container())
	r.ui.labels = append(r.ui.labels, label)
}

// CreateImage creates an image component and adds it as a direct child.
func (r *Root) CreateImage(width, height layout.Size, source image.Image) *Image {
	component := NewImage(width, height, source)
	r.AddImage(component)
	return component
}

// AddImage adds an existing image component as a direct child.
func (r *Root) AddImage(component *Image) {
	r.c.Children = append(r.c.Children, component.Container())
	r.ui.images = append(r.ui.images, component)
}

// CreateTextInput creates a single-line input and adds it as a direct child.
func (r *Root) CreateTextInput(width, height layout.Size, text string) *TextInput {
	input := NewTextInput(width, height, text)
	r.AddTextInput(input)
	return input
}

// AddTextInput adds an existing single-line input as a direct child.
func (r *Root) AddTextInput(input *TextInput) {
	r.c.Children = append(r.c.Children, input.Container())
	r.ui.textInputs = append(r.ui.textInputs, input)
}

// CreateTextArea creates a multiline text area and adds it as a direct child.
func (r *Root) CreateTextArea(width, height layout.Size, text string) *TextArea {
	area := NewTextArea(width, height, text)
	r.AddTextArea(area)
	return area
}

// AddTextArea adds an existing multiline text area as a direct child.
func (r *Root) AddTextArea(area *TextArea) {
	r.c.Children = append(r.c.Children, area.Container())
	r.ui.textAreas = append(r.ui.textAreas, area)
}

// CreateMenuBar creates a new menu bar and adds it as a direct child of the root.
func (r *Root) CreateMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	m := NewMenuBar(height, widthMode)
	r.AddMenuBar(m)
	return m
}

// AddMenuBar adds an existing menu bar as a direct child of the root.
func (r *Root) AddMenuBar(m *MenuBar) {
	m.ui = r.ui
	r.c.Children = append(r.c.Children, m.c)
	r.ui.menus = append(r.ui.menus, m)
}

// Panel is a container that draws a background and can contain more panels or buttons.
// Background is optional; if nil the renderer uses its default.
// Create with NewPanel for reuse, or use CreatePanel to create and add in one step.
type Panel struct {
	ui         *UI
	c          *layout.Container
	Background *colors.Color
}

// NewPanel creates a standalone panel (not in the tree). Add it with root.AddPanel(panel) or parent.AddPanel(panel).
func NewPanel(width, height layout.Size) *Panel {
	return &Panel{c: layout.NewContainer(width, height)}
}

// Container returns the layout node for this panel (internal use).
func (p *Panel) Container() *layout.Container { return p.c }

// SetAlignment sets how direct children are positioned inside this panel.
func (p *Panel) SetAlignment(horizontal, vertical layout.Alignment) {
	p.c.HorizontalAlign = horizontal
	p.c.VerticalAlign = vertical
}

// SetPadding sets uniform padding between the panel bounds and its children.
func (p *Panel) SetPadding(padding float64) {
	p.c.Padding = max(0, padding)
}

// SetBackground sets panel background color.
func (p *Panel) SetBackground(c colors.Color) {
	p.Background = &c
}

// SetBackgroundHex parses and sets panel background from #RGB/#RRGGBB.
// Returns false if the hex value is invalid.
func (p *Panel) SetBackgroundHex(hex string) bool {
	c, ok := colors.ParseHex(hex)
	if !ok {
		return false
	}
	p.Background = &c
	return true
}

// Bounds returns the computed layout rect after Layout.
func (p *Panel) Bounds() layout.Rect { return p.c.Bounds }

// CreatePanel creates a new child panel and adds it. Returns the panel.
func (p *Panel) CreatePanel(width, height layout.Size) *Panel {
	child := NewPanel(width, height)
	p.AddPanel(child)
	return child
}

// AddPanel adds an existing panel (e.g. from NewPanel) as a child. Reusable panels.
func (p *Panel) AddPanel(child *Panel) {
	child.ui = p.ui
	p.c.Children = append(p.c.Children, child.c)
	p.ui.panels = append(p.ui.panels, child)
}

// CreateButton creates a new button and adds it to this panel. Returns the button.
func (p *Panel) CreateButton(width, height layout.Size, label string) *Button {
	b := NewButton(width, height, label)
	p.AddButton(b)
	return b
}

// AddButton adds an existing button (e.g. from NewButton) to this panel.
func (p *Panel) AddButton(b *Button) {
	p.c.Children = append(p.c.Children, b.c)
	p.ui.buttons = append(p.ui.buttons, b)
}

// CreateLabel creates a label and adds it to this panel.
func (p *Panel) CreateLabel(width, height layout.Size, text string) *Label {
	label := NewLabel(width, height, text)
	p.AddLabel(label)
	return label
}

// AddLabel adds an existing label to this panel.
func (p *Panel) AddLabel(label *Label) {
	p.c.Children = append(p.c.Children, label.Container())
	p.ui.labels = append(p.ui.labels, label)
}

// CreateImage creates an image component and adds it to this panel.
func (p *Panel) CreateImage(width, height layout.Size, source image.Image) *Image {
	component := NewImage(width, height, source)
	p.AddImage(component)
	return component
}

// AddImage adds an existing image component to this panel.
func (p *Panel) AddImage(component *Image) {
	p.c.Children = append(p.c.Children, component.Container())
	p.ui.images = append(p.ui.images, component)
}

// CreateTextInput creates a single-line input and adds it to this panel.
func (p *Panel) CreateTextInput(width, height layout.Size, text string) *TextInput {
	input := NewTextInput(width, height, text)
	p.AddTextInput(input)
	return input
}

// AddTextInput adds an existing single-line input to this panel.
func (p *Panel) AddTextInput(input *TextInput) {
	p.c.Children = append(p.c.Children, input.Container())
	p.ui.textInputs = append(p.ui.textInputs, input)
}

// CreateTextArea creates a multiline text area and adds it to this panel.
func (p *Panel) CreateTextArea(width, height layout.Size, text string) *TextArea {
	area := NewTextArea(width, height, text)
	p.AddTextArea(area)
	return area
}

// AddTextArea adds an existing multiline text area to this panel.
func (p *Panel) AddTextArea(area *TextArea) {
	p.c.Children = append(p.c.Children, area.Container())
	p.ui.textAreas = append(p.ui.textAreas, area)
}

// CreateMenuBar creates a new menu bar and adds it to this panel.
func (p *Panel) CreateMenuBar(height layout.Size, widthMode MenuBarWidthMode) *MenuBar {
	m := NewMenuBar(height, widthMode)
	p.AddMenuBar(m)
	return m
}

// AddMenuBar adds an existing menu bar to this panel.
func (p *Panel) AddMenuBar(m *MenuBar) {
	m.ui = p.ui
	p.c.Children = append(p.c.Children, m.c)
	p.ui.menus = append(p.ui.menus, m)
}

// CreateCheckbox creates a new checkbox and adds it to this panel. Returns the checkbox.
func (p *Panel) CreateCheckbox(width, height layout.Size, label string) *Checkbox {
	cb := NewCheckbox(width, height, label)
	p.AddCheckbox(cb)
	return cb
}

// AddCheckbox adds an existing checkbox to this panel.
func (p *Panel) AddCheckbox(cb *Checkbox) {
	p.c.Children = append(p.c.Children, cb.Container())
	p.ui.checkboxes = append(p.ui.checkboxes, cb)
}

// CreateRadioGroup creates a new radio group and adds it to this panel. Returns the radio group.
func (p *Panel) CreateRadioGroup(width, height layout.Size, options []RadioOption) *RadioGroup {
	rg := NewRadioGroup(width, height, options)
	p.AddRadioGroup(rg)
	return rg
}

// AddRadioGroup adds an existing radio group to this panel.
func (p *Panel) AddRadioGroup(rg *RadioGroup) {
	p.c.Children = append(p.c.Children, rg.Container())
	p.ui.radioGroups = append(p.ui.radioGroups, rg)
}

// CreateSlider creates a new slider and adds it to this panel. Returns the slider.
func (p *Panel) CreateSlider(width, height layout.Size, label string, min, max, initial float64) *Slider {
	s := NewSlider(width, height, label, min, max, initial)
	p.AddSlider(s)
	return s
}

// AddSlider adds an existing slider to this panel.
func (p *Panel) AddSlider(s *Slider) {
	p.c.Children = append(p.c.Children, s.Container())
	p.ui.sliders = append(p.ui.sliders, s)
}

// CreateDropdown creates a new dropdown and adds it to this panel. Returns the dropdown.
func (p *Panel) CreateDropdown(width, height layout.Size, label string, options []DropdownOption) *Dropdown {
	dd := NewDropdown(width, height, label, options)
	p.AddDropdown(dd)
	return dd
}

// AddDropdown adds an existing dropdown to this panel.
func (p *Panel) AddDropdown(dd *Dropdown) {
	p.c.Children = append(p.c.Children, dd.Container())
	p.ui.dropdowns = append(p.ui.dropdowns, dd)
}

// AddContextMenu adds a context menu to this panel (not part of layout tree).
func (p *Panel) AddContextMenu(cm *ContextMenu) {
	p.ui.contextMenus = append(p.ui.contextMenus, cm)
}

func (p *Panel) Draw(renderer *sdl.Renderer, theme PanelTheme) {
	b := p.Bounds()
	fill := theme.DefaultFill
	if p.Background != nil {
		fill = *p.Background
	}
	rendering.FillRect(renderer, b.X, b.Y, b.W, b.H, fill)
	rendering.DrawStrokeRect(renderer, b.X, b.Y, b.W, b.H, 1.0, theme.Stroke)
}
