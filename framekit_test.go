package framekit

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	s := Spec{
		Format: "portrait",
		Scenes: []Scene{
			{Kind: "terminal", Duration: 2},
			{Kind: "stats", Duration: 3},
		},
	}
	if err := s.Normalize(); err != nil {
		t.Fatal(err)
	}
	if s.FPS != 30 {
		t.Fatalf("fps = %d", s.FPS)
	}
	if s.DurationSeconds() != 5 {
		t.Fatalf("duration = %v", s.DurationSeconds())
	}
	if s.Size.W != 1080 || s.Size.H != 1920 {
		t.Fatalf("size = %+v", s.Size)
	}
	if s.Theme.Name != "gruvbox" {
		t.Fatalf("theme = %q", s.Theme.Name)
	}
}

func TestPresets(t *testing.T) {
	if got := Presets["landscape"]; got.W != 1920 || got.H != 1080 {
		t.Fatalf("landscape preset = %+v, want 1920x1080", got)
	}
	for _, name := range []string{"landscape", "landscape720", "square", "portrait"} {
		if _, ok := Presets[name]; !ok {
			t.Fatalf("missing preset %s", name)
		}
	}
	s := Spec{Format: "nope", Scenes: []Scene{{Duration: 1}}}
	if err := s.Normalize(); err == nil {
		t.Fatal("expected error for bad format")
	}
}

func TestThemeMergeAndColors(t *testing.T) {
	base := Gruvbox()
	custom := &Theme{Green: "#00ff00", Brand: "Custom"}
	merged := base.Merge(custom)
	if merged.Green != "#00ff00" || merged.Brand != "Custom" {
		t.Fatalf("merge failed: %+v", merged)
	}
	if merged.BG != base.BG {
		t.Fatal("merge overwrote BG")
	}
	c, err := ParseHex("#b8bb26")
	if err != nil || c.R != 0xb8 || c.G != 0xbb || c.B != 0x26 || c.A != 255 {
		t.Fatalf("ParseHex = %+v %v", c, err)
	}
}

