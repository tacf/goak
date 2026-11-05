// Command demo shows the immediate-mode widget set.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tacf/goak"
	"github.com/tacf/goak/backend/sdl3"
	"github.com/tacf/goak/colors"
	"github.com/tacf/goak/internal/draw"
)

// gradient makes an RGBA test texture.
func gradient(n int) []byte {
	px := make([]byte, n*n*4)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			o := (y*n + x) * 4
			px[o], px[o+1], px[o+2], px[o+3] = byte(x*255/(n-1)), byte(y*255/(n-1)), 160, 255
		}
	}
	return px
}

type state struct {
	clicks   int
	featA    bool
	featB    bool
	featC    bool
	radio    int
	volume   float32
	steps    int
	drag     float32
	progress float32
	color    int
	name     string
	password string
	notes    string
	list     int
	showTab2 bool
	light    bool
	lastMenu string
	imageTex draw.TextureID
}

func main() {
	renderer := flag.String("renderer", "software", "renderer driver: software, gpu, auto or an SDL driver name")
	flag.Parse()

	b, err := sdl3.New(sdl3.Config{Title: "goak demo", Width: 960, Height: 720, Renderer: *renderer, Resizable: true})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	ctx, err := goak.NewContext(b, goak.Options{FontSize: 18})
	if err != nil {
		log.Fatal(err)
	}

	s := &state{featC: true, volume: 50, steps: 3, drag: 1, name: "goak", showTab2: true,
		notes: "Multiline text.\nSelect, copy and paste."}
	if s.imageTex, err = b.NewTexture(64, 64, gradient(64)); err != nil {
		log.Fatal(err)
	}
	colorsList := []string{"Red", "Green", "Blue", "Yellow", "Purple"}
	files := []string{"README.md", "main.go", "go.mod", "scene.go", "window.go", "layout.go", "theme.go", "menu.go"}

	var in goak.Input
	for b.NewFrame(&in) {
		ctx.NewFrame(&in, b.DisplaySize())
		s.build(ctx, colorsList, files)
		dd := ctx.Render()
		_ = b.Clear(colors.RGB(24, 24, 28))
		if err := b.Render(dd); err != nil {
			log.Println("render:", err)
		}
		_ = b.Present()
		if ctx.TextErr != nil {
			log.Println("text:", ctx.TextErr)
			ctx.TextErr = nil
		}
		if in.KeyPressed[goak.KeyEscape] && ctx.FocusID() == 0 && !ctx.AnyPopupOpen() {
			break
		}
	}
}

func (s *state) build(c *goak.Context, colorsList, files []string) {
	s.mainMenu(c)
	s.widgetsWindow(c, colorsList, files)
	s.textWindow(c)
	s.tabsWindow(c)
}

