# Copy CTA delivery — 2026-11-25

- `e0b8f50` wrapped the copy CTA in `DocumentWithCaptionMessage` (+ `bot biz_bot`), copying upstream. In production WhatsApp rejected it with `server returned error 473` (same as upstream issue #110). Rolled back to `cbd558a`.
- With `cbd558a`, copy CTA reaches customers (PIX read on 06/11 and 23/11), but an iPhone recipient got nothing while the server ACKed (`sent`).
- This change reverts `e0b8f50` and only adds a fresh 32-byte `MessageContextInfo.MessageSecret` when a `copy` button is present (required by iOS to render interactive messages). No wrapper, no `bot` node, same `cta_copy` params and `biz` node. Public `/send/button` contract unchanged; carousel/reply/pix paths untouched.
- Validate on an iPhone and an Android recipient. Rollback: `ghcr.io/luizglopesx/cm-interactive-provider:cbd558a7d4d25e93e3949b716804ab75c9e8b758`.