func TestSynthMusic(t *testing.T) {
	cfg := Music{Enabled: true, Volume: 0.25}
	a, err := SynthMusic(4, cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SynthMusic(4, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("synth is not deterministic")
	}
	if string(a[:4]) != "RIFF" || string(a[8:12]) != "WAVE" {
		t.Fatal("bad WAV header")
	}
	want := 36 + 4*sampleRate*2
	got := int(binary.LittleEndian.Uint32(a[4:8]))
	if got != want {
		t.Fatalf("RIFF size = %d, want %d", got, want)
	}
}

func TestTypedAndReveal(t *testing.T) {
	if got := typed("hello", 0.5, 1, 10); got != "" {
		t.Fatalf("typed early = %q", got)
	}
	if got := typed("hello", 2, 1, 10); got != "hello" {
		t.Fatalf("typed done = %q", got)
	}
	if got := reveal(1, 0, 0.5); got != 1 {
		t.Fatalf("reveal done = %v", got)
	}
	if got := reveal(0.1, 0, 0.5); got <= 0 || got >= 1 {
		t.Fatalf("reveal mid = %v", got)
	}
}

func TestTemplates(t *testing.T) {
	cfg := DefaultConfig()
	for _, name := range []string{"intro", "reel", "pitch", "presentation", "quote", "image-showcase"} {
		s, err := RenderTemplate(name, map[string]string{}, cfg)
		if err != nil {
			t.Fatalf("template %s: %v", name, err)
		}
		if len(s.Scenes) == 0 {
			t.Fatalf("template %s has no scenes", name)
		}
	}
	s, err := RenderTemplate("quote", map[string]string{"quote": "Ship it."}, cfg)
	if err != nil || s.Scenes[0].Title != "Ship it." {
		t.Fatalf("variable substitution failed: %+v %v", s, err)
	}
}

func TestLintAndSecurity(t *testing.T) {
	cfg := DefaultConfig()
	s := Spec{
		Title:  "Test",
		Scenes: []Scene{{Kind: "terminal", Duration: 2, Lines: []Line{{Text: strings.Repeat("x", 80)}}}},
	}
	if err := s.Normalize(); err != nil {
		t.Fatal(err)
	}
	a, err := Analyze(s, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	findings := Lint(s, cfg, "", a)
	codes := map[string]bool{}
	for _, f := range findings {
		codes[f.Code] = true
	}
	if !codes["DURATION_SHORT"] || !codes["LINE_TOO_LONG"] {
		t.Fatalf("missing findings: %+v", codes)
	}
	sec := Spec{Fonts: FontConfig{Regular: "../outside.ttf"}}
	if got := SecurityAudit(sec, cfg); len(got) == 0 {
		t.Fatal("expected traversal finding")
	}
	if got := SanitizeText("hello\x00world\n\n\nnext"); got != "hello world\n\nnext" {
		t.Fatalf("sanitize = %q", got)
	}
}

func TestQualityProfiles(t *testing.T) {
	cfg := DefaultConfig()
	for _, n := range []string{"draft", "standard", "high", "social", "master"} {
		q, err := cfg.QualityFor(n)
		if err != nil {
			t.Fatal(err)
		}
		if q.FPS != 60 {
			t.Fatalf("profile %s = %d fps, want 60", n, q.FPS)
		}
	}
	spec := Spec{Format: "landscape", Size: &Size{W: 1280, H: 720}, Scenes: []Scene{{Duration: 1}}}
	if err := spec.Normalize(); err != nil {
		t.Fatal(err)
	}
	got := EffectiveSize(spec, QualityProfiles["draft"])
	if got.W != 640 || got.H != 360 {
		t.Fatalf("draft size = %+v", got)
	}
}

func TestTutorialScenes(t *testing.T) {
	cfg := DefaultConfig()
	tut, err := RenderTemplate("tutorial", map[string]string{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]bool{}
	for _, sc := range tut.Scenes {
		kinds[sc.Kind] = true
	}
	for _, k := range []string{"code", "diagram", "draw"} {
		if !kinds[k] {
			t.Fatalf("tutorial missing %s scene", k)
		}
	}
	if err := tut.Normalize(); err != nil {
		t.Fatal(err)
	}
	a, err := Analyze(tut, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Lint(tut, cfg, "", a) {
		if f.Severity == SevError {
			t.Fatalf("lint: %s %s", f.Code, f.Message)
		}
	}
	// Tokenizer smoke test.
	toks := tokenize(`func main() { ch := make(chan int) // hi`, "go")
	found := false
	for _, tk := range toks {
		if tk.kind == "keyword" && tk.text == "func" {
			found = true
		}
	}
	if !found {
		t.Fatalf("tokenizer missed func: %+v", toks)
	}
	// Lint must catch broken edges and strokes.
	bad := tut
	bad.Scenes = append([]Scene{}, tut.Scenes...)
	bad.Scenes[2].Diagram.Edges[0].To = "missing"
	if err := bad.Normalize(); err != nil {
		t.Fatal(err)
	}
	a2, _ := Analyze(bad, cfg, "")
	ok := false
	for _, f := range Lint(bad, cfg, "", a2) {
		if f.Code == "DIAGRAM_BAD_EDGE" {
			ok = true
		}
	}
	if !ok {
		t.Fatal("expected DIAGRAM_BAD_EDGE")
	}
}

func TestMath(t *testing.T) {
	if !mathValid(`\frac{a}{b} + x^2`) {
		t.Fatal("valid expression rejected")
	}
	if mathValid(`\frac{a}{b`) {
		t.Fatal("unbalanced braces accepted")
	}
	items := parseMath(`T(n) = O(n \log n)`)
	joined := mathPlain(`T(n) = O(n \log n)`)
	if !strings.Contains(joined, "O(n log n)") {
		t.Fatalf("command expansion failed: %q (%+v)", joined, items)
	}
	if got := flattenGroup(`\log_2 2`); !strings.Contains(got, "log₂") {
		t.Fatalf("flattenGroup = %q", got)
	}
	if u := mathHasUnknownCommand(`\foobar x`); len(u) != 1 || u[0] != "foobar" {
		t.Fatalf("unknown commands = %v", u)
	}
	ex, err := RenderTemplate("explainer", map[string]string{}, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, sc := range ex.Scenes {
		if sc.Kind == "formula" && sc.Formula != nil {
			found = true
		}
	}
	if !found {
		t.Fatal("explainer template has no formula scene")
	}
}

func TestEasing(t *testing.T) {
	if got := Ease("linear", 0.5); got != 0.5 {
		t.Fatalf("linear = %v", got)
	}
	for _, kind := range []string{"easeIn", "easeInOut", "easeOutCubic", "easeOutQuint", "spring"} {
		if got := Ease(kind, 1); math.Abs(got-1) > 1e-6 {
			t.Fatalf("%s(1) = %v", kind, got)
		}
	}
}

func TestThemePresets(t *testing.T) {
	if got := ThemeByName("catppuccin"); got.BG != "#1e1e2e" {
		t.Fatalf("catppuccin bg = %q", got.BG)
	}
	s := Spec{Theme: &Theme{Name: "nord"}, Scenes: []Scene{{Duration: 1}}}
	if err := s.Normalize(); err != nil {
		t.Fatal(err)
	}
	if s.Theme.BG != "#2e3440" || s.Branding.Name != "SazarDev" {
		t.Fatalf("theme/branding = %+v / %+v", s.Theme, s.Branding)
	}
}

func TestEmbeddedFonts(t *testing.T) {
	for _, w := range []string{"regular", "bold"} {
		b, err := FontBytes(w)
		if err != nil || len(b) < 1000 {
			t.Fatalf("font %s: %v (%d bytes)", w, err, len(b))
		}
	}
}
