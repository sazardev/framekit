package framekit

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the project guardrail file (.framekit.json). It defines quality
// profiles, hard rules, security boundaries and defaults. It is optional:
// sane defaults apply when the file is absent.
type Config struct {
	Defaults  Defaults                   `json:"defaults"`
	Quality   map[string]Quality         `json:"quality"`
	Rules     Rules                      `json:"rules"`
	Security  Security                   `json:"security"`
	Templates map[string]json.RawMessage `json:"templates"`
}

type Defaults struct {
	Format  string `json:"format"`
	Quality string `json:"quality"`
	Theme   string `json:"theme"`
	Strict  bool   `json:"strict"`
}

// Quality controls the encoder. Profiles trade fidelity for size.
type Quality struct {
	Scale        float64 `json:"scale"`
	FPS          int     `json:"fps"`
	CRF          int     `json:"crf"`
	Preset       string  `json:"preset"`
	Tune         string  `json:"tune"`
	Profile      string  `json:"profile"`
	Level        string  `json:"level"`
	MaxRate      string  `json:"maxrate"`
	BufSize      string  `json:"bufsize"`
	AudioBitrate string  `json:"audioBitrate"`
	Description  string  `json:"description"`
}

// Rules are guardrails enforced by Lint. They are deliberately explicit so an
// agent cannot silently produce off-brand or oversized videos.
type Rules struct {
	MinDuration      float64  `json:"minDuration"`
	MaxDuration      float64  `json:"maxDuration"`
	MaxFileSizeMB    float64  `json:"maxFileSizeMB"`
	AllowedFormats   []string `json:"allowedFormats"`
	AllowedThemes    []string `json:"allowedThemes"`
	MaxLineChars     int      `json:"maxLineChars"`
	MaxNoteChars     int      `json:"maxNoteChars"`
	MaxItemsPerScene int      `json:"maxItemsPerScene"`
	MaxScenes        int      `json:"maxScenes"`
	MinContrast      float64  `json:"minContrast"`
	RequireBranding  bool     `json:"requireBranding"`
	RequireMusic     bool     `json:"requireMusic"`
	RequireEyebrows  bool     `json:"requireEyebrows"`
	MaxVoices        int      `json:"maxVoices"`
	ForbiddenWords   []string `json:"forbiddenWords"`
	SafeZonePct      float64  `json:"safeZonePct"`
	Require60FPS     bool     `json:"require60fps"`
	AllowLowerFPS    bool     `json:"allowLowerFPS"`
}

// Security defines the local sandbox for assets and remote calls.
type Security struct {
	AllowedAssetRoots []string `json:"allowedAssetRoots"`
	AllowedVoiceHosts []string `json:"allowedVoiceHosts"`
	MaxAssetSizeMB    float64  `json:"maxAssetSizeMB"`
	DenyExtensions    []string `json:"denyExtensions"`
	AllowRemoteFetch  bool     `json:"allowRemoteFetch"`
}

// QualityProfiles are tuned for flat, text-heavy graphics: x264 "animation"
// tune, capped bitrate and faststart give crisp type at a small size.
var QualityProfiles = map[string]Quality{
	"draft": {
		Scale: 0.5, FPS: 60, CRF: 31, Preset: "veryfast", Tune: "animation",
		Profile: "main", Level: "4.2", MaxRate: "600k", BufSize: "1200k",
		AudioBitrate: "64k", Description: "quick 60fps preview, half resolution",
	},
	"standard": {
		Scale: 1, FPS: 60, CRF: 26, Preset: "veryfast", Tune: "animation",
		Profile: "high", Level: "4.2", MaxRate: "1200k", BufSize: "2400k",
		AudioBitrate: "96k", Description: "balanced 60fps: crisp motion, small file",
	},
	"high": {
		Scale: 1, FPS: 60, CRF: 21, Preset: "medium", Tune: "animation",
		Profile: "high", Level: "4.2", MaxRate: "2400k", BufSize: "4800k",
		AudioBitrate: "128k", Description: "default premium 60fps delivery",
	},
	"master": {
		Scale: 1, FPS: 60, CRF: 17, Preset: "slow", Tune: "animation",
		Profile: "high", Level: "4.2", MaxRate: "6000k", BufSize: "12000k",
		AudioBitrate: "192k", Description: "60fps archival / re-encode source",
	},
	"social": {
		Scale: 1, FPS: 60, CRF: 24, Preset: "veryfast", Tune: "animation",
		Profile: "high", Level: "4.2", MaxRate: "900k", BufSize: "1800k",
		AudioBitrate: "96k", Description: "60fps with platform-safe bitrate cap",
	},
}

