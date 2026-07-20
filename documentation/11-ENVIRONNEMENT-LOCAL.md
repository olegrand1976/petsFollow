# Environnement local — petsFollow

## Prérequis

Docker, Go, Node (Nuxt), Flutter (face pets), `make`.

## Démarrage

```bash
# Terminal 1 — infra + API
make up-infra && make migrate && make seed && make api-dev
# API http://localhost:8291

# Terminal 2 — Nuxt Pro
make nuxtjs-dev
# http://localhost:3002
```

Après modification des tokens brand : `make brand-sync`.

## Commandes utiles

| Cible | Effet |
|-------|-------|
| `make up-infra` | Postgres + Redis |
| `make migrate` | Migrations SQL |
| `make seed` | Comptes / données démo |
| `make api-dev` | API Go :8291 |
| `make nuxtjs-dev` | Nuxt :3002 |
| `make smoke` | Smoke API |
| `make test-go` / `test-nuxt` / `test-flutter` | Suites unitaires |

## Comptes démo

Voir [AGENTS.md](../AGENTS.md).

Mots de passe : `VetDemo123!` · `ClientDemo123!` · `AdminDemo123!` · `CommercialDemo123!`

## Variables clés

Copier depuis `.env.example`. Billing local : `BILLING_MOCK_ENABLED=true` sans clé Stripe. Google OAuth / 2FA optionnels (`GOOGLE_OAUTH_CLIENT_ID`, `NUXT_PUBLIC_GOOGLE_CLIENT_ID`). Import clients admin : `GEMINI_API_KEY` (mapping colonnes) — voir [24](24-IMPORT-CLIENTS-ADMIN.md).

Médias locaux : `./data/uploads` servi sous `/media/`.

## Staging

[10-GCP-DEPLOIEMENT.md](10-GCP-DEPLOIEMENT.md) · `make smoke-staging` / `make gcp-smoke`.
