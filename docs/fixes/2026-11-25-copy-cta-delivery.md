# Copy CTA delivery — 2026-11-25

Production at the start: `cbd558a7d4d25e93e3949b716804ab75c9e8b758`.

The Campaign Manager test returned server ACK for a single copy CTA, but the recipient reported no Pix message. A plain text control through the same provider reached that recipient. This narrows the observed failure to the copy-message path; it does not independently establish the cause.

The upstream [SendButton implementation](https://github.com/evolution-foundation/evolution-go/blob/main/pkg/sendMessage/service/send_service.go) includes an internal copy ID, JSON encoding, DocumentWithCaption envelope, a fresh 32-byte message secret and direct-chat business metadata. This fork's copy path lacked those pieces.

The patch applies these pieces only when a `copy` button is present. Public `displayText`/`copyCode` input is preserved. It does not switch to native `pix` or change carousel/reply paths. Tests cover code and label preservation with JSON escaping, internal ID, envelope contents, fresh secrets and group versus direct-chat metadata. The GHCR workflow runs the message transport tests before building/deploying the commit-tagged image.

Delivery remains pending until a new identified test is received and the copy action is checked on the recipient's device. The old uncertain send is not automatically replayed.

Rollback reference: `ghcr.io/luizglopesx/cm-interactive-provider:cbd558a7d4d25e93e3949b716804ab75c9e8b758`, using the existing deployment workflow.
