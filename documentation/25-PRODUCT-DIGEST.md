# 25 — Digest produit quotidien (email interne)

Synthèse **fonctionnelle** (non technique) des évolutions du jour, envoyée aux profils :

- `admin`
- `commercial`
- `commercial_manager`

Heure d’envoi : **18:00 Europe/Brussels**.

Le mail précise la **branche / environnement** (ex. `staging`) dans le sujet, l’intro et le corps.

## Flux

```text
17:45 Brussels (approx.)     GitHub Action product-digest.yml
        │  checkout branch (défaut: staging)
        │  git log --since=24h
        ▼
POST /api/v1/internal/product-digest/ingest
        │  header X-Product-Digest-Secret
        │  body: commits + branch + environment
        │  Gemini → résumé FR/EN/NL/ES/ET/IT
        ▼
ops.product_digests (status=ready|empty, meta.branch)

18:00 Europe/Brussels        Cloud Scheduler
        ▼
POST /api/v1/internal/product-digest/run
        │  emails SMTP brandés « … [{branch}] »
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
| GitHub `PRODUCT_DIGEST_API_URL` | ex. API staging Cloud Run |

Payload ingest (optionnel) : `branch`, `environment`. Libellé email = `environment` si fourni, sinon `branch`, sinon `APP_ENV`, sinon `local`.  
Échec SMTP : la ligne d’idempotence est effacée → retry au prochain `run` ; `status=sent` seulement si au moins un envoi OK et zéro échec.

## Déploiement Scheduler

```bash
PRODUCT_DIGEST_SECRET='…' ./infra/gcp/setup-product-digest-scheduler.sh
# puis redéployer l’API pour monter le secret (pf_api_secrets)
```

## Test manuel local

```bash
# Prérequis : make up-infra && make migrate && make seed && make api-dev
# PRODUCT_DIGEST_SECRET dans .env (ou export) ; SMTP → MailHog :8026

# 1. Ingest (Gemini requis si commits non vides)
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/ingest \
  -H "Content-Type: application/json" \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -d '{
    "branch": "staging",
    "environment": "staging",
    "commits":[{"sha":"abc","subject":"feat: rappels soins visibles sur timeline","body":"","author":"dev"}]
  }'

# 2. Envoi
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/run \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}'

# 3. MailHog UI : http://localhost:8027 — sujet contient [staging]
```

Sans Gemini : upsert SQL d’un digest `ready` (voir tests) ou ingest avec `"commits":[]` → `empty` (pas d’email).

## Tables

- `ops.product_digests` — une ligne / jour (`digest_date`) ; `meta` JSON (`branch`, `environment`, `source`, …)
- `ops.product_digest_sends` — idempotence `(digest_date, user_id)`
