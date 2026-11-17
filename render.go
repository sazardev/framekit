package framekit

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type painter struct {
	ctx   *gg.Context
	lay   Layout
	theme Theme
	faces map[string]font.Face
	bg    color.RGBA
	f     FontRamp
}

// Render produces the final MP4 in out: video + optional synthesized music +
// optional voices, muxed by ffmpeg. Everything local unless a voice provider
// is explicitly configured.
func Render(spec Spec, out string) error {
	return RenderWithConfig(spec, out, DefaultConfig(), "")
}

// RenderWithConfig renders using a quality profile from the guardrail config.
func RenderWithConfig(spec Spec, out string, cfg Config, quality string) error {
	SanitizeSpec(&spec)
	if err := spec.Normalize(); err != nil {
		return err
	}
	q, err := cfg.QualityFor(quality)
	if err != nil {
		return err
	}
	if out == "" {
		out = spec.Output
	}
	if out == "" {
		return fmt.Errorf("no output path")
	}
	ffmpeg := spec.FFmpeg
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	if _, err := exec.LookPath(ffmpeg); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	tmp, err := os.MkdirTemp("", "framekit-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	silent := filepath.Join(tmp, "silent.mp4")
	if err := renderSilent(spec, ffmpeg, silent, q); err != nil {
		return err
	}

	audios := []string{}
	if spec.Music.Enabled {
		wav, err := SynthMusic(spec.DurationSeconds(), spec.Music)
		if err != nil {
			return err
		}
		p := filepath.Join(tmp, "music.wav")
		if err := os.WriteFile(p, wav, 0o644); err != nil {
			return err
		}
		audios = append(audios, p)
	}
	for i, v := range spec.Voices {
		p, err := resolveVoice(v)
		if err != nil {
			return fmt.Errorf("voice %d: %w", i, err)
		}
		if p != "" {
			audios = append(audios, p)
		}
	}

	if len(audios) == 0 {
		return os.Rename(silent, out)
	}
	return mux(ffmpeg, silent, audios, spec, out, q)
}

func renderSilent(spec Spec, ffmpeg, out string, q Quality) error {
	fps := q.FPS
	if fps <= 0 {
		fps = spec.FPS
	}
	size := EffectiveSize(spec, q)
	lay := ComputeLayout(size)
	faces, err := loadFaces(lay.Fonts(), spec.Fonts)
	if err != nil {
		return err
	}
	cmd := exec.Command(ffmpeg, "-y", "-loglevel", "error",
		"-f", "rawvideo", "-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", lay.W, lay.H),
		"-r", fmt.Sprint(fps), "-i", "-",
		"-an", "-c:v", "libx264",
		"-preset", q.Preset, "-crf", fmt.Sprint(q.CRF),
		"-tune", q.Tune, "-profile:v", q.Profile, "-level", q.Level,
		"-maxrate", q.MaxRate, "-bufsize", q.BufSize,
		"-pix_fmt", "yuv420p", "-movflags", "+faststart", out)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	total := int(math.Ceil(spec.DurationSeconds() * float64(fps)))
	p := &painter{
		ctx:   gg.NewContext(lay.W, lay.H),
		lay:   lay,
		theme: *spec.Theme,
		faces: faces,
		bg:    spec.Theme.Color("bg"),
		f:     lay.Fonts(),
	}
	sceneStart, sceneIdx := 0.0, 0
	for i := 0; i < total; i++ {
		t := float64(i) / float64(fps)
		for sceneIdx < len(spec.Scenes)-1 && t >= sceneStart+spec.Scenes[sceneIdx].Duration {
			sceneStart += spec.Scenes[sceneIdx].Duration
			sceneIdx++
		}
		drawFrame(p, spec, sceneIdx, t-sceneStart, t)
		img := p.ctx.Image().(*image.RGBA)
		if _, err := stdin.Write(img.Pix); err != nil {
			msg, _ := io.ReadAll(stderr)
			return fmt.Errorf("write frame %d: %w: %s", i, err, strings.TrimSpace(string(msg)))
		}
	}
	_ = stdin.Close()
	msg, _ := io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, strings.TrimSpace(string(msg)))
	}
	return nil
}

