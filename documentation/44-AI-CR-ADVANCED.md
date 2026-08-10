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
| `GET` | `/v1/tasks/{id}/events` | IAM + secret — SSE (Phase 3) |
| `POST` | `/v1/tasks/submit` | IAM + `X-Crew-Secret` (+ `X-Crew-API-Version: 1`) ; `"async": true` → 202 |

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

## Phase 3 livrée — SSE Agent Loader + improve-advanced

### Flux

1. `POST /api/v1/visits/{visitID}/report/improve-advanced` → `rag.improve_runs` + CrewAI `async=true` → **202** `{ runId }`
2. `GET …/improve-advanced/{runId}/events` → SSE (`step` / `warning` / `final` / `error` / `ping`)
3. `POST …/improve-advanced/{runId}/cancel` → best-effort
4. BFF flat : `/api/visits/{id}/report-improve-advanced` (+ `/…/{runId}/events` stream, `/cancel`)
5. UI : bouton `visit-report-improve-advanced` (flag Nuxt) + `ProAgentLoader`

Orchestrateur : `GET /v1/tasks/{id}/events` (SSE) ; submit avec `"async": true` → 202 `running`.

**Ops V1** : le hub SSE côté API Go est **in-proc** ; si l’EventSource atterrit sur une autre instance Cloud Run, le client **poll** `rag.improve_runs` jusqu’au statut terminal (pas de live thought cross-instance). Orchestrateur staging : `max-instances=1` (bus SSE mémoire). Prod multi-instance → Redis/PubSub plus tard.

### Clôture Phase 3

- **RGPD** : export `GET /me/export` → clé `ragImproveRuns` ; tombstone Pro (`DeleteProAccount`) purge `rag.improve_runs` (les CR `visit_reports` restent).
- **E2E** : Playwright `@p1` `03h-visit-report-improve-advanced.spec.ts` (mocks POST + SSE, pas de CrewAI live).
- **Smoke ops** (manuel) : `CREWAI_BASE_URL=… CREWAI_SHARED_SECRET=… CREWAI_USE_ID_TOKEN=true make crewai-smoke`.

### Suites suivantes

- Phases 1–4 : **livrées** (voir sections ci-dessus).
- Phase 5 : **livrée** — runbook ops + checklist recette + C2.28 (pas de badge GA : tag `dev` conservé).

### Phase 4 — Exports CR

| Surface | Détail |
|---------|--------|
| API | `GET /api/v1/visits/{visitID}/report/pdf` — auth + `canManageVisit` ; PDF branded (`consultationpdf.BuildPDF`) ; attachment |
| GET report | champ optionnel `lastImproveCitations` (dernier `improve_runs` `completed`, flag advanced on) |
| BFF | `/api/visits/{visitId}/report-pdf` (`proxyBinary`) |
| UI | `visit-report-copy-md` / `download-md` / `download-pdf` + liste `visit-report-citations` |
| Utils | `nuxtjs/utils/visit-report-export.ts` |

## Phase 5 livrée — Recette, runbook & DoD

Tag **`dev`** inchangé jusqu’à décision GA explicite. Pas d’use case commercial dédié (règle modules tag `dev`) : la démo VetPro reste sur **Améliorer (IA)** classique ([`UC-VP-04`](../useCase/01-vetpro/UC-VP-04-nouvelle-consultation.md)).

### Runbook ops

| Action | Commande / détail |
|--------|-------------------|
| Flags | API `AI_CR_ADVANCED_ENABLED` · Nuxt `NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED` (staging on ; **prod off** par défaut) |
| Secrets API | `CREWAI_BASE_URL`, `CREWAI_SHARED_SECRET`, `CREWAI_USE_ID_TOKEN` · `RAG_REINDEX_SECRET` |
| Smoke orchestrateur | `CREWAI_BASE_URL=… CREWAI_SHARED_SECRET=… CREWAI_USE_ID_TOKEN=true make crewai-smoke` |
| Deploy orchestrateur | `cd ../crewai-orchestrator && bash infra/gcp/deploy-staging.sh` (`max-instances=1`) |
| Rollback immédiat | flags advanced → `false` / retirer env Nuxt (BFF 404 `*_disabled`, bouton masqué) — improve classique inchangé |
| Extension Cloud SQL | `CREATE EXTENSION IF NOT EXISTS vector;` (one-shot) |

### Dette connue (hors V1)

- Hub SSE API **in-proc** + poll DB cross-instance ; Redis/PubSub plus tard.
- Cancel CrewAI remote best-effort (statut DB `cancelled` côté petsFollow).
- Monitoring GCP (p50/p95, tokens, budget) : pas de dashboard dédié — s’appuyer sur logs Cloud Run + table `rag.improve_runs` (`latency_ms`, `error_code`, `status`).

### Checklist recette staging (C2.28)

Jouer sur un cabinet seed (`vet.demo`) avec flag on + au moins un doc RAG `ready` (platform ou practice approved).

| # | Jeu | Attendu |
|---|-----|---------|
| 1 | Notes synthétiques → **Améliorer IA avancé** | ≥3 steps loader → CR éditable ; **pas** auto-finalize |
| 2 | Improve **classique** après / à côté | Toujours OK (non-régression) |
| 3 | RAG hit (posologie dans corpus) | Citation visible (`visit-report-citations`) ou section références |
| 4 | RAG miss / query hors corpus | Warning ou absence de dose inventée |
| 5 | Doc cabinet `pending` | Absent du search / des citations |
| 6 | Locale `nl` ou `auto` | Corps / titres cohérents |
| 7 | Exports | Copy MD · download MD · PDF (`report-pdf`) |
| 8 | Kill-switch flag off | Bouton avancé masqué ; POST → 404 |
| 9 | Ops | `make crewai-smoke` green sur staging |

### Protocole praticiens (go-live soft)

1. 10 CR réels **anonymisés** (pas de PHI en ticket / Slack).
2. Score utilité (1–5) + noter si édition manuelle avant finalize.
3. Critère soft : **≥ 80 %** « utilisables avec retouches mineures ».
4. Sign-off : date + auteur dans le ticket ops / note de release (pas de merge GA sans ça).

### Auto (filet)

```bash
cd go && go test ./internal/handlers/ -run 'TestImproveAdvanced|TestVisitReportPDF|TestRAG' -count=1
cd nuxtjs && npm test -- tests/unit/visit-report-export.spec.ts tests/unit/locales-parity.spec.ts
# e2e P1 (API + Nuxt up) :
cd nuxtjs && npx playwright test tests/e2e/specs/03h-visit-report-improve-advanced.spec.ts tests/e2e/specs/03i-visit-report-export.spec.ts
```

Voir aussi [`32-MODULE-IA-CR.md`](32-MODULE-IA-CR.md) (CR IA synchrone conservé).
