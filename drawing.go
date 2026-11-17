package framekit

import (
	"math"
)

type pt struct{ x, y float64 }

func drawDrawing(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Drawing == nil || len(sc.Drawing.Strokes) == 0 {
		return
	}
	l, S := p.lay, p.lay.Scale
	area := contentRect(l)
	if sc.Title != "" {
		p.text(sc.Title, area.x, float64(l.FrameY0)+72*S, "h2-bold", "fg", fade)
	}

	for _, st := range sc.Drawing.Strokes {
		at := st.At
		dur := st.Duration
		if dur <= 0 {
			dur = 0.8
		}
		progress := clamp01((lt - at) / dur)
		if lt < at {
			continue
		}
		full := strokePoints(st, area, S)
		if len(full) < 2 {
			continue
		}
		shown := partialPolyline(full, progress)
		if len(shown) < 2 {
			continue
		}
		color := st.Color
		if color == "" {
			color = "green"
		}
		c := p.theme.Color(color)
		if fade < 0.999 {
			c = Mix(c, p.bg, fade)
		}
		p.ctx.SetColor(c)
		p.ctx.SetLineWidth(maxFloat(st.Width*S, 1.5))
		p.ctx.DrawLine(shown[0].x, shown[0].y, shown[0].x, shown[0].y)
		for i := 1; i < len(shown); i++ {
			p.ctx.DrawLine(shown[i-1].x, shown[i-1].y, shown[i].x, shown[i].y)
		}
		p.ctx.Stroke()

		if st.Kind == "arrow" && progress >= 1 {
			a := shown[len(shown)-2]
			b := shown[len(shown)-1]
			drawArrowHead(p, a.x, a.y, b.x, b.y, color, fade, S)
		}

		// moving pen marker
		if progress < 1 {
			head := shown[len(shown)-1]
			size := 8 * S
			p.ctx.SetColor(p.theme.Color("yellow"))
			p.ctx.DrawRectangle(head.x-size/2, head.y-size/2, size, size)
			p.ctx.Fill()
		}
		if st.Label != "" && progress > 0.8 {
			last := shown[len(shown)-1]
			p.text(st.Label, last.x+12*S, last.y-10*S, "small", "dim", fade)
		}
	}
}

func strokePoints(st Stroke, area rect, S float64) []pt {
	switch st.Kind {
	case "line", "arrow", "path":
		pts := []pt{}
		for _, q := range st.Points {
			x, y := normPoint(area, q)
			pts = append(pts, pt{x, y})
		}
		return pts
	case "rect":
		x0, y0 := normPoint(area, Point{X: st.X, Y: st.Y})
		x1 := x0 + st.W*area.w
		y1 := y0 + st.H*area.h
		return []pt{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}, {x0, y0}}
	case "circle":
		cx, cy := normPoint(area, Point{X: st.X, Y: st.Y})
		rx := st.W * area.w
		ry := st.H * area.h
		pts := []pt{}
		for i := 0; i <= 72; i++ {
			a := float64(i) / 72 * 2 * math.Pi
			pts = append(pts, pt{cx + rx*math.Cos(a), cy + ry*math.Sin(a)})
		}
		return pts
	default:
		return nil
	}
}

// partialPolyline reveals a polyline by cumulative length (whiteboard feel).
func partialPolyline(pts []pt, progress float64) []pt {
	if progress <= 0 || len(pts) < 2 {
		return nil
	}
	if progress >= 1 {
		return pts
	}
	lengths := make([]float64, len(pts))
	total := 0.0
	for i := 1; i < len(pts); i++ {
		d := math.Hypot(pts[i].x-pts[i-1].x, pts[i].y-pts[i-1].y)
		total += d
		lengths[i] = total
	}
	if total <= 0 {
		return pts[:1]
	}
	target := total * progress
	out := []pt{pts[0]}
	for i := 1; i < len(pts); i++ {
		if lengths[i] <= target {
			out = append(out, pts[i])
			continue
		}
		seg := lengths[i] - lengths[i-1]
		if seg <= 0 {
			break
		}
		t := (target - lengths[i-1]) / seg
		out = append(out, pt{
			pts[i-1].x + (pts[i].x-pts[i-1].x)*t,
			pts[i-1].y + (pts[i].y-pts[i-1].y)*t,
		})
		break
	}
	return out
}
