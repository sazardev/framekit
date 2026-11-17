package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sazardev/framekit"
)

func schemaCmd(args []string) {
	_ = flag.NewFlagSet("schema", flag.ExitOnError).Parse(args)
	fmt.Print(specSchema)
}

func validateCmd(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	p := fs.String("spec", "", "spec path")
	_ = fs.Parse(args)
	spec := loadSpec(*p)
	if err := spec.Normalize(); err != nil {
		fatal(err)
	}
	fmt.Printf("valid · %s · %s · %.0fs · %d scenes · theme %s\n",
		spec.Format, spec.Title, spec.DurationSeconds(), len(spec.Scenes), spec.Theme.Name)
}

func describeCmd(args []string) {
	fs := flag.NewFlagSet("describe", flag.ExitOnError)
	p := fs.String("spec", "", "spec path")
	_ = fs.Parse(args)
	spec := loadSpec(*p)
	if err := spec.Normalize(); err != nil {
		fatal(err)
	}
	fmt.Printf("%s [%s · %dx%d · %s · %.0fs]\n", spec.Title, spec.Format,
		spec.Size.W, spec.Size.H, spec.Theme.Name, spec.DurationSeconds())
	start := 0.0
	for i, sc := range spec.Scenes {
		fmt.Printf("  %02d  %6.1fs → %6.1fs  %-9s  %s\n", i+1, start, start+sc.Duration, sc.Kind, sc.Name)
		start += sc.Duration
	}
	if spec.Music.Enabled {
		fmt.Printf("  music: %s @ %.0f bpm, volume %.2f\n", spec.Music.Progression, spec.Music.Tempo, spec.Music.Volume)
	}
	for i, v := range spec.Voices {
		fmt.Printf("  voice %d: %s @ %.1fs\n", i+1, v.Provider, v.At)
	}
}

func themesCmd(args []string) {
	_ = flag.NewFlagSet("themes", flag.ExitOnError).Parse(args)
	for _, n := range framekit.ThemeNames() {
		fmt.Println(n)
	}
}

func skillsCmd(args []string) {
	fs := flag.NewFlagSet("skills", flag.ExitOnError)
	show := fs.Bool("show", false, "print the skill document")
	_ = fs.Parse(args)
	files := []string{
		".agents/skills/framekit-video/SKILL.md",
		".opencode/command/video.md",
	}
	for _, f := range files {
		if *show {
			b, err := os.ReadFile(f)
			if err == nil {
				fmt.Printf("===== %s =====\n%s\n", f, b)
				continue
			}
		}
		if _, err := os.Stat(f); err == nil {
			fmt.Println(f)
		} else {
			fmt.Println(f, "(not found — run from the framekit repo)")
		}
	}
}

// planCmd deterministically turns a natural-language brief into a spec. No LLM
// is required; an agent can also emit the spec directly using --schema.
func planCmd(args []string) {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	brief := fs.String("brief", "", "natural language brief")
	out := fs.String("out", "", "write spec to file instead of stdout")
	_ = fs.Parse(args)
	if strings.TrimSpace(*brief) == "" {
		fatal(fmt.Errorf("--brief is required"))
	}
	spec := planFromBrief(*brief)
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fatal(err)
	}
	if *out != "" {
		if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
			fatal(err)
		}
		fmt.Println("wrote", *out)
		return
	}
	fmt.Println(string(data))
}

