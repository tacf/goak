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
	label      string
	min        float64
	max        float64
	value      float64
	step       float64
	onChanged  func(SliderChangedEvent)
	isDragging bool
	showValue  bool
}

// NewSlider creates a standalone slider. Add it with panel.AddSlider(slider).
func NewSlider(width, height layout.Size, label string, min, max, initial float64) *Slider {
	if min > max {
		min, max = max, min
	}
	return &Slider{
		c:         layout.NewContainer(width, height),
		label:     label,
		min:       min,
		max:       max,
		value:     clamp(initial, min, max),
		step:      (max - min) / 100.0,
		showValue: true,
	}
}

// Bounds returns the computed layout rect after Layout.
func (s *Slider) Bounds() layout.Rect { return s.c.Bounds }

// Container returns the layout node for this slider (internal use).
func (s *Slider) Container() *layout.Container { return s.c }

// Label returns the slider text.
func (s *Slider) Label() string { return s.label }

// SetLabel updates the slider text.
func (s *Slider) SetLabel(label string) { s.label = label }

// Min returns the lower bound.
func (s *Slider) Min() float64 { return s.min }

// Max returns the upper bound.
func (s *Slider) Max() float64 { return s.max }

// Value returns the current slider value.
func (s *Slider) Value() float64 { return s.value }

// SetOnChanged assigns the slider value change callback.
func (s *Slider) SetOnChanged(onChanged func(SliderChangedEvent)) {
	s.onChanged = onChanged
}

// SetStep sets the increment step for the slider.
func (s *Slider) SetStep(step float64) {
	s.step = max(0, step)
}

// Step returns the slider increment.
func (s *Slider) Step() float64 { return s.step }

// SetValue updates the slider and emits a change event when necessary.
func (s *Slider) SetValue(value float64) {
	value = clamp(value, s.min, s.max)
	if s.step > 0 {
		value = s.min + float64(int((value-s.min)/s.step+0.5))*s.step
		value = clamp(value, s.min, s.max)
	}
	if value == s.value {
		return
	}
	previous := s.value
	s.value = value
	if s.onChanged != nil {
		s.onChanged(SliderChangedEvent{
			Slider:   s,
			Previous: previous,
			Value:    value,
		})
	}
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

	if s.label != "" {
		labelHeight = textHeight(s.label, font)
		rendering.DrawText(renderer, s.label, font, bound.X, bound.Y, theme.Text)
	}

	// Track position
	trackY := bound.Y + labelHeight + 4
	trackWidth := bound.W
	if s.showValue {
		trackWidth -= 50 // Reserve space for value display
	}

	rendering.FillRect(renderer, bound.X, trackY, trackWidth, trackHeight, theme.TrackFill)
	rendering.DrawStrokeRect(renderer, bound.X, trackY, trackWidth, trackHeight, 1.0, theme.TrackStroke)

	normalizedValue := 0.0
	if s.max > s.min {
		normalizedValue = (s.value - s.min) / (s.max - s.min)
	}
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
		valueStr := fmt.Sprintf("%.1f", s.value)
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
	if trackWidth <= 0 {
		return
	}

	normalizedX := (mouseX - bound.X) / trackWidth
	if normalizedX < 0 {
		normalizedX = 0
	}
	if normalizedX > 1 {
		normalizedX = 1
	}

	s.SetValue(s.min + normalizedX*(s.max-s.min))
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

func clamp(value, minValue, maxValue float64) float64 {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}
	return max(minValue, min(value, maxValue))
}
