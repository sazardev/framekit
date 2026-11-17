package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sazardev/framekit"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "render":
		renderCmd(os.Args[2:])
	case "init":
		initCmd(os.Args[2:])
	case "presets":
		for name, s := range framekit.Presets {
			fmt.Printf("%-10s %dx%d\n", name, s.W, s.H)
		}
	case "agent":
		fmt.Print(agentPrompt)
	case "schema":
		schemaCmd(os.Args[2:])
	case "plan":
		planCmd(os.Args[2:])
	case "validate":
		validateCmd(os.Args[2:])
	case "describe":
		describeCmd(os.Args[2:])
	case "themes":
		themesCmd(os.Args[2:])
	case "skills":
		skillsCmd(os.Args[2:])
	case "version":
		fmt.Println("framekit", version)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`framekit — declarative video framework (Go, local-first, gruvbox)

Usage:
  framekit render   --spec spec.json [--out video.mp4]
  framekit plan     --brief "30s portrait reel about my Flutter apps, noir music"
  framekit validate --spec spec.json
  framekit describe --spec spec.json
  framekit schema                 # JSON Schema of the spec
  framekit agent                  # full contract + rules for AI agents
  framekit skills [--show]        # agent skill documents
  framekit themes                 # available palettes
  framekit init [--dir .]
  framekit presets
  framekit version

Formats: landscape (1280x720), square (1080x1080), portrait (1080x1920).
`)
}

func renderCmd(args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	specPath := fs.String("spec", "", "path to JSON spec")
	out := fs.String("out", "", "output .mp4 path")
	_ = fs.Parse(args)
	if *specPath == "" {
		fmt.Fprintln(os.Stderr, "--spec is required")
		os.Exit(1)
	}
	data, err := os.ReadFile(*specPath)
	if err != nil {
		fatal(err)
	}
	var spec framekit.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		fatal(err)
	}
	if err := framekit.Render(spec, *out); err != nil {
		fatal(err)
	}
	dest := *out
	if dest == "" {
		dest = spec.Output
	}
	fmt.Println("rendered:", dest)
}

func initCmd(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dir := fs.String("dir", ".", "target directory")
	_ = fs.Parse(args)
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		fatal(err)
	}
	p := filepath.Join(*dir, "intro.json")
	if err := os.WriteFile(p, []byte(exampleSpec), 0o644); err != nil {
		fatal(err)
	}
	fmt.Println("wrote", p)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

const exampleSpec = `{
  "title": "SazarDev — Freelance Developer Intro",
  "format": "landscape",
  "fps": 30,
  "effects": {
    "typewriter": true, "cursor": true, "fade": 0.38,
    "slide": 10, "progress": true, "chrome": true
  },
  "music": { "enabled": true, "volume": 0.25, "style": "ambient" },
  "scenes": [
    {
      "kind": "terminal", "name": "01 / IDENTITY", "duration": 4,
      "command": "whoami",
      "lines": [
        { "text": "Omar Flores", "font": "huge", "at": 1.3 },
        { "text": "Full-Stack Freelance Developer", "color": "green", "font": "bold", "at": 1.9 }
      ]
    },
    {
      "kind": "stats", "name": "04 / IMPACT", "duration": 5,
      "stats": [
        { "value": "29", "label": "APPS SHIPPED" },
        { "value": "26", "label": "PROJECTS" },
        { "value": "300+", "label": "GITHUB STARS" }
      ]
    },
    {
      "kind": "credits", "name": "05 / CONTACT", "duration": 4,
      "title": "SazarDev",
      "subtitle": "Let's build something solid together.",
      "lines": [ { "text": "@salazarprograms  ·  Fiverr" } ]
    }
  ],
  "output": "intro.mp4"
}
`

const agentPrompt = `You are a framekit video agent. framekit renders local, flat-design videos
with Go + ffmpeg. Get the machine-readable contract with ` + "`framekit schema`" + ` and the
skill document with ` + "`framekit skills --show`" + `. Workflow:
1. ` + "`framekit plan --brief \"<brief>\" --out spec.json`" + ` for a first draft.
2. Edit spec.json using the schema.
3. ` + "`framekit validate --spec spec.json`" + ` and ` + "`framekit describe --spec spec.json`" + `.
4. ` + "`framekit render --spec spec.json --out video.mp4`" + `.

Respond with ONLY a JSON object matching this contract:

{
  "title": string,
  "format": "landscape" | "square" | "portrait",
  "fps": 30,
  "theme": {"name": "gruvbox|leader|catppuccin|tokyo-night|nord|dracula|solarized|light"},
  "branding": {"name":"Omar Flores", "subtitle":"video studio",
               "meta":"...", "windowTitle":"video",
               "footerLeft":"...", "footerRight":"..."},
  "fonts": {"regular":"/path/regular.ttf","bold":"/path/bold.ttf"},
  "effects": {"typewriter":true,"cursor":true,"fade":0.38,"slide":10,
              "progress":true,"chrome":true,
              "easing":"linear|easeIn|easeInOut|easeOutCubic|easeOutQuint|spring",
              "transition":"bgfade|wipe|none","scanline":false,
              "tracking":1.5,"stagger":0.4},
  "music": {"enabled":true,"volume":0.25,
            "progression":"ambient|uplift|noir|minimal","tempo":80},
  "scenes": [
    {"kind":"terminal|code|title|stats|credits","name":"01 / LABEL",
     "duration":seconds,"accent":"green",
     "command":"prompt text","title":"...","subtitle":"...",
     "lines":[{"text":"...","color":"green|aqua|blue|yellow|fg|dim",
               "font":"normal|bold|huge","at":seconds}],
     "stats":[{"value":"29","label":"APPS SHIPPED","at":0.4}],
     "items":[{"name":"shiki","note":"331 stars","at":0.8}]}
  ],
  "voices": [{"provider":"cloudflare","text":"...","model":"@cf/facebook/mms-tts-eng","at":0,"volume":1}],
  "output":"video.mp4"
}

Rules:
- Total duration 15-60s. Prefer 5 scenes with an eyebrow "name" per scene.
- "premium"/"cinematic": easeOutQuint + transition wipe + scanline + tracking 2.5.
- Personalize branding and palette; keep flat design (no gradients/shadows/radii).
- Lines never wrap: keep them short. Use ` + "`at`" + ` for staggered reveals.
- Voices are opt-in; Cloudflare needs CF_ACCOUNT_ID and CF_API_TOKEN. Never send data otherwise.
- Save to spec.json, validate, describe, then render.
`