func planFromBrief(brief string) framekit.Spec {
	lower := strings.ToLower(brief)
	s := framekit.Spec{
		Title:  titleFromBrief(brief),
		Format: "landscape",
		FPS:    60,
		Effects: framekit.Effects{
			Typewriter: true, Cursor: true, Chrome: true,
			Progress: true, Fade: 0.38, Slide: 10,
			Easing: "easeOutCubic", Transition: "bgfade", Stagger: 0.4,
			Tracking: 1.5,
		},
		Music: framekit.Music{Enabled: true, Volume: 0.25, Progression: "ambient", Tempo: 80},
	}

	if strings.Contains(lower, "portrait") || strings.Contains(lower, "reel") ||
		strings.Contains(lower, "short") || strings.Contains(lower, "tiktok") {
		s.Format = "portrait"
	} else if strings.Contains(lower, "square") || strings.Contains(lower, "feed") {
		s.Format = "square"
	}
	for _, t := range framekit.ThemeNames() {
		if strings.Contains(lower, t) {
			s.Theme = &framekit.Theme{Name: t}
			break
		}
	}
	for _, mood := range []string{"uplift", "noir", "minimal", "ambient"} {
		if strings.Contains(lower, mood) {
			s.Music.Progression = mood
		}
	}
	if strings.Contains(lower, "no music") || strings.Contains(lower, "silent") {
		s.Music.Enabled = false
	}
	if strings.Contains(lower, "cinematic") || strings.Contains(lower, "premium") {
		s.Effects.Easing = "easeOutQuint"
		s.Effects.Scanline = true
		s.Effects.Transition = "wipe"
		s.Effects.Tracking = 2.5
	}

	// Scene selection by keyword.
	scenes := []framekit.Scene{}
	if strings.Contains(lower, "stack") || strings.Contains(lower, "flutter") ||
		strings.Contains(lower, "react") || strings.Contains(lower, "go ") || strings.Contains(lower, "rust") {
		scenes = append(scenes, framekit.Scene{
			Kind: "terminal", Name: "02 / STACK", Duration: 5, Command: "cat stack.txt",
			Items: []framekit.Item{
				{Name: "Flutter / Dart", Note: "mobile apps"},
				{Name: "React / TypeScript", Note: "web apps"},
				{Name: "Go / Rust", Note: "CLIs & systems"},
				{Name: "Firebase / Postgres", Note: "data & infra"},
			},
		})
	}
	if strings.Contains(lower, "project") || strings.Contains(lower, "portfolio") || strings.Contains(lower, "app") {
		scenes = append(scenes, framekit.Scene{
			Kind: "terminal", Name: "03 / PROJECTS", Duration: 6, Command: "ls ~/portfolio",
			Items: []framekit.Item{
				{Name: "shiki", Note: "331 stars · TUI in Rust"},
				{Name: "oxid", Note: "preview environments"},
				{Name: "road-to", Note: "10 apps on Google Play"},
				{Name: "+ 20 more", Note: "open source & platforms"},
			},
		})
	}
	if strings.Contains(lower, "number") || strings.Contains(lower, "stat") || strings.Contains(lower, "metric") {
		scenes = append(scenes, framekit.Scene{
			Kind: "stats", Name: "04 / IMPACT", Duration: 4,
			Stats: []framekit.Stat{
				{Value: "29", Label: "APPS SHIPPED"},
				{Value: "26", Label: "PROJECTS"},
				{Value: "300+", Label: "GITHUB STARS"},
			},
		})
	}
	if strings.Contains(lower, "hire") || strings.Contains(lower, "contact") || strings.Contains(lower, "fiverr") {
		scenes = append(scenes, framekit.Scene{
			Kind: "credits", Name: "05 / CONTACT", Duration: 4,
			Title: "Let's build", Subtitle: "something solid together.",
			Lines: []framekit.Line{{Text: "@salazarprograms · Fiverr", Color: "aqua", At: 1.5}},
		})
	}
	if len(scenes) == 0 {
		scenes = append(scenes, framekit.Scene{
			Kind: "title", Name: "01 / INTRO", Duration: 3,
			Title: s.Title, Subtitle: "Freelance Developer",
		})
	}
	// Always open with identity and close with contact when missing.
	hasCredits := false
	for _, sc := range scenes {
		if sc.Kind == "credits" {
			hasCredits = true
		}
	}
	if !hasCredits {
		scenes = append(scenes, framekit.Scene{
			Kind: "credits", Name: "09 / CONTACT", Duration: 3,
			Title: "SazarDev", Subtitle: "Hire me on Fiverr",
			Lines: []framekit.Line{{Text: "@salazarprograms", Color: "aqua", At: 1.2}},
		})
	}
	s.Scenes = scenes

	// Duration hint scales scenes proportionally; always keep 15-60s.
	re := regexp.MustCompile(`(\d{2,3})\s*(?:s|sec|seconds)`)
	if m := re.FindStringSubmatch(lower); len(m) == 2 {
		var target float64
		if _, err := fmt.Sscanf(m[1], "%f", &target); err == nil && target >= 15 && target <= 60 {
			current := 0.0
			for _, sc := range s.Scenes {
				current += sc.Duration
			}
			if current > 0 {
				scale := target / current
				for i := range s.Scenes {
					s.Scenes[i].Duration *= scale
				}
			}
		}
	}
	// Voice is opt-in and only when explicitly requested.
	if strings.Contains(lower, "voice") || strings.Contains(lower, "narrat") {
		s.Voices = []framekit.Voice{{
			Provider: "cloudflare", Text: "Hi, I'm Omar, a freelance developer.",
			Model: "@cf/facebook/mms-tts-eng", At: 0.5, Volume: 1,
		}}
	}
	s.Output = "video.mp4"
	return s
}

