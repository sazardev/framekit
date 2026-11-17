---
name: framekit-video
description: >
  Create premium, flat-design videos with the framekit Go framework. Use when the
  user asks for an intro video, reel, promo, demo, "video para Fiverr", "make a
  video", "crea un video", or mentions framekit. Covers landscape/square/portrait,
  gruvbox design system, synthesized music, Cloudflare AI voices and validation.
---

# Framekit Video

`framekit` renders local, flat-design videos in Go. Frames are drawn with Geist Mono
and gruvbox, piped to ffmpeg, then muxed with synthesized music and optional voices.

Repo: `/home/sazar/Projects/framekit` (module `github.com/sazardev/framekit`).

## Golden rules

1. **Flat design only.** No gradients, no shadows, no border radius. 1px lines,
   solid palette, Geist Mono. This mirrors `sazardev.stream`.
2. **Local by default.** Never set a Cloudflare voice unless the user asks for a
   voice explicitly. `CF_ACCOUNT_ID`/`CF_API_TOKEN` are opt-in.
3. **Never wrap lines.** Keep every `text`/`note` short; the renderer does not wrap.
4. **Validate before rendering.** Always run `validate` and `describe` first.
5. **15-60 seconds.** Fiverr/YouTube need at least ~20s; keep it in range.
6. **Start from a template or `plan`, never from a blank spec.**
7. **Everything is 60 fps.** Pick quality deliberately: `draft` for previews,
   `social` for platform caps, `high` for delivery, `master` for re-encoding.
   Use `--fps 30` only if the user explicitly asks for it.
8. **Assets must be inside `allowedAssetRoots`** (default: repo root). Remote
   fetches are disabled unless the user opts in.

## Tool chain

```bash
cd /home/sazar/Projects/framekit
go build -o framekit ./cmd/framekit      # only needed once
./framekit schema                         # machine-readable contract
./framekit templates                      # built-in templates
./framekit themes                         # available palettes
./framekit quality                        # quality/compression profiles
./framekit plan --brief "..." --out spec.json
./framekit use --template reel --vars vars.json --out spec.json
./framekit lint --spec spec.json          # guardrails (errors exit 2)
./framekit analyze --spec spec.json       # metrics as JSON
./framekit validate --spec spec.json
./framekit describe --spec spec.json
./framekit generate --spec spec.json --count 3 --seed 42 --dir variants
./framekit render --spec spec.json --out video.mp4 --quality high
```

`render` lints automatically. Never use `--no-strict` without the user's OK.

If `ffmpeg` is missing, stop and tell the user.

## Workflow

### 1. Understand the brief

Extract: purpose (intro / reel / demo / promo), audience, duration, format, tone
(premium, calm, noir, uplift), whether a voice is wanted, and the CTA/branding.

### 2. Choose format

| Brief mentions | Format |
|---|---|
| Fiverr, YouTube, LinkedIn, intro, demo | `landscape` (1280x720) |
| feed, post, cuadrado | `square` (1080x1080) |
| reel, short, TikTok, portrait, vertical | `portrait` (1080x1920) |

### 3. Draft the spec

Prefer `framekit plan` for the first draft, then edit it. A solid structure is
5 scenes with eyebrows: `01 / IDENTITY`, `02 / STACK`, `03 / PROJECTS`,
`04 / IMPACT`, `05 / CONTACT`.

Use the real portfolio facts (do not invent metrics):
- 29 apps on Google Play · 26 portfolio projects · 300+ GitHub stars
- shiki (331 stars, Rust TUI), oxid (Rust preview envs), leader (~20ms Neovim),
  go-pretty-pdf (Markdown → PDF), Road To (10 Flutter apps), Orca (agent control).
- Stack: Flutter/Dart, React/TypeScript, Go, Rust, Python, Firebase, PostgreSQL, Docker.

### 4. Make it premium

For a premium/cinematic feel set:

```json
"effects": {
  "easing": "easeOutQuint",
  "transition": "wipe",
  "scanline": true,
  "tracking": 2.5,
  "stagger": 0.4,
  "fade": 0.45
}
```

For a calm portfolio intro keep `easeOutCubic`, `bgfade`, `scanline: false`.

### 5. Personalize

- `theme.name`: gruvbox (default), leader, catppuccin, tokyo-night, nord, dracula,
  solarized, light. Or pass hex overrides in `theme`.
- `branding`: `name`, `subtitle`, `meta`, `windowTitle`, `footerLeft`, `footerRight`.
- `fonts`: local TTF paths to replace the embedded Geist Mono.
- `music.progression`: ambient, uplift, noir, minimal. `music.tempo` in BPM.

### 6. Voice (only if requested)

```json
"voices": [{"provider":"cloudflare","text":"...","model":"@cf/facebook/mms-tts-eng","at":0.5,"volume":1}]
```

Requires `CF_ACCOUNT_ID` and `CF_API_TOKEN`. For a local file:
`{"provider":"local","file":"/abs/path.wav","at":2}`.

### 7. Validate, render, verify

```bash
./framekit validate --spec spec.json
./framekit describe --spec spec.json
./framekit render --spec spec.json --out video.mp4
ffprobe -v error -show_entries format=duration,size \
  -show_entries stream=codec_type,codec_name,width,height \
  -of default=noprint_wrappers=1 video.mp4
```

Confirm it has both a video (h264) and audio (aac) stream. Take a screenshot frame
with ffmpeg if visual QA is needed:

```bash
ffmpeg -y -ss 3 -i video.mp4 -frames:v 1 /tmp/frame.png
```

## Common mistakes

- Forgetting `"enabled": true` in `music` → no audio track.
- Setting `at` larger than the scene duration → element never appears.
- Too many items in one scene → they overflow. Split into two scenes.
- Using `style` instead of `progression` for music (both work; prefer progression).
- Expecting text wrapping — it does not exist.

## Deliverable

Report the output path, duration, format, theme and whether it has audio/voice.
Save the final spec next to the video so the user can re-render it.
