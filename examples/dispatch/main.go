package main

import (
	"context"
	"fmt"
	"time"

	"goak/internal/goak"
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
)

func main() {
	app := goak.NewApp()
	defer app.Destroy()

	app.InitWindow("Goak Dispatch Example", 520, 220)

	ui := components.NewUI()
	root := ui.Root()
	root.SetPadding(20)

	content := root.CreatePanel(layout.PercentOf(100), layout.PercentOf(100))
	content.SetBackground(colors.DarkGray)
	content.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	content.SetPadding(16)

	status := content.CreateLabel(layout.StaticPx(360), layout.StaticPx(32), "Idle")
	status.SetAlignment(layout.AlignCenter, layout.AlignCenter)

	start := content.CreateButton(layout.StaticPx(180), layout.StaticPx(36), "Start stream")
	start.SetOnClick(func(event components.ButtonClickEvent) {
		if event.Button.Label() == "Streaming..." {
			return
		}
		event.Button.SetLabel("Streaming...")
		status.SetText("Starting...")
		go consumeProgress(app, event.Button, status, progressStream(app.Context()))
	})

	app.Run(ui)
}

func progressStream(ctx context.Context) <-chan int {
	updates := make(chan int)
	go func() {
		defer close(updates)
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()

		for progress := 0; progress <= 100; progress += 2 {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			select {
			case <-ctx.Done():
				return
			case updates <- progress:
			}
		}
	}()
	return updates
}

func consumeProgress(
	app *goak.App,
	start *components.Button,
	status *components.Label,
	updates <-chan int,
) {
	for progress := range updates {
		text := fmt.Sprintf("Progress: %d%%", progress)
		if err := app.DispatchLatest("progress", func() {
			status.SetText(text)
		}); err != nil {
			return
		}
	}
	_ = app.Dispatch(func() {
		status.SetText("Complete")
		start.SetLabel("Start stream")
	})
}
