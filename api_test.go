package goak_test

import (
	"testing"

	"goak"
	"goak/components"
	"goak/layout"
)

type testScene struct{}

func (testScene) Draw(*goak.SceneContext) {}

func TestPublicAPIComposesRetainedAndSceneInterfaces(t *testing.T) {
	app := goak.NewApp()
	ui := components.NewUI()
	ui.Root().CreatePanel(layout.AutoSize(), layout.AutoSize())

	if app == nil || ui == nil {
		t.Fatal("public constructors returned nil")
	}
	var _ goak.Scene = testScene{}
}
