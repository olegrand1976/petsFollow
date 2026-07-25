# 25 — Digest produit quotidien (email interne)

Synthèse **fonctionnelle** (non technique) des évolutions du jour, envoyée aux profils :

- `admin`
- `commercial`
- `commercial_manager`

Heure d’envoi : **18:00 Europe/Brussels**.

## Flux

```text
17:45 Brussels (approx.)     GitHub Action product-digest.yml
        │  git log --since=24h
        ▼
POST /api/v1/internal/product-digest/ingest
        │  header X-Product-Digest-Secret
        │  Gemini → résumé FR/EN/NL/ES
        ▼
ops.product_digests (status=ready|empty)

18:00 Europe/Brussels        Cloud Scheduler
        ▼
POST /api/v1/internal/product-digest/run
        │  emails SMTP brandés
        ▼
admin / commercial / commercial_manager
```

Si aucun commit ou aucun impact produit : status `empty` → **pas d’email**.

## Secrets / env

| Variable | Où |
|----------|-----|
| `PRODUCT_DIGEST_SECRET` | API Go + Secret Manager `petsfollow-product-digest-secret` |
| `GEMINI_API_KEY` | déjà requis pour l’ingest (résumé LLM) |
| GitHub `PRODUCT_DIGEST_SECRET` | même valeur |
| GitHub `PRODUCT_DIGEST_API_URL` | ex. `https://api.petsfollow.ll-it-sc.be` |

## Déploiement Scheduler

```bash
PRODUCT_DIGEST_SECRET='…' ./infra/gcp/setup-product-digest-scheduler.sh
# puis redéployer l’API pour monter le secret (pf_api_secrets)
```

## Test manuel local

```bash
# 1. Ingest
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/ingest \
  -H "Content-Type: application/json" \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -d '{"commits":[{"sha":"abc","subject":"feat: rappels soins visibles sur timeline","body":"","author":"dev"}]}'

# 2. Envoi
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/run \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}'
```

Vérifier dans MailHog (`:8026`).

## Tables

- `ops.product_digests` — une ligne / jour (`digest_date`)
- `ops.product_digest_sends` — idempotence `(digest_date, user_id)`
