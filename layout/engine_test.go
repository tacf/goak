package layout

import (
	"math"
	"testing"
)

func TestColumnSharesMainAxisAndFillsCrossAxis(t *testing.T) {
	first := NewContainer(AutoSize(), AutoSize())
	second := NewContainer(AutoSize(), AutoSize())
	root := NewContainer(AutoSize(), AutoSize(), first, second)

	Layout(root, 120, 90)

	assertRect(t, first.Bounds, Rect{X: 0, Y: 0, W: 120, H: 45})
	assertRect(t, second.Bounds, Rect{X: 0, Y: 45, W: 120, H: 45})
}

func TestLayoutDoesNotAllocatePerPass(t *testing.T) {
	children := make([]*Container, 64)
	for index := range children {
		children[index] = NewContainer(AutoSize().WithMin(2), AutoSize())
	}
	root := NewContainer(AutoSize(), AutoSize(), children...)
	root.Direction = Row

	if allocations := testing.AllocsPerRun(100, func() { Layout(root, 1280, 720) }); allocations != 0 {
		t.Fatalf("Layout allocations = %v, want 0", allocations)
	}
}

func TestColumnGapParticipatesInAutoSizing(t *testing.T) {
	first := NewContainer(PercentOf(100), AutoSize())
	second := NewContainer(PercentOf(100), AutoSize())
	root := NewContainer(AutoSize(), AutoSize(), first, second)
	root.Gap = 10

	Layout(root, 100, 100)

	assertRect(t, first.Bounds, Rect{X: 0, Y: 0, W: 100, H: 45})
	assertRect(t, second.Bounds, Rect{X: 0, Y: 55, W: 100, H: 45})
}

func TestRowDistributesMainAxisAndFillsCrossAxis(t *testing.T) {
	first := NewContainer(StaticPx(20), AutoSize())
	second := NewContainer(AutoSize(), AutoSize())
	third := NewContainer(AutoSize(), AutoSize())
	root := NewContainer(AutoSize(), AutoSize(), first, second, third)
	root.Direction = Row
	root.Gap = 10

	Layout(root, 120, 50)

	assertRect(t, first.Bounds, Rect{X: 0, Y: 0, W: 20, H: 50})
	assertRect(t, second.Bounds, Rect{X: 30, Y: 0, W: 40, H: 50})
	assertRect(t, third.Bounds, Rect{X: 80, Y: 0, W: 40, H: 50})
}

func TestRowAlignmentPositionsGroupAndIndividualChildren(t *testing.T) {
	first := NewContainer(StaticPx(20), StaticPx(10))
	second := NewContainer(StaticPx(30), StaticPx(20))
	root := NewContainer(AutoSize(), AutoSize(), first, second)
	root.Direction = Row
	root.Gap = 10
	root.HorizontalAlign = AlignCenter
	root.VerticalAlign = AlignEnd

	Layout(root, 200, 50)

	assertRect(t, first.Bounds, Rect{X: 70, Y: 40, W: 20, H: 10})
	assertRect(t, second.Bounds, Rect{X: 100, Y: 30, W: 30, H: 20})
}

func TestPaddingInsetsChildLayout(t *testing.T) {
	child := NewContainer(PercentOf(50), StaticPx(20))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 10

	Layout(root, 100, 80)

	want := Rect{X: 10, Y: 10, W: 40, H: 20}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}

func TestPaddingConstrainsCenteredContent(t *testing.T) {
	child := NewContainer(StaticPx(20), StaticPx(10))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 10
	root.HorizontalAlign = AlignCenter
	root.VerticalAlign = AlignCenter

	Layout(root, 100, 60)

	want := Rect{X: 40, Y: 25, W: 20, H: 10}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}

func TestPaddingIsClampedToContainer(t *testing.T) {
	child := NewContainer(PercentOf(100), PercentOf(100))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 100

	Layout(root, 40, 20)

	want := Rect{X: 20, Y: 10, W: 0, H: 0}
	if child.Bounds != want {
		t.Fatalf("child bounds = %+v, want %+v", child.Bounds, want)
	}
}

