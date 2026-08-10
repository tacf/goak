package goak

import (
	"testing"

	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

type inputRecordingScene struct {
	events []Event
}

func (*inputRecordingScene) Draw(*SceneContext) {}

func (scene *inputRecordingScene) HandleEvent(event Event) bool {
	scene.events = append(scene.events, event)
	return true
}

func TestSceneUIInputPoliciesRouteHandledAndUnhandledClicks(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(100), layout.StaticPx(50))
	var clicks int
	button := panel.CreateButton(layout.StaticPx(60), layout.StaticPx(30), "Action")
	button.SetAction(func() { clicks++ })
	layout.Layout(ui.Root().Container(), 200, 100)

	tests := []struct {
		name       string
		policy     UIInputPolicy
		x          float32
		wantClicks int
		wantEvents int
	}{
		{name: "overlay handled", policy: UIInputOverlay, x: 10, wantClicks: 1},
		{name: "overlay unhandled", policy: UIInputOverlay, x: 150, wantEvents: 1},
		{name: "modal unhandled", policy: UIInputModal, x: 150},
		{name: "passthrough control", policy: UIInputPassthrough, x: 10, wantEvents: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clicks = 0
			scene := &inputRecordingScene{}
			win := &Window{width: 200, height: 100, ui: ui, scene: scene}
			ctx := &SceneContext{window: win, uiInputPolicy: test.policy}
			win.sceneCtx = ctx
			uiAcceptsInput := test.policy != UIInputPassthrough
			uiHandled := false
			if uiAcceptsInput {
				uiHandled = win.mouseDown(MouseLeft, float64(test.x), 10)
			}
			win.dispatchSceneEvent(
				Event{Type: EventMouseDown, X: test.x, Y: 10, Button: MouseLeft},
				uiHandled, test.policy == UIInputModal,
			)

			if clicks != test.wantClicks || len(scene.events) != test.wantEvents {
				t.Fatalf("clicks/events = %d/%d, want %d/%d",
					clicks, len(scene.events), test.wantClicks, test.wantEvents)
			}
		})
	}
}

func TestModalSceneUIStillForwardsQuit(t *testing.T) {
	scene := &inputRecordingScene{}
	ui := components.NewUI()
	win := &Window{ui: ui, scene: scene, running: true}
	win.sceneCtx = &SceneContext{window: win, uiInputPolicy: UIInputModal}

	win.dispatchSceneEvent(Event{Type: EventQuit}, false, true)

	if len(scene.events) != 1 || scene.events[0].Type != EventQuit || !win.running {
		t.Fatalf("modal quit routing = events %v running %v", scene.events, win.running)
	}
}

func TestModalSceneUIBlocksSceneWhenControlsAreNonInteractive(t *testing.T) {
	ui := components.NewUI()
	ui.SetInteractive(false)
	win := &Window{ui: ui}

	accepts, blocks := win.uiInputRouting(UIInputModal)
	if accepts || !blocks {
		t.Fatalf("modal routing accepts=%v blocks=%v, want false,true", accepts, blocks)
	}
	accepts, blocks = win.uiInputRouting(UIInputOverlay)
	if accepts || blocks {
		t.Fatalf("overlay routing accepts=%v blocks=%v, want false,false", accepts, blocks)
	}
}

func TestUIPointerOwnershipTracksButtonsAndSurvivesUIClear(t *testing.T) {
	ui := components.NewUI()
	win := &Window{ui: ui}
	const (
		left  = uint8(1)
		right = uint8(3)
	)
	win.setUIPointerPress(left, true)
	win.setUIPointerPress(right, true)

	win.releaseUI()
	if win.uiPointerPress == 0 {
		t.Fatal("clearing the UI discarded an in-flight pointer sequence")
	}
	if !win.finishUIPointerRelease(left) {
		t.Fatal("left release was not owned by the UI")
	}
	if win.uiPointerPress&pointerButtonBit(right) == 0 {
		t.Fatal("left release cleared right-button ownership")
	}
	if !win.finishUIPointerRelease(right) || win.uiPointerPress != 0 {
		t.Fatal("right release did not finish the remaining UI pointer sequence")
	}
}

func TestUpdateUIAllowsSliderCallbackToClearUI(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(120), layout.StaticPx(40))
	slider := panel.CreateSlider(layout.StaticPx(120), layout.StaticPx(30), "", 0, 100, 0)
	win := &Window{width: 120, height: 40, ui: ui, mouseX: 100}
	slider.SetOnChanged(func(components.SliderChangedEvent) { win.setUI(nil) })
	slider.StartDrag()

	win.updateUI()
	if win.ui != nil {
		t.Fatal("slider callback did not clear the UI")
	}
}

