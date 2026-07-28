package components

import (
	"fmt"
	"goak/internal/goak/layout"
	"goak/internal/goak/rendering"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Slider is a horizontal slider control for selecting values in a range.
type Slider struct {
	c          *layout.Container
	Label      string
	Min        float64
	Max        float64
	Value      float64
	Step       float64
	OnChanged  func(float64)
	isDragging bool
	showValue  bool
}

// NewSlider creates a standalone slider. Add it with panel.AddSlider(slider).
func NewSlider(width, height layout.Size, label string, min, max, initial float64) *Slider {
	return &Slider{
		c:         layout.NewContainer(width, height),
		Label:     label,
		Min:       min,
		Max:       max,
		Value:     initial,
		Step:      (max - min) / 100.0,
		showValue: true,
	}
}

// Bounds returns the computed layout rect after Layout.
func (s *Slider) Bounds() layout.Rect { return s.c.Bounds }

// Container returns the layout node for this slider (internal use).
func (s *Slider) Container() *layout.Container { return s.c }

// SetStep sets the increment step for the slider.
func (s *Slider) SetStep(step float64) {
	s.Step = step
}

// SetShowValue controls whether the current value is displayed.
func (s *Slider) SetShowValue(show bool) {
	s.showValue = show
}

func (s *Slider) Draw(renderer *sdl.Renderer, font *rendering.Font, theme SliderTheme) {
	bound := s.Bounds()

	// Calculate dimensions
	labelHeight := 0.0
	trackHeight := 6.0
	thumbRadius := 8.0

	if s.Label != "" {
		labelHeight = textHeight(s.Label, font)
		rendering.DrawText(renderer, s.Label, font, bound.X, bound.Y, theme.Text)
	}

	// Track position
	trackY := bound.Y + labelHeight + 4
	trackWidth := bound.W
	if s.showValue {
		trackWidth -= 50 // Reserve space for value display
	}

	rendering.FillRect(renderer, bound.X, trackY, trackWidth, trackHeight, theme.TrackFill)
	rendering.DrawStrokeRect(renderer, bound.X, trackY, trackWidth, trackHeight, 1.0, theme.TrackStroke)

	normalizedValue := (s.Value - s.Min) / (s.Max - s.Min)
	if normalizedValue < 0 {
		normalizedValue = 0
	}
	if normalizedValue > 1 {
		normalizedValue = 1
	}
	fillWidth := trackWidth * normalizedValue
	if fillWidth > 0 {
		rendering.FillRect(renderer, bound.X, trackY, fillWidth, trackHeight, theme.FillColor)
	}

	thumbX := bound.X + fillWidth
	thumbY := trackY + trackHeight/2
	rendering.DrawFilledCircle(renderer, thumbX, thumbY, thumbRadius, theme.ThumbFill)
	rendering.DrawCircleStroke(renderer, thumbX, thumbY, thumbRadius, 1.5, theme.ThumbStroke)

	if s.showValue {
		valueStr := fmt.Sprintf("%.1f", s.Value)
		valueX := int(bound.X + trackWidth + 8)
		valueY := textTopY(valueStr, font, thumbY-thumbRadius, thumbRadius*2)
		rendering.DrawText(renderer, valueStr, font, float64(valueX), valueY, theme.Text)
	}
}

// UpdateValue sets the slider value from a mouse X coordinate.
func (s *Slider) UpdateValue(mouseX float64) {
	bound := s.Bounds()
	trackWidth := bound.W
	if s.showValue {
		trackWidth -= 50
	}

	normalizedX := (mouseX - bound.X) / trackWidth
	if normalizedX < 0 {
		normalizedX = 0
	}
	if normalizedX > 1 {
		normalizedX = 1
	}

	newValue := s.Min + normalizedX*(s.Max-s.Min)

	// Apply step
	if s.Step > 0 {
		newValue = float64(int(newValue/s.Step+0.5)) * s.Step
	}

	if newValue != s.Value {
		s.Value = newValue
		if s.OnChanged != nil {
			s.OnChanged(s.Value)
		}
	}
}

// StartDrag begins a drag operation.
func (s *Slider) StartDrag() {
	s.isDragging = true
}

// StopDrag ends a drag operation.
func (s *Slider) StopDrag() {
	s.isDragging = false
}

// IsDragging returns whether the slider is currently being dragged.
func (s *Slider) IsDragging() bool {
	return s.isDragging
}
