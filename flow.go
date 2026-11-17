package framekit

import "math"

// partialLine returns the first `progress` fraction of a segment.
func partialLine(a, b pt, progress float64) (pt, pt) {
	progress = clamp01(progress)
	return a, pt{
		a.x + (b.x-a.x)*progress,
		a.y + (b.y-a.y)*progress,
	}
}

// drawFlowEdge draws a connection that grows from source to target, with a
// traveling pulse and an arrow head that only appears when it arrives.
func drawFlowEdge(p *painter, spec Spec, sc Scene, lt, fade float64, from, to rect, e DiagramEdge, index, S float64) {
	at := e.At
	if at == 0 {
		at = 0.45 + index*0.35
	}
	win := 1.15
	progress := clamp01((lt - at) / win)
	if lt < at {
		return
	}
	alpha := fade * EasedReveal(spec, lt, at)
	if alpha <= 0.01 {
		return
	}

	x0, y0 := from.x+from.w/2, from.y+from.h/2
	x1, y1 := to.x+to.w/2, to.y+to.h/2
	dx, dy := x1-x0, y1-y0
	if dx == 0 && dy == 0 {
		return
	}
	// Stop the line at each box border.
	t0 := math.Min((from.w/2+10*S)/math.Max(math.Abs(dx), 0.001), (from.h/2+10*S)/math.Max(math.Abs(dy), 0.001))
	t1 := math.Min((to.w/2+14*S)/math.Max(math.Abs(dx), 0.001), (to.h/2+14*S)/math.Max(math.Abs(dy), 0.001))
	sx, sy := x0+dx*t0, y0+dy*t0
	ex, ey := x1-dx*t1, y1-dy*t1

	color := e.Accent
	if color == "" {
		color = "aqua"
	}
	c := Mix(p.theme.Color(color), p.bg, alpha*0.25)

	// grow the line
	a, b := partialLine(pt{sx, sy}, pt{ex, ey}, Ease("easeInOut", progress))
	if progress > 0.02 {
		p.ctx.SetColor(c)
		p.ctx.SetLineWidth(maxFloat(2.2*S, 1.5))
		p.ctx.DrawLine(a.x, a.y, b.x, b.y)
		p.ctx.Stroke()
	}

	// traveling pulse
	if progress < 1 {
		p.ctx.SetColor(p.theme.Color("yellow"))
		r := 5 * S
		p.ctx.DrawRectangle(b.x-r/2, b.y-r/2, r, r)
		p.ctx.Fill()
	}

	// arrow head on arrival
	if progress >= 0.98 {
		drawArrowHead(p, a.x, a.y, b.x, b.y, color, alpha, S*1.3)
	}
	if e.Label != "" && progress > 0.6 {
		la := alpha * Ease("easeOutCubic", (progress-0.6)/0.4)
		mx, my := (sx+ex)/2, sy-30*S
		w := p.textWidth(e.Label, "normal") + 24*S
		p.ctx.SetColor(Mix(p.theme.Color("bg"), p.bg, 1-la))
		p.ctx.DrawRectangle(mx-w/2, my-20*S, w, 34*S)
		p.ctx.Fill()
		p.textAnchored(e.Label, mx, my, "normal", "dim", la)
	}
}