func TestOpenDropdownListPrecedesOverlappingButton(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(100), layout.StaticPx(60))
	button := panel.CreateButton(layout.StaticPx(100), layout.StaticPx(20), "Button")
	dropdown := panel.CreateDropdown(
		layout.StaticPx(100), layout.StaticPx(20), "Choose",
		[]components.DropdownOption{{Label: "First", Value: "first"}},
	)
	// Deliberately overlap the button with the expanded list to exercise popup
	// ordering independently of the normal non-overlapping layout flow.
	panel.Container().Children[0].Bounds = layout.Rect{X: 0, Y: 20, W: 100, H: 20}
	panel.Container().Children[1].Bounds = layout.Rect{X: 0, Y: 0, W: 100, H: 20}
	var buttonClicks int
	button.SetAction(func() { buttonClicks++ })
	dropdown.Open()
	win := &Window{ui: ui}

	if !win.mouseDown(MouseLeft, 10, 25) {
		t.Fatal("open dropdown list did not handle the click")
	}
	if dropdown.SelectedIndex() != 0 || buttonClicks != 0 {
		t.Fatalf("selection=%d button clicks=%d, want 0 and 0",
			dropdown.SelectedIndex(), buttonClicks)
	}
}

func TestSceneContextMenusRequireExplicitSceneOpening(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	menu := components.NewContextMenu(nil)
	panel.AddContextMenu(menu)
	win := &Window{ui: ui, scene: &inputRecordingScene{}}

	if win.openContextMenu(10, 10) || menu.IsOpen() {
		t.Fatal("scene overlay context menu auto-opened without scene context")
	}
}

func TestPassthroughPolicySynchronouslyReconcilesRetainedInput(t *testing.T) {
	ui := components.NewUI()
	panel := ui.Root().CreatePanel(layout.StaticPx(120), layout.StaticPx(60))
	input := panel.CreateTextInput(layout.StaticPx(120), layout.StaticPx(30), "")
	menu := components.NewContextMenu(nil)
	panel.AddContextMenu(menu)
	win := &Window{ui: ui}
	ctx := &SceneContext{window: win}
	win.focusTextInput(input)
	menu.Open(0, 0)

	ctx.SetUIInputPolicy(UIInputPassthrough)
	if win.focusedInput != nil || input.Focused() || menu.IsOpen() {
		t.Fatal("passthrough left retained focus or popup state active")
	}
}

func TestSceneContextRetainedUICoordinateConversion(t *testing.T) {
	ui := components.NewUI()
	if err := ui.Root().SetScale(2); err != nil {
		t.Fatal(err)
	}
	win := &Window{width: 200, height: 100, ui: ui}
	ctx := &SceneContext{window: win}

	if scale := ctx.UIScale(); scale != 2 {
		t.Fatalf("UI scale = %v, want 2", scale)
	}
	if viewport := ctx.UIViewport(); viewport != (layout.Rect{W: 100, H: 50}) {
		t.Fatalf("UI viewport = %+v, want 100x50", viewport)
	}
	if x, y := ctx.UIPoint(40, 30); x != 20 || y != 15 {
		t.Fatalf("UI point = %v,%v, want 20,15", x, y)
	}
}

func TestSceneContextOpensRegisteredContextMenuInUISpace(t *testing.T) {
	ui := components.NewUI()
	if err := ui.Root().SetScale(2); err != nil {
		t.Fatal(err)
	}
	panel := ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())
	menu := components.NewContextMenu([]components.ContextMenuItem{
		components.NewContextMenuAction("Open", nil),
	})
	panel.AddContextMenu(menu)
	win := &Window{width: 200, height: 100, ui: ui}
	ctx := &SceneContext{window: win}

	if ctx.OpenContextMenu(components.NewContextMenu(nil), 40, 30) {
		t.Fatal("opened a context menu not registered with the attached UI")
	}
	if !ctx.OpenContextMenu(menu, 40, 30) || !menu.IsOpen() {
		t.Fatal("registered context menu did not open")
	}
	bounds := menu.Bounds()
	viewport := ctx.UIViewport()
	if bounds.X < viewport.X || bounds.Y < viewport.Y ||
		bounds.X+bounds.W > viewport.X+viewport.W ||
		bounds.Y+bounds.H > viewport.Y+viewport.H {
		t.Fatalf("menu bounds %+v escape viewport %+v", bounds, viewport)
	}

	ctx.CloseUIPopups()
	if menu.IsOpen() {
		t.Fatal("CloseUIPopups left the context menu open")
	}
}
