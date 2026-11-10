// Command docking shows a workspace with draggable docked panels.
package main

import (
	"flag"
	"log"

	"github.com/tacf/goak"
	"github.com/tacf/goak/backend/sdl3"
	"github.com/tacf/goak/colors"
)

type workspace struct {
	initialized     bool
	showConsole     bool
	light           bool
	keepCenterEmpty bool
	passThrough     bool
	selected        int
	notes           string
}

func main() {
	renderer := flag.String("renderer", "software", "renderer driver: software, gpu, auto or an SDL driver name")
	flag.Parse()
	b, err := sdl3.New(sdl3.Config{Title: "goak docking", Width: 1000, Height: 720, Renderer: *renderer, Resizable: true})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	ctx, err := goak.NewContext(b, goak.Options{FontSize: 18})
	if err != nil {
		log.Fatal(err)
	}
	s := workspace{showConsole: true, passThrough: true, notes: "An empty workspace with floating windows.\n\nDrag a panel title to detach it.\nRelease away from docking targets to stay floating.\nUse the center target to fill an empty region.\nDrag the dividers to resize panels.\nHold Shift to disable docking, or Escape to cancel."}
	var in goak.Input
	for b.NewFrame(&in) {
		ctx.NewFrame(&in, b.DisplaySize())
		s.build(ctx)
		if err := ctx.Present(colors.RGB(24, 24, 28)); err != nil {
			log.Fatal(err)
		}
	}
}

func (s *workspace) build(c *goak.Context) {
	redockConsole := false
	if c.BeginMainMenuBar() {
		if c.BeginMenu("View") {
			c.MenuItemCheck("Keep center empty", "", &s.keepCenterEmpty)
			c.MenuItemCheck("Pass through empty center", "", &s.passThrough)
			if c.MenuItemCheck("Console", "", &s.showConsole) && s.showConsole {
				redockConsole = true
			}
			if c.MenuItemCheck("Light theme", "", &s.light) {
				if s.light {
					c.Style = goak.LightStyle()
				} else {
					c.Style = goak.DarkStyle()
				}
			}
			c.EndMenu()
		}
	}
	c.EndMainMenuBar()

	var flags goak.DockSpaceFlags
	if s.keepCenterEmpty {
		flags |= goak.DockNoDockingInCentralNode
	}
	if s.passThrough {
		flags |= goak.DockPassthruCentralNode
	}
	root := c.DockSpaceOverViewport("Workspace", flags)
	if !s.initialized {
		files, main := c.SplitDockNode(root, goak.DockLeft, 0.24)
		console, _ := c.SplitDockNode(main, goak.DockBottom, 0.28)
		c.DockWindow("Files", files)
		c.DockWindow("Console", console)
		s.initialized = true
	}

	if redockConsole {
		center := c.DockSpaceCentralNode(root)
		console, _ := c.SplitDockNode(center, goak.DockBottom, 0.28)
		c.DockWindow("Console", console)
	}

	c.SetNextWindowSize(goak.Vec2{X: 260, Y: 320}, goak.CondOnce)
	if c.Begin("Files", nil, 0) {
		for i, name := range []string{"README.md", "main.go", "go.mod", "docking.go"} {
			if c.Selectable(name, s.selected == i) {
				s.selected = i
			}
		}
		c.Separator()
		c.TextWrapped("Drag titles onto docking targets to rearrange panels. Release elsewhere to keep a window floating.")
	}
	c.End()

	c.SetNextWindowPos(goak.Vec2{X: c.DisplaySize.X * 0.3, Y: c.DisplaySize.Y * 0.14}, goak.CondOnce)
	c.SetNextWindowSize(goak.Vec2{X: 500, Y: 320}, goak.CondOnce)
	if c.Begin("Editor", nil, 0) {
		c.Text("Scratch pad")
		c.InputTextMultiline("##notes", &s.notes, c.ContentRegionAvail(), 0)
	}
	c.End()

	c.SetNextWindowSize(goak.Vec2{X: 450, Y: 200}, goak.CondOnce)
	if c.Begin("Console", &s.showConsole, 0) {
		if c.BeginChild("log", goak.Vec2{}, 0) {
			c.Text("Workspace ready.")
			c.Text("Editor starts floating over the empty center.")
			c.Text("Close this panel and reopen it from View.")
			c.Textf("Selected file: %d", s.selected+1)
		}
		c.EndChild()
	}
	c.End()
}
