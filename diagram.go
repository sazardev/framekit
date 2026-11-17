package framekit

import "math"

type rect struct{ x, y, w, h float64 }

func contentRect(l Layout) rect {
	S := l.Scale
	return rect{
		x: float64(l.FrameX0) + 40*S,
		y: float64(l.FrameY0) + 40*S,
		w: float64(l.FrameX1-l.FrameX0) - 80*S,
		h: float64(l.FrameY1-l.FrameY0) - 80*S,
	}
}

func normPoint(area rect, p Point) (float64, float64) {
	return area.x + p.X*area.w, area.y + p.Y*area.h
}

func drawArrowHead(p *painter, x0, y0, x1, y1 float64, color string, alpha, S float64) {
	angle := math.Atan2(y1-y0, x1-x0)
	length := 13 * S
	spread := 0.38
	base := Mix(p.theme.Color(color), p.bg, alpha*0.25)
	for _, sign := range []float64{1, -1} {
		a := angle + math.Pi + sign*spread
		p.line(x1, y1, x1+length*math.Cos(a), y1+length*math.Sin(a), color, alpha)
	}
	_ = base
}

func drawDiagram(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Diagram == nil || len(sc.Diagram.Nodes) == 0 {
		return
	}
	d := sc.Diagram
	l := p.lay
	W, H, S := float64(l.W), float64(l.H), l.Scale
	area := contentRect(l)
	if sc.Title != "" {
		p.textAnchored(sc.Title, W/2, H*0.16, "h1-bold", "fg", fade)
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, fade*0.85))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawLine(W/2-50*S, H*0.22, W/2+50*S, H*0.22)
		p.ctx.Stroke()
	}

	// Resolve node rectangles.
	rects := map[string]rect{}
	n := len(d.Nodes)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := int(math.Ceil(float64(n) / float64(cols)))
	for i, node := range d.Nodes {
		w, h := 210*S, 92*S
		if node.W > 0 {
			w = area.w * node.W
		}
		if node.H > 0 {
			h = area.h * node.H
		}
		nx, ny := node.X, node.Y
		switch d.Layout {
		case "explicit":
		case "grid":
			nx = (float64(i%cols) + 0.5) / float64(cols)
			ny = (float64(i/cols) + 0.5) / float64(rows)
		default:
			nx = (float64(i) + 0.5) / float64(n)
			ny = 0.5
		}
		cx, cy := normPoint(area, Point{X: nx, Y: ny})
		rects[node.ID] = rect{cx - w/2, cy - h/2, w, h}
	}

	// Arrows first, drawn progressively with a traveling pulse.
	for i, e := range d.Edges {
		from, ok1 := rects[e.From]
		to, ok2 := rects[e.To]
		if !ok1 || !ok2 {
			continue
		}
		drawFlowEdge(p, spec, sc, lt, fade, from, to, e, float64(i), S)
	}

	// Nodes fade in with a subtle scale-in ("breathing" arrival).
	for i, node := range d.Nodes {
		r := rects[node.ID]
		at := node.At
		if at == 0 {
			at = 0.2 + float64(i)*0.28
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		scale := 0.94 + 0.06*Ease("easeOutQuint", a)
		cx, cy := r.x+r.w/2, r.y+r.h/2
		w, h := r.w*scale, r.h*scale
		rx, ry := cx-w/2, cy-h/2

		color := node.Accent
		if color == "" {
			color = "aqua"
		}
		c := Mix(p.theme.Color(color), p.bg, a*0.3)
		p.ctx.SetColor(Mix(p.theme.Color("deep"), p.bg, 0.4))
		p.ctx.DrawRectangle(rx, ry, w, h)
		p.ctx.Fill()
		p.ctx.SetColor(c)
		p.ctx.SetLineWidth(maxFloat(1.5*S, 1))
		if node.Shape == "diamond" {
			p.ctx.DrawLine(cx, ry, rx+w, cy)
			p.ctx.DrawLine(rx+w, cy, cx, ry+h)
			p.ctx.DrawLine(cx, ry+h, rx, cy)
			p.ctx.DrawLine(rx, cy, cx, ry)
		} else {
			p.ctx.DrawRectangle(rx+0.5, ry+0.5, w-1, h-1)
		}
		p.ctx.Stroke()
		p.textAnchored(node.Label, cx, cy+9*S, "h2", "fg", a)
	}
}
