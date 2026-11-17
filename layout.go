package framekit

// Layout is derived from the output size so the same scenes render correctly
// in landscape, square and portrait without per-format specs.
type Layout struct {
	W, H     int
	Scale    float64
	Margin   int
	FrameX0  int
	FrameY0  int
	FrameX1  int
	FrameY1  int
	HeaderY  int
	DividerY int
	ContentX int
	CenterY  int
	Portrait bool
	Square   bool
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
	margin := int(80 * scale)
	if margin < 32 {
		margin = 32
	}
	// The "frame" fields are now just the safe stage: no chrome is drawn.
	top := int(float64(h) * 0.10)
	bottom := int(float64(h) * 0.90)
	return Layout{
		W: w, H: h, Scale: scale, Margin: margin,
		FrameX0: margin, FrameY0: top,
		FrameX1: w - margin, FrameY1: bottom,
		HeaderY: 0, DividerY: 0,
		ContentX: margin + int(30*scale),
		CenterY:  h / 2,
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
		Small:  17 * l.Scale,
		Normal: 25 * l.Scale,
		Code:   31 * l.Scale,
		H2:     38 * l.Scale,
		H1:     56 * l.Scale,
		Hero:   70 * l.Scale,
	}
}