func (s *state) mainMenu(c *goak.Context) {
	if c.BeginMainMenuBar() {
		if c.BeginMenu("File") {
			if c.MenuItem("New", "Ctrl+N", false) {
				s.lastMenu = "File -> New"
			}
			if c.BeginMenu("Open Recent") {
				c.MenuItem("main.go", "", false)
				c.MenuItem("README.md", "", false)
				c.EndMenu()
			}
			if c.MenuItem("Exit", "", false) {
				s.lastMenu = "File -> Exit"
			}
			c.EndMenu()
		}
		if c.BeginMenu("Edit") {
			for _, n := range []string{"Cut", "Copy", "Paste"} {
				if c.MenuItem(n, "", false) {
					s.lastMenu = "Edit -> " + n
				}
			}
			c.EndMenu()
		}
		if c.BeginMenu("View") {
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
}

func (s *state) widgetsWindow(c *goak.Context, colorsList, files []string) {
	c.SetNextWindowPos(goak.Vec2{X: 16, Y: 40}, goak.CondOnce)
	c.SetNextWindowSize(goak.Vec2{X: 440, Y: 660}, goak.CondOnce)
	if c.Begin("Widgets", nil, 0) {
		if c.CollapsingHeader("Buttons") {
			if c.Button("Click Me!") {
				s.clicks++
			}
			c.SameLine(0, -1)
			if c.Button("Press Me!") {
				s.clicks += 10
			}
			if c.IsItemHovered() {
				c.SetTooltip("Adds ten")
			}
			c.Textf("clicks %d, last menu: %s", s.clicks, s.lastMenu)
			if c.ImageButton("imgbtn", s.imageTex, goak.Vec2{X: 24, Y: 24}, goak.Vec2{}, goak.Vec2{X: 1, Y: 1}, colors.Transparent, colors.White) {
				s.clicks++
			}
			c.SameLine(0, -1)
			c.Image(s.imageTex, goak.Vec2{X: 48, Y: 48}, goak.Vec2{}, goak.Vec2{X: 1, Y: 1}, colors.White)
			c.SmallButton("Small")
			c.SameLine(0, -1)
			c.ArrowButton("left", goak.DirLeft)
			c.SameLine(0, -1)
			c.ArrowButton("right", goak.DirRight)
			c.SameLine(0, -1)
			c.ArrowButton("up", goak.DirUp)
			c.SameLine(0, -1)
			c.ArrowButton("down", goak.DirDown)
			c.InvisibleButton("hit", goak.Vec2{X: 120, Y: 20})
			c.SameLine(0, -1)
			c.Text("Invisible button")
		}
		if c.CollapsingHeader("Text") {
			c.Text("Plain text")
			c.TextColored(colors.RGB(240, 180, 60), "Colored text")
			c.TextDisabled("Disabled text")
			c.Bullet()
			c.SameLine(0, -1)
			c.Text("Bullet")
			c.TextWrapped("Wrapped text breaks at spaces to fit the space left in the window, so long sentences stay readable.")
		}
		if c.CollapsingHeader("Checkboxes and radios") {
			c.Checkbox("Enable feature A", &s.featA)
			c.Checkbox("Enable feature B", &s.featB)
			c.Checkbox("Enable feature C", &s.featC)
			c.Separator()
			for i, n := range []string{"Option 1", "Option 2", "Option 3", "Option 4"} {
				c.RadioButtonInt(n, &s.radio, i)
				if i < 3 {
					c.SameLine(0, -1)
				}
			}
		}
		if c.CollapsingHeader("Sliders and progress") {
			c.SliderFloat("Volume", &s.volume, 0, 100, "%.1f")
			c.SliderInt("Steps", &s.steps, 0, 10, "%d")
			c.DragFloat("Drag", &s.drag, 0.05, 0, 10, "%.2f")
			s.progress += 0.003
			if s.progress > 1 {
				s.progress = 0
			}
			c.ProgressBar(s.progress, goak.Vec2{X: -1}, fmt.Sprintf("%d%%", int(s.progress*100)))
		}
		if c.CollapsingHeader("Combo and list") {
			c.Combo("Color", &s.color, colorsList)
			c.ListBox("Files", &s.list, files, 5)
		}
		if c.CollapsingHeader("Tree") {
			if c.TreeNode("Root") {
				if c.TreeNode("Child A") {
					c.BulletText("leaf one")
					c.BulletText("leaf two")
					c.TreePop()
				}
				if c.TreeNodeEx("Leaf node", goak.TreeLeaf) {
					c.TreePop()
				}
				c.TreePop()
			}
		}
		scopeSection(c)
		plotLinesSection(c)
		c.SeparatorText("Context menu")
		c.Text("Right click this text")
		if c.BeginPopupContextItem("ctx") {
			for _, n := range []string{"Copy", "Paste", "Delete"} {
				if c.MenuItem(n, "", false) {
					s.lastMenu = "Context -> " + n
				}
			}
			c.EndPopup()
		}
	}
	c.End()
}

func (s *state) textWindow(c *goak.Context) {
	c.SetNextWindowPos(goak.Vec2{X: 480, Y: 40}, goak.CondOnce)
	c.SetNextWindowSize(goak.Vec2{X: 440, Y: 330}, goak.CondOnce)
	if c.Begin("Text input", nil, 0) {
		c.InputTextWithHint("Name", "type a name", &s.name, 0)
		c.InputText("Password", &s.password, goak.InputPassword)
		c.InputText("Read only", &s.name, goak.InputReadOnly)
		c.InputTextMultiline("##notes", &s.notes, goak.Vec2{Y: 120}, 0)
		c.Text("Double click the title bar to collapse.")
	}
	c.End()
}

func (s *state) tabsWindow(c *goak.Context) {
	c.SetNextWindowPos(goak.Vec2{X: 480, Y: 390}, goak.CondOnce)
	c.SetNextWindowSize(goak.Vec2{X: 440, Y: 220}, goak.CondOnce)
	if c.Begin("Tabs", nil, goak.MenuBar) {
		if c.BeginMenuBar() {
			if c.BeginMenu("Tabs") {
				c.MenuItemCheck("Show second tab", "", &s.showTab2)
				c.EndMenu()
			}
			c.EndMenuBar()
		}
		if c.BeginTabBar("tabs") {
			if c.BeginTabItem("First", nil) {
				c.Text("First tab content")
				c.EndTabItem()
			}
			if c.BeginTabItem("Second", &s.showTab2) {
				c.Text("Closable second tab")
				c.EndTabItem()
			}
			c.EndTabBar()
		}
	}
	c.End()
}