func loadFaces(r FontRamp, cfg FontConfig) (map[string]font.Face, error) {
	reg, err := FontBytesFrom(cfg.Regular, "regular")
	if err != nil {
		return nil, err
	}
	bold, err := FontBytesFrom(cfg.Bold, "bold")
	if err != nil {
		return nil, err
	}
	fr, err := opentype.Parse(reg)
	if err != nil {
		return nil, err
	}
	fb, err := opentype.Parse(bold)
	if err != nil {
		return nil, err
	}
	sizes := map[string]float64{
		"small": r.Small, "normal": r.Normal, "code": r.Code,
		"h2": r.H2, "h1": r.H1, "hero": r.Hero,
	}
	out := make(map[string]font.Face, len(sizes)*2)
	for name, size := range sizes {
		if size < 6 {
			size = 6
		}
		a, err := opentype.NewFace(fr, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			return nil, err
		}
		b, err := opentype.NewFace(fb, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			return nil, err
		}
		out[name] = a
		out[name+"-bold"] = b
	}
	return out, nil
}

func drawFrame(p *painter, spec Spec, sceneIdx int, localT, totalT float64) {
	p.ctx.SetColor(p.bg)
	p.ctx.DrawRectangle(0, 0, float64(p.lay.W), float64(p.lay.H))
	p.ctx.Fill()

	if spec.Effects.Chrome {
		drawChrome(p, spec, sceneIdx)
	}

	sc := spec.Scenes[sceneIdx]
	fade := EasedReveal(spec, localT, 0) * EasedExit(spec, localT, sc.Duration)

	if spec.Effects.Scanline && localT < 1.2 {
		drawScanline(p, localT/1.2)
	}

	switch strings.ToLower(sc.Kind) {
	case "title":
		drawTitle(p, spec, sc, fade)
	case "stats":
		drawStats(p, spec, sc, localT, fade)
	case "credits":
		drawCredits(p, spec, sc, fade)
	case "image":
		drawImage(p, spec, sc, localT, fade)
	case "slide":
		drawSlide(p, spec, sc, localT, fade)
	case "code":
		if sc.Code != nil {
			drawCode(p, spec, sc, localT, fade)
		} else {
			drawTerminal(p, spec, sc, localT, fade, true)
		}
	case "diagram":
		drawDiagram(p, spec, sc, localT, fade)
	case "draw":
		drawDrawing(p, spec, sc, localT, fade)
	case "formula", "math":
		drawFormula(p, spec, sc, localT, fade)
	default:
		drawTerminal(p, spec, sc, localT, fade, false)
	}

	if spec.Effects.Transition == "wipe" {
		drawWipe(p, localT, sc.Duration)
	}

	if spec.Effects.Progress {
		drawProgress(p, spec, totalT)
	}
}

// drawScanline is a single flat 1px accent sweeping down at scene start.
func drawScanline(p *painter, progress float64) {
	y := float64(p.lay.FrameY0) + float64(p.lay.FrameY1-p.lay.FrameY0)*clamp01(progress)
	p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, 0.35))
	p.ctx.SetLineWidth(1)
	p.ctx.DrawLine(float64(p.lay.FrameX0), y, float64(p.lay.FrameX1), y)
	p.ctx.Stroke()
}

// drawWipe reveals the scene with a moving vertical accent line.
func drawWipe(p *painter, local, duration float64) {
	if local > 0.6 {
		return
	}
	x := float64(p.lay.FrameX0) + float64(p.lay.FrameX1-p.lay.FrameX0)*Ease("easeInOut", local/0.6)
	p.ctx.SetColor(p.theme.Color("green"))
	p.ctx.SetLineWidth(2)
	p.ctx.DrawLine(x, float64(p.lay.FrameY0), x, float64(p.lay.FrameY1))
	p.ctx.Stroke()
}

