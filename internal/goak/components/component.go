package components

import (
	"goak/internal/goak/colors"
	"goak/internal/goak/layout"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

// AddButton adds an existing button (e.g. from NewButton) to this panel. Reuse same style, set OnClick per instance.
func (p *Panel) AddButton(b *Button) {
	p.c.Children = append(p.c.Children, b.c)
	p.ui.buttons = append(p.ui.buttons, b)
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

// PanelTheme controls panel drawing colors.
type PanelTheme struct {
	DefaultFill colors.Color
	Stroke      colors.Color
}

// DefaultPanelTheme returns the default panel theme.
func DefaultPanelTheme() PanelTheme {
	return PanelTheme{
		DefaultFill: colors.HexOr("#2d2d2d", colors.RGB(45, 45, 45)),
		Stroke:      colors.HexOr("#555", colors.RGB(85, 85, 85)),
	}
}

// Draw renders the panel rectangle with fill and border.
func (p *Panel) Draw(dst *ebiten.Image, theme PanelTheme) {
	b := p.Bounds()
	fill := theme.DefaultFill
	if p.Background != nil {
		fill = *p.Background
	}
	vector.FillRect(dst, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), fill, true)
	drawPanelStrokeRect(dst, b.X, b.Y, b.W, b.H, theme.Stroke)
}

func drawPanelStrokeRect(dst *ebiten.Image, x, y, w, h float64, c colors.Color) {
	const t = 1.0
	ebitenutil.DrawRect(dst, x, y, w, t, c)
	ebitenutil.DrawRect(dst, x, y+h-t, w, t, c)
	ebitenutil.DrawRect(dst, x, y, t, h, c)
	ebitenutil.DrawRect(dst, x+w-t, y, t, h, c)
}
