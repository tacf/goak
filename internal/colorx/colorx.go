// Package colorx converts between RGB and HSV with components in 0..1.
package colorx

// HSVToRGB converts hue, saturation and value to red, green and blue.
func HSVToRGB(h, s, v float32) (r, g, b float32) {
	h -= float32(int(h))
	if h < 0 {
		h++
	}
	h *= 6
	i := int(h)
	f := h - float32(i)
	p, q, t := v*(1-s), v*(1-s*f), v*(1-s*(1-f))
	switch i % 6 {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	}
	return v, p, q
}

// RGBToHSV converts red, green and blue to hue, saturation and value.
// Hue is 0 for grays.
func RGBToHSV(r, g, b float32) (h, s, v float32) {
	mx, mn := max(r, g, b), min(r, g, b)
	v = mx
	d := mx - mn
	if mx <= 0 || d <= 0 {
		return 0, 0, v
	}
	s = d / mx
	switch mx {
	case r:
		h = (g - b) / d
	case g:
		h = 2 + (b-r)/d
	default:
		h = 4 + (r-g)/d
	}
	h /= 6
	if h < 0 {
		h++
	}
	return h, s, v
}