func drawChrome(p *painter, spec Spec, sceneIdx int) {
	l, th := p.lay, p.theme
	S := l.Scale
	// header
	p.ctx.SetColor(mustHex(th.Green))
	p.ctx.DrawRectangle(float64(l.Margin), float64(l.HeaderY), 14*S, 14*S)
	p.ctx.Fill()
	p.text(spec.Branding.Name+"  /  "+spec.Branding.Subtitle, float64(l.Margin)+24*S, float64(l.HeaderY)+13*S, "normal", "fg", 1)
	meta := spec.Branding.Meta
	ax := p.textWidth(meta, "small")
	p.text(meta, float64(l.W-l.Margin)-ax, float64(l.HeaderY)+12*S, "small", "faint", 1)
	p.line(float64(l.Margin), float64(l.DividerY), float64(l.W-l.Margin), float64(l.DividerY), "faint", 1)

	// terminal window
	p.ctx.SetColor(mustHex(th.Faint))
	p.ctx.SetLineWidth(1)
	p.ctx.DrawRectangle(float64(l.FrameX0)+0.5, float64(l.FrameY0)+0.5,
		float64(l.FrameX1-l.FrameX0-1), float64(l.FrameY1-l.FrameY0-1))
	p.ctx.Stroke()
	p.ctx.SetColor(mustHex(th.Deep))
	p.ctx.DrawRectangle(float64(l.FrameX0)+1, float64(l.FrameY0)+1,
		float64(l.FrameX1-l.FrameX0-2), 45*S)
	p.ctx.Fill()
	p.line(float64(l.FrameX0), float64(l.FrameY0)+46*S, float64(l.FrameX1), float64(l.FrameY0)+46*S, "faint", 1)
	for i, c := range []string{"yellow", "aqua", "green"} {
		p.ctx.SetColor(th.Color(c))
		p.ctx.DrawRectangle(float64(l.FrameX0)+22*S+float64(i)*18*S, float64(l.FrameY0)+18*S, 10*S, 10*S)
		p.ctx.Fill()
	}
	user := strings.ToLower(strings.Fields(spec.Branding.Name)[0])
	if user == "" {
		user = "user"
	}
	title := user + "@" + strings.ToLower(strings.ReplaceAll(spec.Branding.Name, " ", "")) + ": ~/" + spec.Branding.WindowTitle
	tw := p.textWidth(title, "small")
	p.text(title, float64(l.FrameX0+l.FrameX1)/2-tw/2, float64(l.FrameY0)+18*S+12*S, "small", "dim", 1)

	if sc := spec.Scenes[sceneIdx]; sc.Name != "" {
		nw := p.textWidth(sc.Name, "small")
		p.text(sc.Name, float64(l.FrameX1)-28*S-nw, float64(l.FrameY0)+76*S, "small", "faint", 0.85)
	}
}

func drawProgress(p *painter, spec Spec, t float64) {
	total := spec.DurationSeconds()
	if total <= 0 {
		return
	}
	l, S := p.lay, p.lay.Scale
	y := float64(l.H) - 52*S
	p.line(float64(l.Margin), y, float64(l.W-l.Margin), y, "faint", 1)
	prog := Ease(spec.Effects.Easing, t/total)
	p.ctx.SetColor(mustHex(p.theme.Green))
	p.ctx.SetLineWidth(2)
	p.ctx.DrawLine(float64(l.Margin), y, float64(l.Margin)+float64(l.W-2*l.Margin)*prog, y)
	p.ctx.Stroke()
	left := spec.Branding.FooterLeft
	p.text(left, float64(l.Margin), float64(l.H)-40*S, "small", "faint", 1)
	tag := spec.Branding.FooterRight
	tw := p.textWidth(tag, "small")
	p.text(tag, float64(l.W-l.Margin)-tw, float64(l.H)-40*S, "small", "faint", 1)
}

