# Module CR IA avancé (multi-agents + RAG)

Tag **`dev`** — flag `AI_CR_ADVANCED_ENABLED` / `NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED`.

## Phase 1 livrée — Ingestion documentaire RAG

### Rôles

| Composant | Rôle |
|-----------|------|
| **Nuxt** `/admin/rag` | Upload corpus plateforme, modération ajouts cabinet, download, réindex |
| **Nuxt** Paramètres → Base documentaire | Upload cabinet → statut `pending` ; download si fichier encore présent |
| **Go** | Réception fichiers, chunking, embeddings Gemini, `pgvector`, ACL |
| **PostgreSQL** | Schéma `rag.documents` / `rag.chunks` + extension `vector` |

### Corpus hybride

- **Platform** (`scope=platform`) : upload admin → **202** + statut `indexing` (goroutine async, timeout 5 min) → `ready` ou `failed`
- **Practice** (`scope=practice`) : upload cabinet → `pending` → **invisible** au search jusqu’à `POST …/approve` admin → **202** `indexing` → `ready`

### Endpoints

| Méthode | Route |
|---------|--------|
| `GET/POST` | `/api/v1/admin/rag/documents` — POST → **202** `indexing` (index async) |
| `POST` | `/api/v1/admin/rag/documents/{id}/approve` → **202** `indexing` |
| `POST` | `/api/v1/admin/rag/documents/{id}/reject` (purge média source) |
| `DELETE` | `/api/v1/admin/rag/documents/{id}` |
| `GET` | `/api/v1/admin/rag/documents/{id}/download` |
| `POST` | `/api/v1/admin/rag/reindex` — auth admin ; force rebuild jusqu’à 50 docs (`ready` inclus) |
| `GET/POST/DELETE` | `/api/v1/practices/me/rag/documents[/{id}]` — perm `practice.settings` |
| `GET` | `/api/v1/practices/me/rag/documents/{id}/download` — même cabinet uniquement |
| `POST` | `/api/v1/internal/rag/search` + `X-Rag-Reindex-Secret` |
| `POST` | `/api/v1/internal/rag/reindex` + `X-Rag-Reindex-Secret` (jobs ops ; même force rebuild) |

Indexation : goroutine + timeout 5 min ; plafonnée à 80 chunks ; `errorMessage` = codes stables uniquement.

### Infra locale / CI

- Image Postgres : **`pgvector/pgvector:pg16`** (remplace `postgres:16`).
- Volume local existant sans extension : recréer (`docker compose down -v` puis `make up-infra && make migrate`).
- Env : `AI_CR_ADVANCED_ENABLED`, `RAG_REINDEX_SECRET`, `GEMINI_EMBEDDING_MODEL` (défaut `text-embedding-004`, 768 dims).
- `make api-dev` / `nuxtjs-dev` activent le flag.

### Ops Cloud SQL / GCP (one-shot, hors migrate auto)

1. **Extension `vector`** (une fois par instance Cloud SQL, avant ou juste après le premier migrate RAG) :

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

2. **Secret reindex** : créer en Secret Manager le secret job `rag-reindex-secret` (préfixe `petsfollow-` / `petsfollow-prod-` selon env), puis redeploy API avec `--update-secrets` pour monter `RAG_REINDEX_SECRET` (`pf_api_mount_job_secret` dans `infra/gcp/lib/deploy-run-args.sh`). Absent = `/internal/rag/*` en 401.

3. **Flags deploy** : staging `AI_CR_ADVANCED_ENABLED=true` + `NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED: "true"` (jamais écrire `"false"` côté Nuxt) ; prod opt-in explicite.

### Tests

```bash
cd go && go test ./internal/rag/ ./internal/platform/gemini/ ./internal/platform/crewai/ -count=1
cd go && go test ./internal/handlers/ -run 'TestRAG' -count=1
cd nuxtjs && npm test -- tests/unit/locales-parity.spec.ts
```

## Phase 2 livrée — Orchestrateur CrewAI partagé + client Go

Service **multi-apps** (petsFollow, Vantura, futures) — repo dédié [`../crewai-orchestrator`](../../crewai-orchestrator/README.md) (hors monorepo petsFollow).

### Contrat

| Méthode | Path | Auth |
|---------|------|------|
| `GET` | `/health` | IAM Cloud Run (ID token) ; local = none |
| `GET` | `/openapi.json` | IAM Cloud Run ; local = none |
| `POST` | `/v1/tasks/submit` | IAM + `X-Crew-Secret` (+ `X-Crew-API-Version: 1`) |
| `GET` | `/v1/tasks/{id}/events` | **501** (Phase 3 SSE) |

Boot fail-closed : `CREW_SHARED_SECRET` obligatoire sauf `ALLOW_INSECURE_AUTH=true` (local only).

Envelope :

```json
{
  "workflow": "petsfollow_cr_improve",
  "tenant": "petsfollow",
  "correlationId": "uuid",
  "payload": { "sourceText": "...", "targetLocale": "fr", "practiceId": "", "countryCode": "BE" }
}
```

Workflows : `staging_smoke_test` · `petsfollow_cr_improve` (Tri → Clinicien+RAG → Rédacteur) · `vantura_credit_stub`.

Tool RAG : orchestrateur → `POST {PETSFOLLOW_API_BASE}/api/v1/internal/rag/search` + `X-Rag-Reindex-Secret` (pas de DSN Postgres dans le crew).

### Client petsFollow

- Package [`go/internal/platform/crewai`](../go/internal/platform/crewai/client.go) : `Health`, `SubmitTask` (+ ID token IAM si `CREWAI_USE_ID_TOKEN`)
- Env API : `CREWAI_BASE_URL`, `CREWAI_SHARED_SECRET`, `CREWAI_USE_ID_TOKEN`
- Staging : URL orchestrateur + `CREWAI_USE_ID_TOKEN=true` par défaut dans `deploy-run-args.sh`
- Smoke : `make crewai-smoke` (skip si URL vide ; ID token si `CREWAI_USE_ID_TOKEN`)
- Deploy : secret monté depuis `petsfollow-*-crewai-shared-secret` ou fallback staging `CREWAI_WEBHOOK_SECRET_STAGING`

### Deploy orchestrateur staging

```bash
cd ../crewai-orchestrator && bash infra/gcp/deploy-staging.sh
```

Cloud Run : `crewai-orchestrator-staging` · région `europe-west9` · **minScale=0** · **IAM-only** + secret obligatoire.

### Clôture ops Phase 2

1. `bash infra/gcp/deploy-staging.sh` (repo orchestrateur)
2. Smoke : `CREWAI_BASE_URL=<url> CREWAI_SHARED_SECRET=… CREWAI_USE_ID_TOKEN=true make crewai-smoke`
3. RAG live optionnel : monter `RAG_REINDEX_SECRET` sur l’orchestrateur (`RAG_REINDEX_SECRET_SM=petsfollow-rag-reindex-secret`) une fois le secret SM créé

### Suites suivantes

- Phase 3 : SSE Agent Loader
- Phase 4 : bouton « Améliorer IA avancé » + exports
- Phase 5 : recette

Voir aussi [`32-MODULE-IA-CR.md`](32-MODULE-IA-CR.md) (CR IA synchrone conservé).
