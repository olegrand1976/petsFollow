# petsFollow — identité visuelle

Marque dual-face unifiée par le **mark circulaire** (chien + mains + pulse + wordmark).

| Asset | Chemin |
|-------|--------|
| Source | [`logo/petsfollow-mark.png`](logo/petsfollow-mark.png) |
| Nuxt Pro | `nuxtjs/public/brand/logo-mark.png` (+ favicon / apple-touch) |
| Flutter pets | `flutter/assets/brand/petsfollow-mark.png` (+ icônes Android/iOS) |
| PDF CR | `go/internal/platform/consultationpdf/assets/emblem.png` (sync via `make brand-sync`) |

| Face | App | Mode |
|------|-----|------|
| **petsFollow Pro** | `nuxtjs/` | Light |
| **petsFollow pets** | `flutter/` | Dark |

Tokens source : `tokens/design-tokens.json` → `make brand-sync`.

Prérequis sync mark : **Pillow** (`pip install Pillow`) — le script normalise le mark en PNG réel (UI 512 px, PDF 256 px) pour gofpdf.
