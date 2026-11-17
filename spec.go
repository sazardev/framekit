package framekit

import "fmt"

// Format presets. Landscape is true 1080p; landscape720 is a lightweight
// fallback. Square and portrait are 1:1 and 9:16 at 1080px wide.
var Presets = map[string]Size{
	"landscape":    {W: 1920, H: 1080},
	"landscape720": {W: 1280, H: 720},
	"square":       {W: 1080, H: 1080},
	"portrait":     {W: 1080, H: 1920},
}

type Size struct {
	W int `json:"w"`
	H int `json:"h"`
}

// Spec is the declarative interface. It is deliberately JSON-friendly so an
// agent can emit it and call `framekit render --spec spec.json`.
type Spec struct {
	Title    string     `json:"title"`
	Format   string     `json:"format,omitempty"`
	Size     *Size      `json:"size,omitempty"`
	FPS      int        `json:"fps,omitempty"`
	Theme    *Theme     `json:"theme,omitempty"`
	Branding Branding   `json:"branding,omitempty"`
	Fonts    FontConfig `json:"fonts,omitempty"`
	Effects  Effects    `json:"effects"`
	Scenes   []Scene    `json:"scenes"`
	Music    Music      `json:"music"`
	Voices   []Voice    `json:"voices,omitempty"`
	Output   string     `json:"output,omitempty"`
	FFmpeg   string     `json:"ffmpeg,omitempty"`

	// Duration is computed from scenes by Normalize.
	Duration float64 `json:"duration,omitempty"`
}

// Branding personalizes all repeated chrome text. Empty fields fall back to
// design-system defaults.
type Branding struct {
	Name        string `json:"name,omitempty"`
	Subtitle    string `json:"subtitle,omitempty"`
	Meta        string `json:"meta,omitempty"`
	WindowTitle string `json:"windowTitle,omitempty"`
	FooterLeft  string `json:"footerLeft,omitempty"`
	FooterRight string `json:"footerRight,omitempty"`
}

// FontConfig allows local TTF overrides. Embedded Geist Mono is the default.
type FontConfig struct {
	Regular string `json:"regular,omitempty"`
	Bold    string `json:"bold,omitempty"`
}

type Effects struct {
	Typewriter bool    `json:"typewriter"`
	Cursor     bool    `json:"cursor"`
	Fade       float64 `json:"fade"`
	Slide      float64 `json:"slide"`
	Progress   bool    `json:"progress"`
	Chrome     bool    `json:"chrome"`

	Easing     string  `json:"easing,omitempty"`     // linear|easeOut|easeInOut|spring
	Transition string  `json:"transition,omitempty"` // bgfade|wipe|none
	Scanline   bool    `json:"scanline,omitempty"`
	Tracking   float64 `json:"tracking,omitempty"` // extra px per char on hero text
	Stagger    float64 `json:"stagger,omitempty"`  // seconds between items
}

// Scene is one "slide". Kind: terminal | title | stats | credits | code.
type Scene struct {
	Kind     string    `json:"kind"`
	Name     string    `json:"name,omitempty"` // eyebrow label, e.g. "03 / PROJECTS"
	Duration float64   `json:"duration"`
	Command  string    `json:"command,omitempty"`
	Title    string    `json:"title,omitempty"`
	Subtitle string    `json:"subtitle,omitempty"`
	Lines    []Line    `json:"lines,omitempty"`
	Stats    []Stat    `json:"stats,omitempty"`
	Items    []Item    `json:"items,omitempty"`
	Accent   string    `json:"accent,omitempty"` // per-scene accent override
	Image    *ImageRef `json:"image,omitempty"`
}

// ImageRef draws a local image in an "image" scene. Fit: cover|contain.
// Zoom > 0 enables a slow Ken Burns push; Pan moves it horizontally.
type ImageRef struct {
	Path   string  `json:"path"`
	Fit    string  `json:"fit,omitempty"`
	Zoom   float64 `json:"zoom,omitempty"`
	Pan    float64 `json:"pan,omitempty"`
	At     float64 `json:"at,omitempty"`
	Border bool    `json:"border,omitempty"`
}

type Line struct {
	Text  string  `json:"text"`
	Color string  `json:"color,omitempty"`
	At    float64 `json:"at,omitempty"`
	Font  string  `json:"font,omitempty"` // normal | bold | huge
}

type Stat struct {
	Value string  `json:"value"`
	Label string  `json:"label"`
	At    float64 `json:"at,omitempty"`
}