func TestInsetsOverrideLegacyPadding(t *testing.T) {
	child := NewContainer(PercentOf(100), PercentOf(100))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Padding = 40
	root.Insets = Insets{Top: 5, Right: 10, Bottom: 15, Left: 20}

	Layout(root, 100, 80)

	assertRect(t, child.Bounds, Rect{X: 20, Y: 5, W: 70, H: 60})
}

func TestOversizedInsetsAreReducedProportionally(t *testing.T) {
	child := NewContainer(PercentOf(100), PercentOf(100))
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Insets = Insets{Top: 30, Right: 30, Bottom: 10, Left: 10}

	Layout(root, 40, 20)

	assertRect(t, child.Bounds, Rect{X: 10, Y: 15, W: 0, H: 0})
}

func TestSizeConstraintsApplyToAllSizingModes(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want float64
	}{
		{name: "static minimum", size: StaticPx(10).WithMin(20), want: 20},
		{name: "percent maximum", size: PercentOf(100).WithMax(60), want: 60},
		{name: "auto minimum", size: AutoSize().WithMin(120), want: 120},
		{name: "minimum wins invalid range", size: AutoSize().WithMin(40).WithMax(20), want: 40},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveSize(test.size, 100); got != test.want {
				t.Fatalf("resolveSize() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAutoMaximumRedistributesUnusedSpace(t *testing.T) {
	limited := NewContainer(AutoSize().WithMax(20), AutoSize())
	flexible := NewContainer(AutoSize(), AutoSize())
	root := NewContainer(AutoSize(), AutoSize(), limited, flexible)
	root.Direction = Row

	Layout(root, 100, 30)

	assertRect(t, limited.Bounds, Rect{X: 0, Y: 0, W: 20, H: 30})
	assertRect(t, flexible.Bounds, Rect{X: 20, Y: 0, W: 80, H: 30})
}

func TestAutoMinimumsAreHonoredWhenTheyOverflow(t *testing.T) {
	first := NewContainer(AutoSize().WithMin(70), AutoSize())
	second := NewContainer(AutoSize().WithMin(50), AutoSize())
	root := NewContainer(AutoSize(), AutoSize(), first, second)
	root.Direction = Row

	Layout(root, 100, 20)

	assertRect(t, first.Bounds, Rect{X: 0, Y: 0, W: 70, H: 20})
	assertRect(t, second.Bounds, Rect{X: 70, Y: 0, W: 50, H: 20})
}

func TestUnknownDirectionFallsBackToColumn(t *testing.T) {
	first := NewContainer(StaticPx(10), StaticPx(10))
	second := NewContainer(StaticPx(10), StaticPx(10))
	root := NewContainer(AutoSize(), AutoSize(), first, second)
	root.Direction = Direction(99)

	Layout(root, 100, 100)

	assertRect(t, first.Bounds, Rect{X: 0, Y: 0, W: 10, H: 10})
	assertRect(t, second.Bounds, Rect{X: 0, Y: 10, W: 10, H: 10})
}

func TestInvalidGeometryIsNormalizedAtPublicAndLayoutBoundaries(t *testing.T) {
	if got := StaticPx(math.NaN()).Value; got != 0 {
		t.Fatalf("NaN static size = %v, want 0", got)
	}
	if got := PercentOf(math.Inf(1)).Value; got != 0 {
		t.Fatalf("infinite percent = %v, want 0", got)
	}
	if got := PercentOf(150).Value; got != 100 {
		t.Fatalf("oversized percent = %v, want 100", got)
	}
	child := &Container{Width: Size{Kind: Static, Value: math.NaN()}, Height: AutoSize()}
	root := NewContainer(AutoSize(), AutoSize(), child)
	root.Gap = math.Inf(1)
	Layout(root, math.Inf(1), 20)
	assertRect(t, root.Bounds, Rect{W: 0, H: 20})
	assertRect(t, child.Bounds, Rect{W: 0, H: 20})
	Layout(nil, 10, 10)
}

func assertRect(t *testing.T, got, want Rect) {
	t.Helper()
	if got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}