func drawTerminal(p *painter, spec Spec, sc Scene, lt float64, fade float64, code bool) {
	l, S := p.lay, p.lay.Scale
	y := float64(l.CenterY) - 90*S
	x := float64(l.ContentX)

	if sc.Command != "" {
		cmd := sc.Command
		if spec.Effects.Typewriter {
			cmd = typed(cmd, lt, 0.4, 28)
		}
		p.text("$ ", x, y, "code", "green", fade)
		p.text(cmd, x+42*S, y, "code", "fg", fade)
		if spec.Effects.Cursor && len(cmd) == len(sc.Command) && int(lt*2.2)%2 == 0 {
			cx := x + 42*S + p.textWidth(cmd, "code") + 3
			p.ctx.SetColor(Mix(mustHex(p.theme.FG), p.bg, fade))
			p.ctx.DrawRectangle(cx, y-22*S, 12*S, 27*S)
			p.ctx.Fill()
		}
	}

	lineH := 49 * S
	if code {
		lineH = 36 * S
	}
	base := y + 70*S
	stagger := spec.Effects.Stagger
	accent := sc.Accent
	if accent == "" {
		accent = "aqua"
	}
	for i, ln := range sc.Lines {
		at := ln.At
		if at == 0 {
			at = 0.8 + float64(i)*stagger
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		s := ln.Text
		if spec.Effects.Typewriter {
			s = typed(s, lt, at, 34)
		}
		dy := spec.Effects.Slide * S * (1 - Ease(spec.Effects.Easing, (lt-at)/spec.Effects.Fade))
		size := "code"
		if ln.Font == "bold" {
			size = "h2-bold"
		} else if ln.Font == "huge" {
			size = "hero-bold"
		}
		colorName := ln.Color
		if colorName == "" {
			colorName = "fg"
		}
		p.text(s, x+42*S, base+float64(i)*lineH+dy, size, colorName, a)
	}

	itemBase := base + float64(len(sc.Lines))*lineH
	for i, it := range sc.Items {
		at := it.At
		if at == 0 {
			at = 0.8 + float64(i)*stagger
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		dy := spec.Effects.Slide * S * (1 - Ease(spec.Effects.Easing, (lt-at)/spec.Effects.Fade))
		yy := itemBase + float64(i)*lineH + dy
		p.text(it.Name, x+42*S, yy, "code", accent, a)
		if it.Note != "" {
			p.text(it.Note, x+minFloat(400*S, 430*S), yy, "small", "dim", a)
		}
	}
}

func drawTitle(p *painter, spec Spec, sc Scene, fade float64) {
	l, S := p.lay, p.lay.Scale
	cx := float64(l.W) / 2
	cy := float64(l.CenterY)
	accent := sc.Accent
	if accent == "" {
		accent = "green"
	}
	if sc.Title != "" {
		if spec.Effects.Tracking > 0 {
			p.textTracked(sc.Title, cx, cy-40*S, "hero-bold", "fg", fade, spec.Effects.Tracking*S)
		} else {
			p.textAnchored(sc.Title, cx, cy-40*S, "hero-bold", "fg", fade)
		}
	}
	if sc.Subtitle != "" {
		p.ctx.SetColor(Mix(p.theme.Color(accent), p.bg, fade))
		p.ctx.SetLineWidth(1)
		p.ctx.DrawLine(cx-200*S, cy+6*S, cx+200*S, cy+6*S)
		p.ctx.Stroke()
		p.textAnchored(sc.Subtitle, cx, cy+70*S, "h2", accent, fade)
	}
}

func drawCredits(p *painter, spec Spec, sc Scene, fade float64) {
	drawTitle(p, spec, sc, fade)
	l, S := p.lay, p.lay.Scale
	if len(sc.Lines) > 0 {
		p.textAnchored(sc.Lines[0].Text, float64(l.W)/2, float64(l.CenterY)+130*S, "normal", "dim", fade)
	}
}

func drawStats(p *painter, spec Spec, sc Scene, lt float64, fade float64) {
	l, S := p.lay, p.lay.Scale
	cx := float64(l.W) / 2
	cy := float64(l.CenterY)
	gap := 320 * S
	if l.Portrait {
		gap = 150 * S
	}
	startX := cx - gap*float64(len(sc.Stats)-1)/2
	accent := sc.Accent
	if accent == "" {
		accent = "green"
	}
	stagger := spec.Effects.Stagger
	for i, st := range sc.Stats {
		at := st.At
		if at == 0 {
			at = 0.4 + float64(i)*stagger
		}
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		yy := cy - 10*S
		xx := startX + gap*float64(i)
		if l.Portrait {
			yy = cy - 180*S + float64(i)*150*S
			xx = cx
		}
		p.textAnchored(st.Value, xx, yy, "hero-bold", accent, a)
		p.textAnchored(st.Label, xx, yy+70*S, "small", "dim", a)
	}
}

// --- small drawing helpers ---

func (p *painter) text(s string, x, y float64, size, colorName string, alpha float64) {
	if alpha <= 0.01 || s == "" {
		return
	}
	face := p.faces[size]
	if face == nil {
		face = p.faces["normal"]
	}
	p.ctx.SetFontFace(face)
	c := p.theme.Color(colorName)
	if alpha < 0.999 {
		c = Mix(c, p.bg, alpha)
	}
	p.ctx.SetColor(c)
	p.ctx.DrawString(s, x, y)
	p.ctx.Fill()
}

func (p *painter) textAnchored(s string, x, y float64, size, colorName string, alpha float64) {
	if alpha <= 0.01 || s == "" {
		return
	}
	face := p.faces[size]
	if face == nil {
		face = p.faces["normal"]
	}
	p.ctx.SetFontFace(face)
	c := p.theme.Color(colorName)
	if alpha < 0.999 {
		c = Mix(c, p.bg, alpha)
	}
	p.ctx.SetColor(c)
	p.ctx.DrawStringAnchored(s, x, y, 0.5, 0.5)
	p.ctx.Fill()
}

// textTracked draws centered text with extra letter spacing (premium look).
func (p *painter) textTracked(s string, cx, y float64, size, colorName string, alpha, tracking float64) {
	face := p.faces[size]
	if face == nil {
		face = p.faces["normal"]
	}
	p.ctx.SetFontFace(face)
	c := p.theme.Color(colorName)
	if alpha < 0.999 {
		c = Mix(c, p.bg, alpha)
	}
	p.ctx.SetColor(c)
	runes := []rune(s)
	total := 0.0
	for _, r := range runes {
		w, _ := p.ctx.MeasureString(string(r))
		total += w + tracking
	}
	total -= tracking
	x := cx - total/2
	for _, r := range runes {
		ch := string(r)
		w, _ := p.ctx.MeasureString(ch)
		p.ctx.DrawString(ch, x, y)
		x += w + tracking
	}
	p.ctx.Fill()
}

func (p *painter) textWidth(s, size string) float64 {
	face := p.faces[size]
	if face == nil {
		face = p.faces["normal"]
	}
	p.ctx.SetFontFace(face)
	w, _ := p.ctx.MeasureString(s)
	return w
}

func (p *painter) line(x0, y0, x1, y1 float64, colorName string, alpha float64) {
	c := p.theme.Color(colorName)
	if alpha < 0.999 {
		c = Mix(c, p.bg, alpha)
	}
	p.ctx.SetColor(c)
	p.ctx.SetLineWidth(1)
	p.ctx.DrawLine(x0, y0, x1, y1)
	p.ctx.Stroke()
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func reveal(t, at, dur float64) float64 {
	if dur <= 0 {
		if t >= at {
			return 1
		}
		return 0
	}
	return clamp01((t - at) / dur)
}

func typed(s string, t, at, speed float64) string {
	if t <= at || speed <= 0 {
		return ""
	}
	n := int((t - at) * speed)
	if n >= len(s) {
		return s
	}
	return s[:n]
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
