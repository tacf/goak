package goak_test

import (
	"errors"
	"testing"

	"github.com/tacf/goak"
	"github.com/tacf/goak/components"
	"github.com/tacf/goak/layout"
)

func TestSceneContextRetainedUIAttachmentPublicAPI(t *testing.T) {
	var _ func(*goak.SceneContext, *components.UI) = (*goak.SceneContext).SetUI
	var _ func(*goak.SceneContext) = (*goak.SceneContext).ClearUI
	var _ func(*goak.SceneContext) *components.UI = (*goak.SceneContext).UI
	var _ func(*goak.SceneContext) float32 = (*goak.SceneContext).UIScale
	var _ func(*goak.SceneContext) layout.Rect = (*goak.SceneContext).UIViewport
	var _ func(*goak.SceneContext, float32, float32) (float64, float64) = (*goak.SceneContext).UIPoint
	var _ func(*goak.SceneContext, *components.ContextMenu, float32, float32) bool = (*goak.SceneContext).OpenContextMenu
	var _ func(*goak.SceneContext) = (*goak.SceneContext).CloseUIPopups

	// A SceneContext is normally supplied by RunScene. Keeping attachment
	// methods nil-safe makes deferred cleanup and optional overlays simple,
	// without claiming that a zero context can retain UI state without a host.
	ctx := new(goak.SceneContext)
	ctx.SetUI(components.NewUI())
	ctx.SetUI(nil)
	ctx.ClearUI()
	ctx.ClearUI()
	ctx.CloseUIPopups()
	if ctx.OpenContextMenu(components.NewContextMenu(nil), 1, 2) {
		t.Fatal("unhosted SceneContext opened a context menu")
	}
	if got := ctx.UI(); got != nil {
		t.Fatalf("unhosted SceneContext UI = %p, want nil", got)
	}
	if ctx.UIScale() != 1 || ctx.UIViewport() != (layout.Rect{}) {
		t.Fatal("unhosted SceneContext returned non-default UI metrics")
	}
	if x, y := ctx.UIPoint(12, 8); x != 12 || y != 8 {
		t.Fatalf("unhosted UI point = %v,%v, want 12,8", x, y)
	}

	var nilContext *goak.SceneContext
	nilContext.SetUI(components.NewUI())
	nilContext.ClearUI()
	if got := nilContext.UI(); got != nil {
		t.Fatalf("nil SceneContext UI = %p, want nil", got)
	}
}

func TestSceneContextUIInputPolicy(t *testing.T) {
	ctx := new(goak.SceneContext)
	if got := ctx.UIInputPolicy(); got != goak.UIInputOverlay {
		t.Fatalf("default UI input policy = %v, want UIInputOverlay", got)
	}

	policies := []goak.UIInputPolicy{
		goak.UIInputOverlay,
		goak.UIInputModal,
		goak.UIInputPassthrough,
	}
	for _, policy := range policies {
		ctx.SetUIInputPolicy(policy)
		if got := ctx.UIInputPolicy(); got != policy {
			t.Errorf("UI input policy after SetUIInputPolicy(%v) = %v", policy, got)
		}
	}

	if err := ctx.SetUIInputPolicy(goak.UIInputPolicy(255)); !errors.Is(err, goak.ErrInvalidUIInputPolicy) {
		t.Fatalf("invalid UI input policy error = %v, want %v", err, goak.ErrInvalidUIInputPolicy)
	}
	if got := ctx.UIInputPolicy(); got != goak.UIInputPassthrough {
		t.Fatalf("invalid UI input policy changed state to %v", got)
	}
}

func TestRunSceneWithUIPublicSignatureAndPreconditions(t *testing.T) {
	var _ func(*goak.App, goak.Scene, *components.UI) error = (*goak.App).RunSceneWithUI

	app := goak.NewApp()
	err := app.RunSceneWithUI(testScene{}, components.NewUI())
	if !errors.Is(err, goak.ErrWindowNotInitialized) {
		t.Fatalf("RunSceneWithUI before window initialization = %v, want %v", err, goak.ErrWindowNotInitialized)
	}
}
