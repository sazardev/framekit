package framekit

import "math"

// drawFormula is a minimal, typography-first explanation scene: one large
// equation, one quiet caption, then a small ordered derivation.
func drawFormula(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Formula == nil || sc.Formula.Expr == "" {
		return
	}
	l := p.lay
	W, H, S := float64(l.W), float64(l.H), l.Scale
	cx := W / 2

	if sc.Title != "" {
		p.textAnchored(sc.Title, cx, H*0.16, "h1-bold", "fg", fade)
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, fade*0.85))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawLine(cx-50*S, H*0.25, cx+50*S, H*0.25)
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
		// Hero equation: the single focal point, much larger than body text.
		face := "h1"
		drawMathFace(p, visible, cx-mathWidthFace(p, visible, face)/2, H*0.38, a, face)
	}

	if sc.Formula.Caption != "" {
		ca := fade * EasedReveal(spec, lt, sc.Formula.At+0.8)
		if ca > 0.01 {
			p.textAnchored(sc.Formula.Caption, cx, H*0.51, "normal", "dim", ca)
		}
	}

	if len(sc.Formula.Steps) > 0 {
		p.line(W*0.16, H*0.585, W*0.84, H*0.585, "faint", fade*0.7)
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
		y := H*0.64 + float64(i)*H*0.105
		x := W * 0.18
		p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, sa))
		p.ctx.DrawRectangle(x, y-16*S, 9*S, 9*S)
		p.ctx.Fill()
		p.text("0"+intStr(i+1), x+24*S, y, "small", "faint", sa*0.9)
		drawMathFace(p, parseMath(st.Expr), x+96*S, y, sa, "h2")
		if st.Note != "" {
			p.text(st.Note, x+96*S, y+40*S, "small", "dim", sa*0.9)
		}
	}
}
