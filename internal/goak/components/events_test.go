package components

import (
	"testing"

	"github.com/tacf/goak/internal/goak/layout"
)

func TestButtonRunsActionAndTypedCallback(t *testing.T) {
	button := NewButton(layout.AutoSize(), layout.AutoSize(), "Run")
	var actionRuns int
	var event ButtonClickEvent
	button.SetAction(func() { actionRuns++ })
	button.SetOnClick(func(got ButtonClickEvent) { event = got })

	button.Click()
	if actionRuns != 1 {
		t.Fatalf("action runs = %d, want 1", actionRuns)
	}
	if event.Button != button {
		t.Fatal("button event has incorrect source")
	}
}

func TestCheckboxChangeEventIncludesPreviousState(t *testing.T) {
	checkbox := NewCheckbox(layout.AutoSize(), layout.AutoSize(), "Enabled")
	var event CheckboxChangedEvent
	checkbox.SetOnChanged(func(got CheckboxChangedEvent) { event = got })

	checkbox.SetChecked(true)
	if !checkbox.Checked() || event.Checkbox != checkbox || event.Previous || !event.Checked {
		t.Fatalf("unexpected checkbox event: %#v", event)
	}

	event = CheckboxChangedEvent{}
	checkbox.SetChecked(true)
	if event.Checkbox != nil {
		t.Fatal("unchanged checkbox emitted an event")
	}
}

func TestRadioAndDropdownEmitTypedSelectionEvents(t *testing.T) {
	radio := NewRadioGroup(
		layout.AutoSize(),
		layout.AutoSize(),
		[]RadioOption{{Label: "One", Value: "one"}, {Label: "Two", Value: "two"}},
	)
	var radioEvent RadioChangedEvent
	radio.SetOnChanged(func(got RadioChangedEvent) { radioEvent = got })
	radio.SetSelectedIndex(1)
	if radioEvent.RadioGroup != radio || radioEvent.PreviousIndex != -1 ||
		radioEvent.Index != 1 || radioEvent.Option.Value != "two" {
		t.Fatalf("unexpected radio event: %#v", radioEvent)
	}

	dropdown := NewDropdown(
		layout.AutoSize(),
		layout.AutoSize(),
		"Choose",
		[]DropdownOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
	)
	dropdown.Open()
	var dropdownEvent DropdownChangedEvent
	dropdown.SetOnChanged(func(got DropdownChangedEvent) { dropdownEvent = got })
	dropdown.SetSelectedIndex(0)
	if dropdownEvent.Dropdown != dropdown || dropdownEvent.PreviousIndex != -1 ||
		dropdownEvent.Index != 0 || dropdownEvent.Option.Value != "a" {
		t.Fatalf("unexpected dropdown event: %#v", dropdownEvent)
	}
	if dropdown.IsOpen() {
		t.Fatal("dropdown remained open after selection")
	}
}

func TestSliderSetterClampsStepsAndEmitsEvent(t *testing.T) {
	slider := NewSlider(layout.AutoSize(), layout.AutoSize(), "Value", 0, 10, 0)
	slider.SetStep(2)
	var event SliderChangedEvent
	slider.SetOnChanged(func(got SliderChangedEvent) { event = got })

	slider.SetValue(9.1)
	if slider.Value() != 10 {
		t.Fatalf("slider value = %v, want 10", slider.Value())
	}
	if event.Slider != slider || event.Previous != 0 || event.Value != 10 {
		t.Fatalf("unexpected slider event: %#v", event)
	}
}

func TestMenuActionsAreReusableAndObservable(t *testing.T) {
	ui := NewUI()
	menu := ui.Root().CreateMenuBar(layout.StaticPx(28), MenuBarWidthFull)
	var runs int
	action := Action(func() { runs++ })
	menu.AddItem("Run", action)
	button := NewButton(layout.AutoSize(), layout.AutoSize(), "Run")
	button.SetAction(action)
	button.Click()
	var event MenuActionEvent
	menu.SetOnAction(func(got MenuActionEvent) { event = got })

	layout.Layout(ui.Root().Container(), 300, 200)
	item := menu.TopItemRects()[0]
	if !menu.OnMouseDown(item.X+1, item.Y+1) {
		t.Fatal("menu click was not consumed")
	}
	if runs != 2 || event.Menu != menu || event.TopIndex != 0 ||
		event.SubIndex != -1 || event.Label != "Run" {
		t.Fatalf("unexpected menu action: runs=%d event=%#v", runs, event)
	}
}

func TestContextMenuActionIsObservable(t *testing.T) {
	var runs int
	menu := NewContextMenu([]ContextMenuItem{
		NewContextMenuAction("Copy", func() { runs++ }),
		NewContextMenuSeparator(),
		NewDisabledContextMenuAction("Unavailable"),
	})
	var event ContextMenuActionEvent
	menu.SetOnAction(func(got ContextMenuActionEvent) { event = got })
	menu.Open(10, 20)
	menu.Click(0)

	if runs != 1 || event.Menu != menu || event.Index != 0 || event.Item.Label != "Copy" {
		t.Fatalf("unexpected context action: runs=%d event=%#v", runs, event)
	}
	if menu.IsOpen() {
		t.Fatal("context menu remained open after action")
	}
}

func TestLabelStateUsesSetters(t *testing.T) {
	label := NewLabel(layout.AutoSize(), layout.AutoSize(), "old")
	label.SetText("new")
	if label.Text() != "new" {
		t.Fatalf("label text = %q, want new", label.Text())
	}
}