func titleFromBrief(brief string) string {
	stop := map[string]bool{
		"a": true, "the": true, "my": true, "about": true, "with": true, "and": true,
		"for": true, "me": true, "hire": true, "fiverr": true, "video": true,
		"reel": true, "short": true, "portrait": true, "landscape": true, "square": true,
		"premium": true, "cinematic": true, "music": true, "voice": true, "make": true,
		"create": true, "s": true, "sec": true, "seconds": true,
		"noir": true, "ambient": true, "uplift": true, "minimal": true,
		"on": true, "in": true, "of": true, "to": true, "from": true, "into": true,
	}
	fields := strings.FieldsFunc(strings.ToLower(brief), func(r rune) bool {
		return r == ' ' || r == ',' || r == '-' || r == '.' || r == ':'
	})
	words := []string{}
	for _, f := range fields {
		if len(f) < 2 || stop[f] || regexp.MustCompile(`^\d`).MatchString(f) {
			continue
		}
		if f == "flutter" || f == "react" || f == "rust" || f == "go" || f == "python" {
			words = append(words, strings.ToUpper(f))
			continue
		}
		words = append(words, strings.ToUpper(f[:1])+f[1:])
		if len(words) == 4 {
			break
		}
	}
	if len(words) == 0 {
		return "SazarDev Video"
	}
	return strings.Join(words, " ")
}

func loadSpec(path string) framekit.Spec {
	if path == "" {
		fatal(fmt.Errorf("--spec is required"))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	var s framekit.Spec
	if err := json.Unmarshal(data, &s); err != nil {
		fatal(err)
	}
	return s
}

const specSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "framekit Spec",
  "type": "object",
  "required": ["scenes"],
  "properties": {
    "title": {"type": "string"},
    "format": {"enum": ["landscape", "square", "portrait"]},
    "size": {"type": "object", "properties": {"w": {"type": "integer"}, "h": {"type": "integer"}}},
    "fps": {"type": "integer", "minimum": 1, "maximum": 120},
    "theme": {"type": "object", "description": "Hex overrides; name selects a preset", "properties": {
      "name": {"type": "string"}, "brand": {"type": "string"},
      "bg": {"type": "string"}, "deep": {"type": "string"}, "soft": {"type": "string"},
      "fg": {"type": "string"}, "dim": {"type": "string"}, "faint": {"type": "string"},
      "green": {"type": "string"}, "yellow": {"type": "string"}, "aqua": {"type": "string"},
      "blue": {"type": "string"}, "red": {"type": "string"}, "purple": {"type": "string"},
      "orange": {"type": "string"}
    }},
    "branding": {"type": "object", "properties": {
      "name": {"type": "string"}, "subtitle": {"type": "string"}, "meta": {"type": "string"},
      "windowTitle": {"type": "string"}, "footerLeft": {"type": "string"}, "footerRight": {"type": "string"}
    }},
    "fonts": {"type": "object", "properties": {
      "regular": {"type": "string", "description": "local TTF path"},
      "bold": {"type": "string"}
    }},
    "effects": {"type": "object", "properties": {
      "typewriter": {"type": "boolean"}, "cursor": {"type": "boolean"},
      "fade": {"type": "number"}, "slide": {"type": "number"},
      "progress": {"type": "boolean"}, "chrome": {"type": "boolean"},
      "easing": {"enum": ["linear", "easeIn", "easeInOut", "easeOutCubic", "easeOutQuint", "spring"]},
      "transition": {"enum": ["bgfade", "wipe", "none"]},
      "scanline": {"type": "boolean"}, "tracking": {"type": "number"}, "stagger": {"type": "number"}
    }},
    "music": {"type": "object", "properties": {
      "enabled": {"type": "boolean"}, "volume": {"type": "number"},
      "progression": {"enum": ["ambient", "uplift", "noir", "minimal"]},
      "tempo": {"type": "number", "description": "BPM"}
    }},
    "voices": {"type": "array", "items": {"type": "object", "properties": {
      "provider": {"enum": ["local", "cloudflare"]}, "text": {"type": "string"},
      "model": {"type": "string"}, "file": {"type": "string"},
      "at": {"type": "number"}, "volume": {"type": "number"}
    }}},
    "scenes": {"type": "array", "minItems": 1, "items": {
      "type": "object", "required": ["duration"],
      "properties": {
        "kind": {"enum": ["terminal", "code", "title", "stats", "credits"]},
        "name": {"type": "string"}, "duration": {"type": "number", "minimum": 0.5},
        "command": {"type": "string"}, "title": {"type": "string"}, "subtitle": {"type": "string"},
        "accent": {"type": "string"},
        "lines": {"type": "array", "items": {"type": "object", "properties": {
          "text": {"type": "string"}, "color": {"type": "string"},
          "font": {"enum": ["normal", "bold", "huge"]}, "at": {"type": "number"}
        }}},
        "stats": {"type": "array", "items": {"type": "object", "properties": {
          "value": {"type": "string"}, "label": {"type": "string"}, "at": {"type": "number"}
        }}},
        "items": {"type": "array", "items": {"type": "object", "properties": {
          "name": {"type": "string"}, "note": {"type": "string"}, "at": {"type": "number"}
        }}}
      }
    }},
    "output": {"type": "string"}
  }
}
`
