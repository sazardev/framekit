package framekit

import (
	"math"
)

type rect struct{ x, y, w, h float64 }

func contentRect(l Layout) rect {
	S := l.Scale
	return rect{
		x: float64(l.FrameX0) + 60*S,
		y: float64(l.FrameY0) + 100*S,
		w: float64(l.FrameX1-l.FrameX0) - 120*S,
		h: float64(l.FrameY1-l.FrameY0) - 180*S,
	}
}

func normPoint(area rect, p Point) (float64, float64) {
	return area.x + p.X*area.w, area.y + p.Y*area.h
}

func drawDiagram(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Diagram == nil || len(sc.Diagram.Nodes) == 0 {
		return
	}
	d := sc.Diagram
	l, S := p.lay, p.lay.Scale
	area := contentRect(l)
	if sc.Title != "" {
		p.text(sc.Title, area.x, float64(l.FrameY0)+72*S, "h2-bold", "fg", fade)
	}

	// Resolve node rectangles.
	rects := map[string]rect{}
	n := len(d.Nodes)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := int(math.Ceil(float64(n) / float64(cols)))
	for i, node := range d.Nodes {
		w, h := 160*S, 66*S
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
		default: // row
			nx = (float64(i) + 0.5) / float64(n)
			ny = 0.5
		}
		cx, cy := normPoint(area, Point{X: nx, Y: ny})
		rects[node.ID] = rect{cx - w/2, cy - h/2, w, h}
	}

	// Edges first so nodes sit on top.
	for i, e := range d.Edges {
		from, ok1 := rects[e.From]
		to, ok2 := rects[e.To]
		if !ok1 || !ok2 {
			continue
		}
		at := e.At
		if at == 0 {
			at = 0.5 + float64(i)*0.28
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		x0, y0 := from.x+from.w/2, from.y+from.h/2
		x1, y1 := to.x+to.w/2, to.y+to.h/2
		dx, dy := x1-x0, y1-y0
		if dx == 0 && dy == 0 {
			continue
		}
		t := math.Min((from.w/2+6*S)/math.Max(math.Abs(dx), 0.001), (from.h/2+6*S)/math.Max(math.Abs(dy), 0.001))
		sx, sy := x0+dx*t, y0+dy*t
		t2 := math.Min((to.w/2+8*S)/math.Max(math.Abs(dx), 0.001), (to.h/2+8*S)/math.Max(math.Abs(dy), 0.001))
		ex, ey := x1-dx*t2, y1-dy*t2
		color := e.Accent
		if color == "" {
			color = "faint"
		}
		p.line(sx, sy, ex, ey, color, a)
		drawArrowHead(p, sx, sy, ex, ey, color, a, S)
		if e.Label != "" {
			p.text(e.Label, (sx+ex)/2, (sy+ey)/2-8*S, "small", "dim", a)
		}
	}

	// Nodes.
	for i, node := range d.Nodes {
		r := rects[node.ID]
		at := node.At
		if at == 0 {
			at = 0.3 + float64(i)*0.22
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		color := node.Accent
		if color == "" {
			color = "aqua"
		}
		c := p.theme.Color(color)
		if a < 0.999 {
			c = Mix(c, p.bg, a)
		}
		p.ctx.SetColor(Mix(p.theme.Color("deep"), p.bg, 1-a))
		p.ctx.DrawRectangle(r.x, r.y, r.w, r.h)
		p.ctx.Fill()
		p.ctx.SetColor(c)
		p.ctx.SetLineWidth(1)
		if node.Shape == "diamond" {
			p.ctx.DrawLine(r.x+r.w/2, r.y, r.x+r.w, r.y+r.h/2)
			p.ctx.DrawLine(r.x+r.w, r.y+r.h/2, r.x+r.w/2, r.y+r.h)
			p.ctx.DrawLine(r.x+r.w/2, r.y+r.h, r.x, r.y+r.h/2)
			p.ctx.DrawLine(r.x, r.y+r.h/2, r.x+r.w/2, r.y)
		} else {
			p.ctx.DrawRectangle(r.x+0.5, r.y+0.5, r.w-1, r.h-1)
		}
		p.ctx.Stroke()
		p.textAnchored(node.Label, r.x+r.w/2, r.y+r.h/2+7*S, "normal", "fg", a)
	}
}

func drawArrowHead(p *painter, x0, y0, x1, y1 float64, color string, alpha, S float64) {
	angle := math.Atan2(y1-y0, x1-x0)
	len := 11 * S
	spread := 0.42
	for _, sign := range []float64{1, -1} {
		a := angle + math.Pi + sign*spread
		p.line(x1, y1, x1+len*math.Cos(a), y1+len*math.Sin(a), color, alpha)
	}
}
