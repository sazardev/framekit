package framekit

import (
	"strings"

	"github.com/fogleman/gg"
)

// drawImage renders a local image with cover/contain fit and optional slow
// push/pan. The frame stays flat: no shadows, no rounded corners.
func drawImage(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Image == nil || sc.Image.Path == "" {
		return
	}
	src, err := gg.LoadImage(sc.Image.Path)
	if err != nil {
		return // linter/security already reported the asset; never crash a render
	}
	l, S := p.lay, p.lay.Scale
	pad := 40 * S
	tw := float64(l.FrameX1-l.FrameX0) - 2*pad
	th := float64(l.FrameY1-l.FrameY0) - 2*pad - 40*S
	progress := clamp01(lt / maxFloat(sc.Duration, 0.1))
	zoom := 1 + sc.Image.Zoom*progress
	pan := sc.Image.Pan * progress * tw * 0.4

	sx, sy := float64(src.Bounds().Dx()), float64(src.Bounds().Dy())
	if sx <= 0 || sy <= 0 {
		return
	}
	scale := tw / sx
	if strings.EqualFold(sc.Image.Fit, "contain") {
		scale = minFloat(tw/sx, th/sy)
	} else {
		scale = maxFloat(tw/sx, th/sy) // cover
	}
	scale *= zoom
	w, h := sx*scale, sy*scale
	cx := float64(l.FrameX0+l.FrameX1)/2 + pan
	cy := float64(l.FrameY0+l.FrameY1)/2 - 10*S
	x, y := cx-w/2, cy-h/2

	// Clip to the inner frame so zoom/pan never spills over the chrome.
	p.ctx.Push()
	p.ctx.DrawRectangle(float64(l.FrameX0)+pad/2, float64(l.FrameY0)+50*S,
		tw+pad, th)
	p.ctx.Clip()
	if reveal(lt, sc.Image.At, spec.Effects.Fade) > 0.01 {
		p.ctx.DrawImage(src, int(x), int(y))
	}
	p.ctx.Pop()
	p.ctx.ResetClip() // gg v1.3 Pop() keeps the mask; clear it explicitly

	if sc.Image.Border {
		p.ctx.SetColor(p.theme.Color("faint"))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawRectangle(x+0.5, y+0.5, w-1, h-1)
		p.ctx.Stroke()
	}
	if sc.Title != "" {
		p.ctx.SetColor(Mix(p.theme.Color("bg"), p.theme.Color("bg"), 0.25))
		p.ctx.DrawRectangle(float64(l.FrameX0)+1, float64(l.FrameY1)-90*S,
			float64(l.FrameX1-l.FrameX0-2), 89*S)
		p.ctx.Fill()
		p.text(sc.Title, float64(l.ContentX), float64(l.FrameY1)-40*S, "h2-bold", "fg", fade)
	}
	if sc.Subtitle != "" {
		p.text(sc.Subtitle, float64(l.ContentX), float64(l.FrameY1)-18*S, "small", "dim", fade)
	}
}

// drawSlide is a presentation slide: large title, accent rule, square bullets.
func drawSlide(p *painter, spec Spec, sc Scene, lt, fade float64) {
	l, S := p.lay, p.lay.Scale
	x := float64(l.ContentX)
	y := float64(l.FrameY0) + 110*S
	accent := sc.Accent
	if accent == "" {
		accent = "green"
	}

	if sc.Title != "" {
		p.text(sc.Title, x, y, "hero-bold", "fg", fade)
	}
	ruleY := y + 34*S
	p.ctx.SetColor(Mix(p.theme.Color(accent), p.bg, fade))
	p.ctx.SetLineWidth(2)
	p.ctx.DrawLine(x, ruleY, x+180*S, ruleY)
	p.ctx.Stroke()

	base := ruleY + 80*S
	gap := 62 * S
	if l.Portrait {
		gap = 76 * S
	}
	stagger := spec.Effects.Stagger
	for i, ln := range sc.Lines {
		at := ln.At
		if at == 0 {
			at = 0.5 + float64(i)*stagger
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		dy := spec.Effects.Slide * S * (1 - Ease(spec.Effects.Easing, (lt-at)/spec.Effects.Fade))
		yy := base + float64(i)*gap + dy
		p.ctx.SetColor(Mix(p.theme.Color(accent), p.bg, a))
		p.ctx.DrawRectangle(x, yy-16*S, 10*S, 10*S)
		p.ctx.Fill()
		p.text(ln.Text, x+30*S, yy, "h2", "fg", a)
	}
	if sc.Subtitle != "" {
		p.text(sc.Subtitle, x, float64(l.FrameY1)-40*S, "normal", "dim", fade)
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
