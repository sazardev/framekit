# Changelog

## 0.2.0 — 2026-11-17

- Default output is now **1920×1080 @ 60 fps**; added `landscape720` preset.
- All five quality profiles render at 60 fps with x264 `animation` tune,
  bitrate caps and level 4.2 (safe for 1080×1920@60).
- Fast encode: `veryfast` on draft/standard/high/social, `faster` on master.
  Frame painter reuses one `gg.Context` across all frames.
- Added declarative **templates** (`intro`, `reel`, `pitch`, `presentation`,
  `image-showcase`, `quote`) with `{{var:default}}` substitution.
- Added **linter + analyzer**: real text-overflow measurement, WCAG contrast,
  pacing, forbidden words, size estimate, asset/remote security audit.
- Added **guardrails** via `.framekit.json` (`rules` + `security`), strict by
  default in `render` with `--no-strict` override.
- Added `image` scenes (cover/contain, Ken Burns zoom/pan) and `slide` scenes.
- Added `generate` for deterministic A/B variants, `--quality`, `--fps`.
- Added interactive **TUI** (`framekit tui`) and local web **desktop** mode
  (`framekit desktop`).
- Sanitization of all agent-generated text; asset roots and remote fetch
  policies are enforced.

## 0.1.0

- Initial Go framework: gruvbox/Geist Mono frame painter, generated ambient
  music, optional Cloudflare Workers AI voices, MP4 muxing and CLI render.
