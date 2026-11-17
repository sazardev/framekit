package framekit

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/fogleman/gg"
)

type Severity string

const (
	SevError Severity = "error"
	SevWarn  Severity = "warn"
	SevInfo  Severity = "info"
)

type Finding struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Scene    int      `json:"scene,omitempty"`
}

type TextMetric struct {
	Text      string  `json:"text"`
	Width     float64 `json:"width"`
	Available float64 `json:"available"`
	Overflow  bool    `json:"overflow"`
}

type ContrastInfo struct {
	FG     float64 `json:"fg"`
	Dim    float64 `json:"dim"`
	Accent float64 `json:"accent"`
}

type SceneAnalysis struct {
	Index    int          `json:"index"`
	Kind     string       `json:"kind"`
	Name     string       `json:"name"`
	Start    float64      `json:"start"`
	End      float64      `json:"end"`
	Lines    []TextMetric `json:"lines"`
	Items    []TextMetric `json:"items"`
	Contrast ContrastInfo `json:"contrast"`
	Density  float64      `json:"density"`
}

type Analysis struct {
	Title       string          `json:"title"`
	Format      string          `json:"format"`
	Quality     string          `json:"quality"`
	FPS         int             `json:"fps"`
	Duration    float64         `json:"duration"`
	EstimatedMB float64         `json:"estimatedMB"`
	VideoKbps   float64         `json:"videoKbps"`
	AudioKbps   float64         `json:"audioKbps"`
	Scenes      []SceneAnalysis `json:"scenes"`
	Contrast    ContrastInfo    `json:"contrast"`
}

// Analyze measures a spec without rendering: real text metrics, WCAG contrast,
// pacing and an estimated output size for the chosen quality profile.
func Analyze(spec Spec, cfg Config, quality string) (Analysis, error) {
	if err := spec.Normalize(); err != nil {
		return Analysis{}, err
	}
	q, err := cfg.QualityFor(quality)
	if err != nil {
		return Analysis{}, err
	}
	lay := ComputeLayout(EffectiveSize(spec, q))
	faces, err := loadFaces(lay.Fonts(), spec.Fonts)
	if err != nil {
		return Analysis{}, err
	}
	measurer := gg.NewContext(1, 1)

	a := Analysis{
		Title: spec.Title, Format: spec.Format, Quality: cfg.QualityName(quality),
		FPS: q.FPS, Duration: spec.DurationSeconds(),
	}
	bg := spec.Theme.Color("bg")
	accentName := "green"
	a.Contrast = ContrastInfo{
		FG:     contrastRatio(spec.Theme.Color("fg"), bg),
		Dim:    contrastRatio(spec.Theme.Color("dim"), bg),
		Accent: contrastRatio(spec.Theme.Color(accentName), bg),
	}

	// Empirical model for flat text graphics with x264 animation tune.
	px := float64(lay.W) * float64(lay.H)
	fpsFactor := float64(q.FPS) / 30
	crfFactor := math.Pow(23.0/float64(maxInt(q.CRF, 1)), 1.6)
	a.VideoKbps = px / (1280 * 720) * fpsFactor * 160 * crfFactor
	a.AudioKbps = 0 // filled by caller if music enabled
	if spec.Music.Enabled {
		a.AudioKbps = parseBitrate(q.AudioBitrate) * 0.6
	}
	a.EstimatedMB = (a.VideoKbps + a.AudioKbps) * a.Duration / 8 / 1024

	start := 0.0
	for i, sc := range spec.Scenes {
		sa := SceneAnalysis{
			Index: i + 1, Kind: sc.Kind, Name: sc.Name,
			Start: start, End: start + sc.Duration,
		}
		ac := accentName
		if sc.Accent != "" {
			ac = sc.Accent
		}
		sa.Contrast = ContrastInfo{
			FG:     contrastRatio(spec.Theme.Color("fg"), bg),
			Dim:    contrastRatio(spec.Theme.Color("dim"), bg),
			Accent: contrastRatio(spec.Theme.Color(ac), bg),
		}
		chars := 0
		for _, ln := range sc.Lines {
			size := "code"
			if ln.Font == "bold" {
				size = "h2-bold"
			}
			if ln.Font == "huge" {
				size = "hero-bold"
			}
			face := faces[size]
			if face == nil {
				face = faces["normal"]
			}
			measurer.SetFontFace(face)
			w, _ := measurer.MeasureString(ln.Text)
			avail := float64(lay.FrameX1-60*int(lay.Scale)) - float64(lay.ContentX+int(42*lay.Scale))
			sa.Lines = append(sa.Lines, TextMetric{
				Text: ln.Text, Width: w, Available: avail, Overflow: w > avail,
			})
			chars += len([]rune(ln.Text))
		}
		for _, it := range sc.Items {
			nameFace := faces["code"]
			measurer.SetFontFace(nameFace)
			nw, _ := measurer.MeasureString(it.Name)
			noteAvail := float64(lay.FrameX1-60*int(lay.Scale)) - float64(lay.ContentX+int(400*lay.Scale))
			measurer.SetFontFace(faces["small"])
			noteW, _ := measurer.MeasureString(it.Note)
			sa.Items = append(sa.Items, TextMetric{
				Text: it.Name, Width: nw, Available: 340 * lay.Scale, Overflow: nw > 340*lay.Scale,
			})
			if it.Note != "" {
				sa.Items = append(sa.Items, TextMetric{
					Text: it.Note, Width: noteW, Available: noteAvail, Overflow: noteW > noteAvail,
				})
			}
			chars += len([]rune(it.Name)) + len([]rune(it.Note))
		}
		chars += len([]rune(sc.Title)) + len([]rune(sc.Subtitle))
		sa.Density = float64(chars) / math.Max(sc.Duration, 0.1)
		a.Scenes = append(a.Scenes, sa)
		start = sa.End
	}
	return a, nil
}

