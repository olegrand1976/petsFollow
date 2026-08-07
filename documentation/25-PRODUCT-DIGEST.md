# 25 — Digest produit (quotidien + hebdo) & Nouveautés

Synthèse **fonctionnelle** (non technique) des évolutions petsFollow.

## Surfaces

| Canal | Quand | Audience |
|-------|-------|----------|
| Email quotidien | **18:00** Europe/Brussels | `admin` / `commercial` / `commercial_manager` |
| Email hebdo | **samedi 08:00** Europe/Brussels | idem **+** responsables cabinet (`reference_vet`) |
| UI Pro `/nouveautes` | à la demande | vet / assist / secretary / admin / commercial / manager |

Le mail précise la **branche / environnement**. Sur staging / local / test : suffixe localisé **« environnement de TEST »** (ex. `staging — environnement de TEST`).

## Flux quotidien

```text
17:45 Brussels (approx.)     GitHub Action product-digest.yml
        │  matrix: staging (branch staging) + production (branch main)
        │  git log --since=24h
        ▼
POST /api/v1/internal/product-digest/ingest
        │  header X-Product-Digest-Secret
        │  body: commits + branch + environment
        │  Gemini → résumé FR/EN/NL/ES/ET/IT/UK/RU
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

## Flux hebdo (samedi)

```text
Samedi 08:00 Brussels        Cloud Scheduler
        ▼
POST /api/v1/internal/product-digest/weekly-run
        │  agrège digests ready/sent du lundi ISO → aujourd’hui (Europe/Brussels)
        │  skip si semaine vide / aucun destinataire hors *.petsfollow.test
        ▼
admin / commercial / commercial_manager / reference_vet (team_members active)
```

Idempotence : `ops.product_digest_weekly_sends` (`week_start` = lundi ISO Europe/Brussels — même borne que l’agrégat).

Quotidien et hebdo : destinataires `*.petsfollow.test` exclus.

## UI Nouveautés

- Page Pro : `/nouveautes` (nav « Nouveautés »)
- API auth : `GET /api/v1/product-digests?limit=30`
- BFF Nuxt : `GET /api/product-digests`

## Secrets / env

| Variable | Où |
|----------|-----|
| `PRODUCT_DIGEST_SECRET` | API Go + Secret Manager `petsfollow-product-digest-secret` (prod : `petsfollow-prod-product-digest-secret`) |
| `GEMINI_API_KEY` | déjà requis pour l’ingest (résumé LLM) |
| GitHub `PRODUCT_DIGEST_SECRET` / `PRODUCT_DIGEST_API_URL` | staging |
| GitHub `PRODUCT_DIGEST_SECRET_PROD` / `PRODUCT_DIGEST_API_URL_PROD` | production |

Payload ingest (optionnel) : `branch`, `environment`. Libellé email = `environment` si fourni, sinon `branch`, sinon `APP_ENV`, sinon `local` — puis tag TEST si non-prod.  
Échec SMTP : la ligne d’idempotence est effacée → retry au prochain `run` ; `status=sent` seulement si au moins un envoi OK et zéro échec (quotidien).

## Déploiement Scheduler

```bash
# Staging (défaut)
PRODUCT_DIGEST_SECRET='…' make gcp-product-digest-scheduler
PRODUCT_DIGEST_SECRET='…' make gcp-product-digest-weekly-scheduler
# ou tous les jobs :
make gcp-all-schedulers

# Production (jobs petsfollow-prod-*, API api.petsfollow.app)
PETSFOLLOW_GCP_ENV=prod PRODUCT_DIGEST_SECRET='…' make gcp-all-schedulers
# puis redéployer l’API prod pour monter les secrets petsfollow-prod-*
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

# 2. Envoi quotidien
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/run \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}'

# 3. Envoi hebdo
curl -sS -X POST http://localhost:8291/api/v1/internal/product-digest/weekly-run \
  -H "X-Product-Digest-Secret: $PRODUCT_DIGEST_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}'

# 4. MailHog UI : http://localhost:8027 — sujet contient [staging — environnement de TEST]
```

Sans Gemini : upsert SQL d’un digest `ready` (voir tests) ou ingest avec `"commits":[]` → `empty` (pas d’email).

## Tables

- `ops.product_digests` — une ligne / jour (`digest_date`) ; `meta` JSON (`branch`, `environment`, `source`, …)
- `ops.product_digest_sends` — idempotence quotidienne `(digest_date, user_id)`
- `ops.product_digest_weekly_sends` — idempotence hebdo `(week_start, user_id)`
