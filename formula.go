package framekit

import "math"

// drawFormula renders an equation, a plain-English caption and an ordered
// derivation. Layout is deliberately hierarchical: one hero equation, one
// quiet explanation, then a tidy numbered list of steps.
func drawFormula(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Formula == nil || sc.Formula.Expr == "" {
		return
	}
	l, S := p.lay, p.lay.Scale
	cx := float64(l.W) / 2
	frameTop := float64(l.FrameY0)

	if sc.Title != "" {
		p.textAnchored(sc.Title, cx, frameTop+72*S, "h2-bold", "fg", fade)
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, fade*0.9))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawLine(cx-70*S, frameTop+100*S, cx+70*S, frameTop+100*S)
		p.ctx.Stroke()
	}

	items := parseMath(sc.Formula.Expr)
	a := fade * EasedReveal(spec, lt, sc.Formula.At)
	if a > 0.01 {
		visible := items
		if sc.Formula.Animate && len(items) > 0 {
			progress := clamp01((lt - sc.Formula.At) / math.Max(spec.Effects.Fade*2.4, 0.1))
			n := int(math.Ceil(progress * float64(len(items))))
			if n < 1 {
				n = 1
			}
			if n > len(items) {
				n = len(items)
			}
			visible = items[:n]
		}
		// Hero equation: large, airy, single focal point.
		drawMathFace(p, visible, cx-mathWidthFace(p, visible, "h2")/2,
			frameTop+190*S, a, "h2")
	}

	if sc.Formula.Caption != "" {
		ca := fade * EasedReveal(spec, lt, sc.Formula.At+0.8)
		if ca > 0.01 {
			p.textAnchored(sc.Formula.Caption, cx, frameTop+250*S, "normal", "dim", ca)
		}
	}

	if len(sc.Formula.Steps) > 0 {
		p.line(float64(l.FrameX0)+80*S, frameTop+290*S,
			float64(l.FrameX1)-80*S, frameTop+290*S, "faint", fade*0.8)
	}

	for i, st := range sc.Formula.Steps {
		at := st.At
		if at == 0 {
			at = sc.Formula.At + 1.4 + float64(i)*1.4
		}
		sa := fade * EasedReveal(spec, lt, at)
		if sa <= 0.01 {
			continue
		}
		y := frameTop + (330+float64(i)*78)*S
		x := float64(l.FrameX0) + 80*S
		// ordered marker
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, sa))
		p.ctx.DrawRectangle(x, y-16*S, 8*S, 8*S)
		p.ctx.Fill()
		p.text("0"+intStr(i+1), x+20*S, y, "small", "faint", sa*0.9)
		// equation and its plain-English note stacked, left aligned
		drawMathAt(p, parseMath(st.Expr), x+80*S, y, sa)
		if st.Note != "" {
			p.text(st.Note, x+80*S, y+32*S, "small", "dim", sa*0.9)
		}
	}
}
