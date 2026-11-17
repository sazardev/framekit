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
  framekit render --spec spec.json [--out video.mp4]
  framekit init   [--dir .]
  framekit presets
  framekit agent        # prints the spec contract for AI agents
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

const agentPrompt = `You are generating a framekit video spec. framekit renders local, flat-design
videos with Go + ffmpeg. Respond with ONLY a JSON object matching this contract:

{
  "title": string,
  "format": "landscape" | "square" | "portrait",
  "fps": 30,
  "effects": {"typewriter": bool, "cursor": bool, "fade": 0.38,
              "slide": 10, "progress": true, "chrome": true},
  "music": {"enabled": true, "volume": 0.25, "style": "ambient"},
  "scenes": [
    {"kind": "terminal"|"title"|"stats"|"credits"|"code",
     "name": "01 / LABEL", "duration": seconds,
     "command": "prompt text",
     "lines": [{"text":"...", "color":"green|aqua|blue|yellow|fg|dim",
                "font":"normal|bold|huge", "at": seconds}],
     "stats": [{"value":"29","label":"APPS"}],
     "items": [{"name":"shiki","note":"331 stars","at":1.0}]}
  ],
  "voices": [{"provider":"cloudflare","text":"...","at":0,"volume":1}],
  "output": "video.mp4"
}

Rules:
- Keep total duration between 15 and 60 seconds.
- Use the default gruvbox theme unless the user asks for another palette.
- Prefer short lines; the renderer never wraps text.
- Colors: bg, deep, soft, fg, dim, faint, green, yellow, aqua, blue, red, purple, orange.
- Voices are opt-in; Cloudflare needs CF_ACCOUNT_ID and CF_API_TOKEN.
- Write the JSON to a file, then run: framekit render --spec spec.json --out video.mp4
`