type Item struct {
	Name string  `json:"name"`
	Note string  `json:"note,omitempty"`
	At   float64 `json:"at,omitempty"`
}

// Music describes the locally synthesized ambient track.
type Music struct {
	Enabled     bool    `json:"enabled"`
	Volume      float64 `json:"volume,omitempty"`
	Style       string  `json:"style,omitempty"`       // legacy alias of Progression
	Progression string  `json:"progression,omitempty"` // ambient|uplift|noir|minimal
	Tempo       float64 `json:"tempo,omitempty"`       // BPM for the arpeggio
}

// Voice is an optional narration. Provider "cloudflare" is opt-in and reads
// CF_ACCOUNT_ID / CF_API_TOKEN from the environment. Nothing is uploaded by
// default: local files are mixed directly.
type Voice struct {
	Provider string  `json:"provider,omitempty"` // local | cloudflare
	Text     string  `json:"text,omitempty"`
	Model    string  `json:"model,omitempty"`
	File     string  `json:"file,omitempty"`
	At       float64 `json:"at,omitempty"`
	Volume   float64 `json:"volume,omitempty"`
}

// Normalize fills defaults and validates the spec.
func (s *Spec) Normalize() error {
	if s.Format == "" {
		s.Format = "landscape"
	}
	size, ok := Presets[s.Format]
	if !ok {
		return fmt.Errorf("unknown format %q (use landscape, square or portrait)", s.Format)
	}
	if s.Size != nil {
		size = *s.Size
	}
	if size.W <= 0 || size.H <= 0 {
		return fmt.Errorf("invalid size %+v", size)
	}
	if s.Size == nil {
		s.Size = &size
	}
	if s.FPS <= 0 {
		s.FPS = 30
	}
	if s.FPS > 120 {
		return fmt.Errorf("fps too high: %d", s.FPS)
	}
	if s.Theme == nil {
		def := Gruvbox()
		s.Theme = &def
	}
	base := ThemeByName(s.Theme.Name).Merge(s.Theme)
	if base.Brand == "" {
		base.Brand = Gruvbox().Brand
	}
	s.Theme = &base

	// Branding defaults come from the theme but can be overridden per spec.
	b := &s.Branding
	if b.Name == "" {
		b.Name = s.Theme.Brand
	}
	if b.Subtitle == "" {
		b.Subtitle = "video studio"
	}
	if b.Meta == "" {
		b.Meta = "framekit · 100% local render"
	}
	if b.WindowTitle == "" {
		b.WindowTitle = "video"
	}
	if b.FooterLeft == "" {
		b.FooterLeft = "FRAMEKIT · PRIVATE RENDER"
	}
	if b.FooterRight == "" {
		b.FooterRight = "GENERATED LOCALLY"
	}

	// Effects default to the full clean system.
	if s.Effects.Fade == 0 {
		s.Effects.Fade = 0.38
	}
	if s.Effects.Slide == 0 {
		s.Effects.Slide = 10
	}
	if s.Effects.Easing == "" {
		s.Effects.Easing = "easeOutCubic"
	}
	if s.Effects.Transition == "" {
		s.Effects.Transition = "bgfade"
	}
	if s.Effects.Stagger <= 0 {
		s.Effects.Stagger = 0.4
	}
	if s.Music.Progression == "" {
		s.Music.Progression = s.Music.Style
	}
	if s.Music.Progression == "" {
		s.Music.Progression = "ambient"
	}
	if s.Music.Tempo <= 0 {
		s.Music.Tempo = 80
	}
	if len(s.Scenes) == 0 {
		return fmt.Errorf("spec has no scenes")
	}
	var total float64
	for i := range s.Scenes {
		sc := &s.Scenes[i]
		if sc.Kind == "" {
			sc.Kind = "terminal"
		}
		if sc.Duration <= 0 {
			sc.Duration = 4
		}
		total += sc.Duration
	}
	s.Duration = total
	if s.Music.Volume <= 0 {
		s.Music.Volume = 0.25
	}
	for i := range s.Voices {
		if s.Voices[i].Volume <= 0 {
			s.Voices[i].Volume = 1
		}
		if s.Voices[i].Provider == "" {
			if s.Voices[i].File != "" {
				s.Voices[i].Provider = "local"
			} else {
				s.Voices[i].Provider = "cloudflare"
			}
		}
	}
	return nil
}

// DurationSeconds returns the total duration computed by Normalize.
func (s Spec) DurationSeconds() float64 {
	return s.Duration
}
