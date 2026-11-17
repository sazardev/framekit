package framekit

// Layout is derived from the output size so the same scenes render correctly
// in landscape, square and portrait without per-format specs.
type Layout struct {
	W, H       int
	Scale      float64
	Margin     int
	FrameX0    int
	FrameY0    int
	FrameX1    int
	FrameY1    int
	HeaderY    int
	DividerY   int
	ContentX   int
	CenterY    int
	Portrait   bool
	Square     bool
}

func ComputeLayout(sz Size) Layout {
	w, h := sz.W, sz.H
	scale := float64(w) / 1280
	if hs := float64(h) / 720; hs < scale {
		scale = hs
	}
	if scale < 0.45 {
		scale = 0.45
	}
	margin := int(70 * scale)
	if margin < 24 {
		margin = 24
	}
	top := int(112 * scale)
	bottom := h - int(78*scale)
	if bottom <= top {
		top, bottom = int(float64(h)*0.12), int(float64(h)*0.88)
	}
	return Layout{
		W: w, H: h, Scale: scale, Margin: margin,
		FrameX0: margin, FrameY0: top,
		FrameX1: w - margin, FrameY1: bottom,
		HeaderY: int(44 * scale), DividerY: int(78 * scale),
		ContentX: margin + int(60*scale),
		CenterY:  (top + bottom) / 2,
		Portrait: float64(h)/float64(w) > 1.4,
		Square:   float64(h)/float64(w) <= 1.4 && float64(h)/float64(w) >= 0.9,
	}
}

// FontSizes returns the type ramp for this layout.
type FontRamp struct {
	Small, Normal, Code, H2, H1, Hero float64
}

func (l Layout) Fonts() FontRamp {
	return FontRamp{
		Small:  16 * l.Scale,
		Normal: 22 * l.Scale,
		Code:   26 * l.Scale,
		H2:     30 * l.Scale,
		H1:     46 * l.Scale,
		Hero:   52 * l.Scale,
	}
}
