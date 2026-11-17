---
description: Create a premium flat-design video with framekit ( landscapes, square, portrait )
---

Load the `framekit-video` skill and create a video for the user.

Brief: $ARGUMENTS

Steps:
1. Read `.agents/skills/framekit-video/SKILL.md`.
2. If the brief is empty, ask for: purpose, format, duration, tone and CTA.
3. Run `framekit plan --brief "<brief>" --out spec.json` in `/home/sazar/Projects/framekit`.
4. Edit the spec: personalize `theme`, `branding`, `effects` and real portfolio facts.
5. Run `validate` and `describe`, then `render`.
6. Verify with ffprobe that the MP4 has h264 + aac streams.
7. Report output path, duration, format, theme and audio status.

Do not enable Cloudflare voices unless the user explicitly asked for a voice.
