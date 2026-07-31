# 40 — PACS Orthanc (imagerie DICOM)

Module **tag `dev`** : lecture / upload DICOM vétérinaire via Orthanc sur Cloud Run (scale-to-zero).

## Architecture

| Couche | Rôle |
|--------|------|
| Cloud Run `petsfollow-orthanc` | Orthanc HTTP `:8080`, min=0 max=10, ingress **all** + IAM (pas d’invoker public) |
| Cloud SQL DB `orthanc` | Index Orthanc (plugin PostgreSQL, `IndexConnectionsCount=4`) |
| GCS `petsfollow-dicom` | Stockage `.dcm` (plugin Google Cloud Storage) + **versioning** |
| API Go | Orchestration status/wake, proxy authé, logs admin, lien `imaging.pet_studies` |
| Redis | Cache statut PACS + ring buffer logs |
| Nuxt Pro | Onglet **Imagerie** fiche animal + `/admin/pacs` |

Le navigateur **ne parle jamais** à Orthanc : uniquement via BFF → Go → Orthanc (IAM `run.invoker` + identity token en GCP).

## Flags

| Env | Défaut |
|-----|--------|
| `PACS_ENABLED` | staging si URL Orthanc · prod `false` · `make api-dev` → `true` |
| `NUXT_PUBLIC_PACS_ENABLED` | idem (`make nuxtjs-dev`) |
| `PACS_ORTHANC_URL` | URL Orthanc (`http://localhost:8042` en local) |
| `PACS_ORTHANC_USE_ID_TOKEN` | `true` en GCP · `false` en local |

404 `pacs_disabled` si flag off.

## Endpoints API

| Méthode | Path | Accès |
|---------|------|-------|
| `GET` | `/api/v1/pacs/status` | auth (véto…) — timeout Orthanc 2s + cache Redis |
| `POST` | `/api/v1/pacs/wake` | auth — cold start |
| `GET/POST` | `/api/v1/pets/{id}/pacs/studies` | practice perm + pet access |
| `GET` | `/api/v1/pacs/studies\|series\|instances/…` | proxy Orthanc |
| `GET` | `/api/v1/admin/pacs/logs` | admin |
| `GET` | `/api/v1/admin/pacs/metrics` | admin |

## Démo locale (présentation)

```bash
make up-infra && make up-pacs && make migrate && make seed
make api-dev          # T1 — PACS_ORTHANC_URL=http://localhost:8042
make nuxtjs-dev       # T2 — http://localhost:3002
make pacs-demo-seed   # upload RX démo sur pet client.demo
```

Scénario pitch (~3 min) :

1. Login `vet.demo@petsfollow.test` / `VetDemo123!`
2. Client Sophie → animal actif → onglet **Imagerie** (`?tab=imaging`)
3. Badge état `ready` · étude « RX thorax demo » · outils zoom / pan / W/L
4. Badge `dev` visible (module non GA)

Orthanc local : `http://127.0.0.1:8042` (basic `petsfollow` / `petsfollow`, bind **loopback only**, config [`infra/orthanc/orthanc.local.json`](../infra/orthanc/orthanc.local.json)). Fixture : `testdata/pacs/demo-rx.dcm` (`scripts/gen-minimal-dicom.py`). Reset stockage local : `docker compose -f deploy/docker-compose.yml --env-file .env rm -sf orthanc && docker volume rm deploy_orthancdata` (nom exact via `docker volume ls | grep orthanc`).

## Infra

**Staging Cloud Build** (`infra/gcp/cloudbuild.yaml`) : build Kaniko **`executor:debug`** (shell `/busybox/sh` pour `_SKIP_ORTHANC`) → image `orthanc:$BUILD_ID` → `deploy-orthanc` (`setup-orthanc.sh`, **best-effort** — échec Orthanc ≠ échec API) → `deploy-api` / `deploy-frontend` résolvent `PACS_ORTHANC_URL` via `orthanc_run_url`. Opt-out build+deploy : `_SKIP_ORTHANC=true`. Re-deploys : mode **deploy-only** auto (pas de reset mot de passe SQL). Orthanc Cloud Run : **`ingress=all` + IAM** (ID token) — `ingress=internal` casse l’appel API→Orthanc quand l’API a `vpc-egress=private-ranges-only`.

Manuel :

```bash
# Build image
docker build -f deploy/Dockerfile.orthanc -t $(ar_image orthanc latest) .
# Deploy bucket + DB + Cloud Run
./infra/gcp/setup-orthanc.sh $(ar_image orthanc latest)
```

Puis monter `PACS_ORTHANC_URL` / `PACS_ORTHANC_PASSWORD` sur l’API (voir `pf_api_secrets`) — automatique en staging via Cloud Build.

## Sécurité & backups

- Ingress Orthanc = **all** + IAM Cloud Run (ID token API→Orthanc) ; auth Orthanc désactivée derrière IAM (local : `ORTHANC_AUTH_ENABLED=true`).
- Proxy DICOM : IDs Orthanc liés à `imaging.pet_studies` du **cabinet** (anti-IDOR cross-tenant).
- Backups : **Cloud SQL automated backups** (DB `orthanc`) + **GCS versioning** sur `petsfollow-dicom`.
- RGPD : export JSON `imagingStudies` ; purge `DELETE /me` / rétention → delete Orthanc study (best-effort) après collect artifacts.
- Pas de DIMSE (port 4242) en V1 — Cloud Run HTTP only ; ingestion via upload `.dcm` Pro.
- Pooling : `IndexConnectionsCount=4` ; max-instances Orthanc=10 → surveiller `max_connections` Cloud SQL.
- Staging : `PACS_ENABLED` auto-on **uniquement** si `PACS_ORTHANC_URL` est défini (sinon offline UI évité).
- Viewer V1 = previews Orthanc PNG (mesures en pixels non calibrées) — pas de Cornerstone WASM.
- Cold start : TTL cache `starting` = 90 s ; UI upload poll jusqu’à `ready`.

## UI

- Fiche animal → onglet Imagerie → `PacsViewerContainer` (badge état, wake, upload, dual-pane).
- Viewer canvas (preview Orthanc) : Zoom, Pan, Window/Level, mesure, flèche, plein écran, comparaison.
- Admin `/admin/pacs` : métriques + logs (poll 5s), badge `nav.tagDev`.

## Tests

```bash
# Go (DB requise pour intégration)
cd go && go test ./internal/handlers/ -run 'TestPacs' -count=1

# Vitest
cd nuxtjs && npm test -- tests/unit/pacsPoll.spec.ts

# Playwright (API + Nuxt + flag on)
make test-e2e-p0   # inclut 20-pacs-admin + 20b-pacs-imaging @p0
```

## Hors V1

C-STORE modalité, partage client des images, PgBouncer dédié, conformité multi-pays / archivage légal long terme.
