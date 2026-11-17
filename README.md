# framekit

Declarative video framework in Go. Local-first, flat-design (gruvbox + Geist Mono),
with synthesized music, optional Cloudflare Workers AI voices and three output
formats: landscape, square and portrait.

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

## Formats

| Preset | Size | Use |
|---|---|---|
| `landscape` | 1280×720 | Fiverr, YouTube, LinkedIn |
| `square` | 1080×1080 | feeds |
| `portrait` | 1080×1920 | Shorts, Reels, TikTok |

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
