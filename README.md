# framekit

Declarative video framework in Go. Local-first, flat-design (gruvbox + Geist Mono),
with synthesized music, optional Cloudflare Workers AI voices and four output
formats: 1080p landscape, 720p landscape, square and portrait — all at 60 fps.

It generates the exact visual language of [sazardev.stream](https://sazardev.stream):
no gradients, no shadows, no border radius — only 1px dividers, gruvbox colors and
Geist Mono. Frames are drawn in pure Go and piped to ffmpeg as raw RGBA.

## Why

- **Private by default.** Everything renders on your machine. Nothing is uploaded
  unless you explicitly declare a Cloudflare voice.
- **Tiny surface.** One binary, one JSON spec, one output file.
- **Agent-friendly.** `framekit agent` prints the exact spec contract so an AI
  agent can emit a spec and render it without reading the source.
- **Reusable design system.** Theme colors are overridable per spec; the default
  is the portfolio palette.

## Install

```bash
git clone https://github.com/sazardev/framekit
cd framekit
go build -o framekit ./cmd/framekit
```

Requires `ffmpeg` in `PATH` (only for encoding/muxing).

## Quick start

```bash
./framekit render --spec examples/intro.json --out intro.mp4
./framekit init --dir my-video      # writes a starter spec
./framekit presets                  # landscape / square / portrait
./framekit agent                    # spec contract for AI agents
```

## Platform: templates, rules, quality and safety

### Templates

Six production templates with `{{variable}}` substitution (`{{var:default}}` works too):

```bash
./framekit templates
./framekit use --template reel --vars vars.json --out spec.json
```

Built-ins: `intro`, `reel`, `pitch`, `presentation`, `image-showcase`, `quote`.
Custom templates go in `.framekit.json` under `"templates"`.

### Quality / compression

**All profiles render at 60 fps.** Tuned for flat text graphics in motion:
x264 `animation` tune, bitrate cap and level 4.2 (safe for 1080×1920@60).

```bash
./framekit quality
./framekit render --spec spec.json --quality draft          # 540p60, crf 32 (fast preview)
./framekit render --spec spec.json --quality social         # 1080p60, 1.5M cap
./framekit render --spec spec.json --quality high           # 1080p60, crf 22, veryfast (default)
./framekit render --spec spec.json --quality master         # 1080p60, crf 18
./framekit render --spec spec.json --quality high --fps 30  # override if needed
```

The linter emits `FPS_BELOW_60` if a profile drops below 60 fps. Disable with
`"rules": {"allowLowerFPS": true}` in `.framekit.json`. `spec.fps` is
documentation; the quality profile wins.

Measured on the 25s 1080p60 intro (1500 frames):
**draft 329 KB / 12 s render · high 881 KB / 39 s render**.
All outputs are faststart H.264 + AAC. The frame painter reuses one context
across all frames to keep allocation overhead low.

### Linter & analyzer

`render` runs the guardrails automatically. Inspect them explicitly with:

```bash
./framekit lint --spec spec.json [--json]      # errors exit 2
./framekit analyze --spec spec.json            # full metrics as JSON
```

Checks include duration bounds, line/note length, **real text overflow measured
with the actual font**, item count, WCAG contrast, scene density, `at` timing,
forbidden words, estimated file size, asset paths and remote-fetch policy.
Use `--no-strict` only when you knowingly override a rule.

### Guardrails & security

```bash
./framekit config --out .framekit.json   # writes the full policy
```

`.framekit.json` defines `rules` (duration, formats, themes, contrast, branding,
music, size cap...) and `security`:

- `allowedAssetRoots` (default `"."`) — rejects traversal and out-of-tree files
- `denyExtensions` — blocks executables/scripts as assets
- `maxAssetSizeMB` — asset size cap
- `allowRemoteFetch` (default `false`) — Cloudflare voices need explicit opt-in
- text sanitization strips control characters from every agent-generated string

### Images, slides and presentations

```json
{"kind":"image","image":{"path":"cover.png","fit":"cover","zoom":0.08,"border":true},
 "title":"My project","subtitle":"Built with framekit"}
```

`image` supports `cover|contain`, a slow Ken Burns `zoom`, horizontal `pan` and a
1px border. `slide` renders presentation slides: title, accent rule and square
bullets. See the `image-showcase` and `presentation` templates.

### Batch generation

```bash
./framekit generate --spec spec.json --count 3 --seed 42 --dir variants
./framekit generate --spec spec.json --count 3 --render --quality social
```

Deterministic variants (theme, easing, progression, pacing) for A/B testing.

## Premium & agentic

```bash
./framekit plan --brief "30s portrait reel about my Flutter apps, noir music, premium" --out spec.json
./framekit validate --spec spec.json     # fail fast
./framekit describe --spec spec.json     # scene timeline + music/voice summary
./framekit schema                        # JSON Schema of the spec
./framekit agent                         # full contract + rules for an AI agent
./framekit skills                        # skill docs for coding agents
./framekit themes                        # gruvbox, leader, catppuccin, tokyo-night, nord, dracula, solarized, light
```

`plan` is deterministic (no LLM needed) and produces a valid first draft from a
natural-language brief. Agents can also emit the JSON directly using `--schema`.

Premium feel is a set of switches, not a rewrite:

```json
"effects": {
  "easing": "easeOutQuint",
  "transition": "wipe",
  "scanline": true,
  "tracking": 2.5,
  "stagger": 0.45
}
```

See `examples/premium.json` (catppuccin, portrait, uplift @ 96 BPM, branding).

## Skills

Ready-to-use agent skills live in the repo:

| File | Purpose |
|---|---|
| `.agents/skills/framekit-video/SKILL.md` | Full skill: workflow, design rules, real portfolio facts |
| `.opencode/command/video.md` | `/video` slash command that loads the skill |

`framekit skills --show` prints both from any directory inside the repo.

## Formats

| Preset | Size | Use |
|---|---|---|
| `landscape` | **1920×1080** | Fiverr, YouTube, LinkedIn (default) |
| `landscape720` | 1280×720 | lightweight fallback |
| `square` | 1080×1080 | feeds |
| `portrait` | 1080×1920 | Shorts, Reels, TikTok |

All at 60 fps. Override with `"size": {"w": ..., "h": ...}`.

Override with `"size": {"w": 1920, "h": 1080}`. Layout, font ramp and stat
placement adapt automatically (stats stack vertically in portrait).

## Spec example

```json
{
  "format": "landscape",
  "music": { "enabled": true, "volume": 0.25 },
  "effects": { "typewriter": true, "cursor": true, "fade": 0.38,
               "slide": 10, "progress": true, "chrome": true },
  "scenes": [
    { "kind": "terminal", "name": "01 / IDENTITY", "duration": 4,
      "command": "whoami",
      "lines": [{ "text": "Omar Flores", "font": "huge", "at": 1.3 }] },
    { "kind": "stats", "duration": 4,
      "stats": [{ "value": "29", "label": "APPS SHIPPED" }] },
    { "kind": "credits", "duration": 4,
      "title": "SazarDev", "subtitle": "Let's build something solid together." }
  ],
  "output": "intro.mp4"
}
```

Scene kinds: `terminal`, `code`, `title`, `stats`, `credits`.

Color names: `bg`, `deep`, `soft`, `fg`, `dim`, `faint`, `green`, `yellow`,
`aqua`, `blue`, `red`, `purple`, `orange`.

## Music

`"music": {"enabled": true}` synthesizes a deterministic ambient WAV in pure Go:
four-chord pad (Am7 → Fmaj7 → Cmaj7 → G6), decaying arpeggio, soft noise ticks
and fade in/out. No samples, no downloads, no licensing.

## Voices (opt-in)

```json
"voices": [
  { "provider": "cloudflare",
    "text": "Hi, I'm Omar, a freelance developer.",
    "model": "@cf/facebook/mms-tts-eng",
    "at": 0.5, "volume": 1 }
]
```

```bash
export CF_ACCOUNT_ID=...
export CF_API_TOKEN=...
```

The audio is downloaded, then ffmpeg delays/mixes it with the music and muxes
it into the video. Local files also work:

```json
{ "provider": "local", "file": "/path/to/voice.wav", "at": 2, "volume": 0.9 }
```

## Library usage

```go
spec := framekit.Spec{
    Format: "portrait",
    Effects: framekit.Effects{Typewriter: true, Cursor: true, Chrome: true, Progress: true},
    Music: framekit.Music{Enabled: true, Volume: 0.25},
    Scenes: []framekit.Scene{
        {Kind: "title", Duration: 3, Title: "SazarDev", Subtitle: "Freelance Developer"},
    },
    Output: "reel.mp4",
}
err := framekit.Render(spec, "reel.mp4")
```

## Architecture

```
theme.go    design system (gruvbox) + embedded Geist Mono + hex helpers
spec.go     declarative spec, format presets, normalization
layout.go   responsive layout / font ramp per aspect ratio
render.go   frame painter + ffmpeg rawvideo encoder
audio.go    music synthesizer + Cloudflare TTS + ffmpeg mux
cmd/        CLI
examples/   landscape, square, portrait specs
```

## License

Code MIT. Geist Mono is bundled under the SIL Open Font License (see `assets/fonts/OFL-Geist.txt`).