// Lint enforces the configured guardrails and returns structured findings.
func Lint(spec Spec, cfg Config, quality string, analysis Analysis) []Finding {
	var out []Finding
	add := func(sev Severity, code, msg string, scene int) {
		out = append(out, Finding{Severity: sev, Code: code, Message: msg, Scene: scene})
	}
	r := cfg.Rules

	if analysis.Duration < r.MinDuration {
		add(SevError, "DURATION_SHORT", fmt.Sprintf("duration %.1fs is below minimum %.1fs", analysis.Duration, r.MinDuration), 0)
	}
	if analysis.Duration > r.MaxDuration {
		add(SevError, "DURATION_LONG", fmt.Sprintf("duration %.1fs exceeds maximum %.1fs", analysis.Duration, r.MaxDuration), 0)
	}
	if r.Require60FPS && analysis.FPS < 60 {
		add(SevWarn, "FPS_BELOW_60", fmt.Sprintf("quality profile uses %d fps; policy targets 60 fps", analysis.FPS), 0)
	}
	if !contains(r.AllowedFormats, spec.Format) {
		add(SevError, "FORMAT_DENIED", "format "+spec.Format+" is not allowed", 0)
	}
	if spec.Theme != nil && len(r.AllowedThemes) > 0 && !contains(r.AllowedThemes, spec.Theme.Name) {
		if _, custom := ThemePresets[spec.Theme.Name]; !custom {
			add(SevWarn, "THEME_CUSTOM", "custom theme "+spec.Theme.Name+" is not in the allowlist", 0)
		}
	}
	if len(spec.Scenes) > r.MaxScenes {
		add(SevError, "TOO_MANY_SCENES", fmt.Sprintf("%d scenes > max %d", len(spec.Scenes), r.MaxScenes), 0)
	}
	if r.RequireBranding && strings.TrimSpace(spec.Branding.Name) == "" {
		add(SevError, "BRANDING_MISSING", "branding.name is required by policy", 0)
	}
	if r.RequireMusic && !spec.Music.Enabled {
		add(SevWarn, "MUSIC_MISSING", "music is disabled but policy requires it", 0)
	}
	if len(spec.Voices) > r.MaxVoices {
		add(SevError, "TOO_MANY_VOICES", fmt.Sprintf("%d voices > max %d", len(spec.Voices), r.MaxVoices), 0)
	}
	if analysis.EstimatedMB > r.MaxFileSizeMB {
		add(SevWarn, "SIZE_ESTIMATE", fmt.Sprintf("estimated %.2f MB > policy %.2f MB", analysis.EstimatedMB, r.MaxFileSizeMB), 0)
	}
	if analysis.Contrast.FG < r.MinContrast {
		add(SevWarn, "CONTRAST_FG", fmt.Sprintf("main text contrast %.2f:1 below %.2f:1", analysis.Contrast.FG, r.MinContrast), 0)
	}
	if analysis.Contrast.Dim < r.MinContrast {
		add(SevWarn, "CONTRAST_DIM", fmt.Sprintf("secondary text contrast %.2f:1 below %.2f:1", analysis.Contrast.Dim, r.MinContrast), 0)
	}

	for i, sc := range spec.Scenes {
		n := i + 1
		if r.RequireEyebrows && sc.Name == "" {
			add(SevWarn, "EYEBROW_MISSING", "scene has no name/eyebrow", n)
		}
		hasVisual := sc.Code != nil || sc.Diagram != nil || sc.Drawing != nil || sc.Image != nil || sc.Formula != nil
		if len(sc.Lines)+len(sc.Items) == 0 && sc.Title == "" && len(sc.Stats) == 0 && !hasVisual {
			add(SevWarn, "EMPTY_SCENE", "scene has no visible content", n)
		}
		if len(sc.Items) > r.MaxItemsPerScene {
			add(SevError, "ITEMS_OVERFLOW", fmt.Sprintf("%d items > max %d", len(sc.Items), r.MaxItemsPerScene), n)
		}
		for _, ln := range sc.Lines {
			if len([]rune(ln.Text)) > r.MaxLineChars {
				add(SevWarn, "LINE_TOO_LONG", fmt.Sprintf("line exceeds %d chars: %q", r.MaxLineChars, truncate(ln.Text, 40)), n)
			}
			for _, w := range r.ForbiddenWords {
				if strings.Contains(strings.ToLower(ln.Text), strings.ToLower(w)) {
					add(SevError, "FORBIDDEN_WORD", "line contains forbidden word: "+w, n)
				}
			}
		}
		for _, it := range sc.Items {
			if len([]rune(it.Note)) > r.MaxNoteChars {
				add(SevWarn, "NOTE_TOO_LONG", fmt.Sprintf("note exceeds %d chars: %q", r.MaxNoteChars, truncate(it.Note, 40)), n)
			}
			if it.At+spec.Effects.Fade > sc.Duration {
				add(SevWarn, "AT_AFTER_END", "item appears too late to finish: "+it.Name, n)
			}
		}
		for _, st := range sc.Stats {
			if st.At+spec.Effects.Fade > sc.Duration {
				add(SevWarn, "AT_AFTER_END", "stat appears too late: "+st.Value, n)
			}
		}

		// Tutorial-specific checks.
		if sc.Code != nil {
			if strings.TrimSpace(sc.Code.Content) == "" {
				add(SevWarn, "CODE_EMPTY", "code scene has no content", n)
			}
			if len(strings.Split(sc.Code.Content, "\n")) > 16 {
				add(SevWarn, "CODE_TOO_TALL", "more than 16 lines will overflow the frame", n)
			}
			if _, ok := langKeywords[strings.ToLower(sc.Code.Language)]; !ok {
				alt := map[string]bool{"ts": true, "js": true, "javascript": true, "py": true, "rs": true, "sh": true}
				if !alt[strings.ToLower(sc.Code.Language)] {
					add(SevInfo, "CODE_LANG_FALLBACK", "unknown language "+sc.Code.Language+", using Go highlighting", n)
				}
			}
		}
		if sc.Diagram != nil {
			ids := map[string]bool{}
			for _, node := range sc.Diagram.Nodes {
				if node.ID == "" {
					add(SevError, "DIAGRAM_NODE_ID", "diagram node needs an id", n)
				}
				ids[node.ID] = true
			}
			for _, e := range sc.Diagram.Edges {
				if !ids[e.From] || !ids[e.To] {
					add(SevError, "DIAGRAM_BAD_EDGE", "edge references missing node: "+e.From+" -> "+e.To, n)
				}
			}
			if len(sc.Diagram.Nodes) > 9 {
				add(SevWarn, "DIAGRAM_CROWDED", "more than 9 nodes will be hard to read", n)
			}
		}
		if sc.Formula != nil {
			if !mathValid(sc.Formula.Expr) {
				add(SevError, "FORMULA_BRACES", "unbalanced braces in formula", n)
			}
			for _, cmd := range mathHasUnknownCommand(sc.Formula.Expr) {
				add(SevWarn, "FORMULA_UNKNOWN_CMD", "unknown math command: "+cmd, n)
			}
			if len(sc.Formula.Steps) > 3 {
				add(SevWarn, "FORMULA_MANY_STEPS", "more than 3 steps will overflow the frame", n)
			}
			for j, st := range sc.Formula.Steps {
				if !mathValid(st.Expr) {
					add(SevError, "FORMULA_BRACES", "unbalanced braces in step "+intStr(j+1), n)
				}
				at := st.At
				if at == 0 {
					at = sc.Formula.At + 1.4 + float64(j)*1.2
				}
				if at+spec.Effects.Fade > sc.Duration {
					add(SevWarn, "FORMULA_AFTER_END", "step appears too late to finish", n)
				}
			}
		}
		if sc.Drawing != nil {
			for j, st := range sc.Drawing.Strokes {
				switch st.Kind {
				case "line", "arrow", "path":
					if len(st.Points) < 2 {
						add(SevError, "DRAW_BAD_STROKE", fmt.Sprintf("stroke %d needs at least 2 points", j+1), n)
					}
				case "rect", "circle":
					if st.W <= 0 || st.H <= 0 {
						add(SevError, "DRAW_BAD_STROKE", fmt.Sprintf("stroke %d needs w and h", j+1), n)
					}
				default:
					add(SevError, "DRAW_BAD_KIND", "unknown stroke kind: "+st.Kind, n)
				}
				if st.At+st.Duration > sc.Duration {
					add(SevWarn, "DRAW_AFTER_END", fmt.Sprintf("stroke %d finishes after scene ends", j+1), n)
				}
			}
		}

		if i < len(analysis.Scenes) {
			for _, m := range analysis.Scenes[i].Lines {
				if m.Overflow {
					add(SevWarn, "TEXT_OVERFLOW", fmt.Sprintf("text may overflow: %q", truncate(m.Text, 36)), n)
				}
			}
			for _, m := range analysis.Scenes[i].Items {
				if m.Overflow {
					add(SevWarn, "TEXT_OVERFLOW", fmt.Sprintf("text may overflow: %q", truncate(m.Text, 36)), n)
				}
			}
			if analysis.Scenes[i].Density > 9 {
				add(SevInfo, "DENSE_SCENE", fmt.Sprintf("high text density %.1f chars/s", analysis.Scenes[i].Density), n)
			}
		}
	}
	return out
}

