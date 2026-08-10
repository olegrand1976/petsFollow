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
cd go && go test ./internal/rag/ ./internal/platform/gemini/ -count=1
cd go && go test ./internal/handlers/ -run 'TestRAG' -count=1
cd nuxtjs && npm test -- tests/unit/locales-parity.spec.ts
```

### Suites suivantes

- Phase 2 : client CrewAI GCP + smoke com
- Phase 3 : SSE Agent Loader
- Phase 4 : bouton « Améliorer IA avancé » + exports
- Phase 5 : recette

Voir aussi [`32-MODULE-IA-CR.md`](32-MODULE-IA-CR.md) (CR IA synchrone conservé).