// DefaultConfig returns the shipped policy.
func DefaultConfig() Config {
	return Config{
		Defaults: Defaults{Format: "landscape", Quality: "high", Theme: "gruvbox", Strict: true},
		Quality:  QualityProfiles,
		Rules: Rules{
			MinDuration: 15, MaxDuration: 60, MaxFileSizeMB: 25,
			AllowedFormats: []string{"landscape", "square", "portrait"},
			AllowedThemes:  ThemeNames(),
			MaxLineChars:   46, MaxNoteChars: 38, MaxItemsPerScene: 8,
			MaxScenes: 12, MinContrast: 4.5,
			RequireBranding: true, RequireMusic: true, RequireEyebrows: false,
			MaxVoices: 2, ForbiddenWords: []string{}, SafeZonePct: 5,
			Require60FPS: true,
		},
		Security: Security{
			AllowedAssetRoots: []string{"."},
			AllowedVoiceHosts: []string{"api.cloudflare.com"},
			MaxAssetSizeMB:    50,
			DenyExtensions:    []string{".exe", ".sh", ".bat", ".php"},
			AllowRemoteFetch:  false,
		},
	}
}

// LoadConfig reads .framekit.json. Missing file means defaults; malformed file
// is an error (never silently ignore a guardrail).
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = ".framekit.json"
	}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	var user Config
	if err := json.Unmarshal(b, &user); err != nil {
		return cfg, fmt.Errorf("config %s: %w", path, err)
	}
	if user.Defaults.Format != "" {
		cfg.Defaults.Format = user.Defaults.Format
	}
	if user.Defaults.Quality != "" {
		cfg.Defaults.Quality = user.Defaults.Quality
	}
	if user.Defaults.Theme != "" {
		cfg.Defaults.Theme = user.Defaults.Theme
	}
	if user.Defaults.Strict {
		cfg.Defaults.Strict = true
	}
	for k, v := range user.Quality {
		cfg.Quality[k] = v
	}
	mergeRules(&cfg.Rules, user.Rules)
	mergeSecurity(&cfg.Security, user.Security)
	if len(user.Templates) > 0 {
		cfg.Templates = user.Templates
	}
	return cfg, nil
}

func mergeRules(dst *Rules, src Rules) {
	if src.MinDuration > 0 {
		dst.MinDuration = src.MinDuration
	}
	if src.MaxDuration > 0 {
		dst.MaxDuration = src.MaxDuration
	}
	if src.MaxFileSizeMB > 0 {
		dst.MaxFileSizeMB = src.MaxFileSizeMB
	}
	if len(src.AllowedFormats) > 0 {
		dst.AllowedFormats = src.AllowedFormats
	}
	if len(src.AllowedThemes) > 0 {
		dst.AllowedThemes = src.AllowedThemes
	}
	if src.MaxLineChars > 0 {
		dst.MaxLineChars = src.MaxLineChars
	}
	if src.MaxNoteChars > 0 {
		dst.MaxNoteChars = src.MaxNoteChars
	}
	if src.MaxItemsPerScene > 0 {
		dst.MaxItemsPerScene = src.MaxItemsPerScene
	}
	if src.MaxScenes > 0 {
		dst.MaxScenes = src.MaxScenes
	}
	if src.MinContrast > 0 {
		dst.MinContrast = src.MinContrast
	}
	if src.RequireBranding {
		dst.RequireBranding = true
	}
	if src.RequireMusic {
		dst.RequireMusic = true
	}
	if src.RequireEyebrows {
		dst.RequireEyebrows = true
	}
	if src.MaxVoices > 0 {
		dst.MaxVoices = src.MaxVoices
	}
	if len(src.ForbiddenWords) > 0 {
		dst.ForbiddenWords = src.ForbiddenWords
	}
	if src.SafeZonePct > 0 {
		dst.SafeZonePct = src.SafeZonePct
	}
	if src.Require60FPS {
		dst.Require60FPS = true
	}
	if src.AllowLowerFPS {
		dst.Require60FPS = false
	}
}

func mergeSecurity(dst *Security, src Security) {
	if len(src.AllowedAssetRoots) > 0 {
		dst.AllowedAssetRoots = src.AllowedAssetRoots
	}
	if len(src.AllowedVoiceHosts) > 0 {
		dst.AllowedVoiceHosts = src.AllowedVoiceHosts
	}
	if src.MaxAssetSizeMB > 0 {
		dst.MaxAssetSizeMB = src.MaxAssetSizeMB
	}
	if len(src.DenyExtensions) > 0 {
		dst.DenyExtensions = src.DenyExtensions
	}
	if src.AllowRemoteFetch {
		dst.AllowRemoteFetch = true
	}
}

// QualityFor resolves a quality profile.
func (c Config) QualityFor(name string) (Quality, error) {
	if name == "" {
		name = c.Defaults.Quality
	}
	q, ok := c.Quality[name]
	if !ok {
		return q, fmt.Errorf("unknown quality %q", name)
	}
	return q, nil
}

// QualityName resolves a possibly-empty profile name to the default.
func (c Config) QualityName(name string) string {
	if name == "" {
		return c.Defaults.Quality
	}
	return name
}

// EffectiveSize applies a quality scale, keeping even dimensions for yuv420p.
func EffectiveSize(spec Spec, q Quality) Size {
	sz := *spec.Size
	if q.Scale <= 0 || q.Scale == 1 {
		return sz
	}
	even := func(v int) int {
		v = int(float64(v) * q.Scale)
		if v%2 != 0 {
			v++
		}
		if v < 64 {
			v = 64
		}
		return v
	}
	return Size{W: even(sz.W), H: even(sz.H)}
}