// SecurityAudit checks asset paths and remote calls against the policy.
func SecurityAudit(spec Spec, cfg Config) []Finding {
	var out []Finding
	add := func(code, msg string) {
		out = append(out, Finding{Severity: SevError, Code: code, Message: msg})
	}
	paths := []string{}
	if spec.Fonts.Regular != "" {
		paths = append(paths, spec.Fonts.Regular)
	}
	if spec.Fonts.Bold != "" {
		paths = append(paths, spec.Fonts.Bold)
	}
	for _, sc := range spec.Scenes {
		if sc.Image != nil && sc.Image.Path != "" {
			paths = append(paths, sc.Image.Path)
		}
	}
	for _, v := range spec.Voices {
		if v.File != "" {
			paths = append(paths, v.File)
		}
	}
	for _, p := range paths {
		if err := CheckAssetPath(p, cfg.Security); err != nil {
			add("ASSET_DENIED", err.Error())
		}
	}
	for _, v := range spec.Voices {
		if v.Provider == "cloudflare" && !cfg.Security.AllowRemoteFetch {
			add("REMOTE_DISABLED", "remote voice fetch is disabled in config; use a local voice or enable AllowRemoteFetch")
		}
	}
	return out
}

// --- helpers ---

func contrastRatio(a, b color.RGBA) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(c color.RGBA) float64 {
	f := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*f(c.R) + 0.7152*f(c.G) + 0.0722*f(c.B)
}

func parseBitrate(s string) float64 {
	var v float64
	var unit string
	if _, err := fmt.Sscanf(s, "%f%s", &v, &unit); err != nil {
		return 128
	}
	switch strings.ToLower(unit) {
	case "m", "mbps":
		return v * 1000
	default:
		return v
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
