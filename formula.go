package framekit

import "math"

// drawFormula renders an equation, its plain-English caption and step-by-step
// derivations. It is the "explain theory visually" scene.
func drawFormula(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Formula == nil || sc.Formula.Expr == "" {
		return
	}
	l, S := p.lay, p.lay.Scale
	cx := float64(l.W) / 2
	if sc.Title != "" {
		p.textAnchored(sc.Title, cx, float64(l.FrameY0)+72*S, "h2-bold", "fg", fade)
	}

	items := parseMath(sc.Formula.Expr)
	a := fade * EasedReveal(spec, lt, sc.Formula.At)
	if a > 0.01 {
		visible := items
		if sc.Formula.Animate && len(items) > 0 {
			progress := clamp01((lt - sc.Formula.At) / math.Max(spec.Effects.Fade*2.2, 0.1))
			n := int(math.Ceil(progress * float64(len(items))))
			if n < 1 {
				n = 1
			}
			if n > len(items) {
				n = len(items)
			}
			visible = items[:n]
		}
		// accent rule above the equation
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, a))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawLine(cx-220*S, float64(l.CenterY)-86*S, cx+220*S, float64(l.CenterY)-86*S)
		p.ctx.Stroke()
		drawMath(p, visible, cx, float64(l.CenterY)-20*S, a)
	}

	if sc.Formula.Caption != "" {
		ca := fade * EasedReveal(spec, lt, sc.Formula.At+0.7)
		if ca > 0.01 {
			p.textAnchored(sc.Formula.Caption, cx, float64(l.CenterY)+48*S, "normal", "aqua", ca)
		}
	}

	base := float64(l.CenterY) + 105*S
	for i, st := range sc.Formula.Steps {
		at := st.At
		if at == 0 {
			at = sc.Formula.At + 1.4 + float64(i)*1.2
		}
		sa := fade * EasedReveal(spec, lt, at)
		if sa <= 0.01 {
			continue
		}
		y := base + float64(i)*66*S
		p.ctx.SetColor(Mix(p.theme.Color("soft"), p.bg, 0.5))
		p.ctx.DrawRectangle(float64(l.FrameX0)+60*S, y-26*S, float64(l.FrameX1-l.FrameX0)-120*S, 54*S)
		p.ctx.Fill()
		drawMathAt(p, parseMath(st.Expr), float64(l.FrameX0)+90*S, y, sa)
		if st.Note != "" {
			p.text(st.Note, float64(l.FrameX1)-300*S, y+8*S, "small", "dim", sa)
		}
	}
}
