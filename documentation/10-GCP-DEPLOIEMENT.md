# GCP — déploiement petsFollow (staging)

Projet partagé : `premedica-prod-2025` · région Run : `europe-west9` · LB : `34.54.99.89`.

| Face | Service Cloud Run | Domaine |
|------|-------------------|---------|
| Pro (Nuxt) | `petsfollow-nuxtjs` | https://petsfollow.ll-it-sc.be |
| API (Go) | `petsfollow-api` | https://api.petsfollow.ll-it-sc.be |

Infra partagée : Cloud SQL `premedica-db-staging` (DB `petsfollow`), Redis VM `shared-redis` (DB **14**), VPC connector `premedica-connector`. Pattern domaine = LB Premedica + Serverless NEG (comme Kore).

**Pharmacie (S6)** — voir aussi [28](28-PLAN-STOCK-PEREMPTION.md) · [37](37-ROADMAP-STOCK-FACTURATION.md) :

Migrations pharmacie récentes (ordre) : `000107` jobs_audit → `000108` pricing/reorder → `000109` orders/BL → `000110` inventaire → `000111` withdrawal → `000112` food_chain (après `000100` visit_soft_delete / `000101` consultation_share). Si une base locale a déjà appliqué d’anciens numéros `000100_pharmacy_*` / `000103–106`, préférer DB propre ou `seed` reset. La migration `000081` exige **`pg_trgm`**.

| Élément | Staging | Notes |
|---------|---------|--------|
| Extension **`pg_trgm`** | Une fois sur Cloud SQL (`cloudsqlsuperuser` / Extensions) | Requis migrate `000081+` |
| `PHARMACY_ENABLED` | `true` (défaut `deploy-run-args` si `APP_ENV=staging`) | Prod = opt-in `false` |
| `VAMREG_DRY_RUN` | **forcé `true`** (staging + prod) | Déclarations dry-run jusqu’à ICD write + P0-1 ; pas `false` avec clé listes |
| `VAMREG_AFMPS_BASE_URL` | `https://app.fagg-afmps.be/vamreg/api` | Listes readonly ICD |
| `PHARMACY_WORKERS_ENABLED` | `false` (défaut) | Enqueue **inline** (timeout détaché) ; `true` = Asynq + Redis |
| Secret `petsfollow-pharmacy-expiry-secret` | Branché si présent | `make gcp-pharmacy-expiry-scheduler` |
| Secret `petsfollow-afmps-import-secret` | Branché si présent | `make gcp-afmps-import-scheduler` (mensuel gate 1) ; CSV `gs://$GCS_MEDIA_BUCKET/afmps-imports/latest.csv` |
| `RESEARCH_ENABLED` | `true` staging (défaut) ; prod opt-in | Observatoire tag `dev` — [42](42-RESEARCH.md) |
| Secrets `petsfollow-research-etl-secret` + `petsfollow-research-anon-salt` | Branchés si présents | `make gcp-research-etl-scheduler` (6 h Brussels) |
| Secret `petsfollow-vamreg-afmps-api-key` | → `VAMREG_AFMPS_API_KEY` | Listes GET (`FAMHP-SEC-KEY`) ; `./infra/gcp/setup-vamreg-afmps-secret.sh` |
| Secret `petsfollow-vamreg-api-key` | **Non monté** tant que dry-run | Write déclarant futur (P0-1) — voir [39](39-VAMREG-AFMPS-READONLY.md) |
| Smoke pilote | Manuel | Receipt → DAF finalize antibio → `vamregStatus=sent` (dry-run) |

DAF→Billit (S5) **gelé** tant que reseller Billit absent.

## Prérequis

- `gcloud` authentifié sur `premedica-prod-2025`
- Artifact Registry repo `petsfollow` (`europe-west1`)
- Secrets SM : `petsfollow-database-url`, `petsfollow-migrate-database-url`, `petsfollow-jwt-signing-key`, `petsfollow-redis-url`
- Bucket GCS médias : `petsfollow-media` (`make gcp-setup` ou legacy `make gcp-setup-media`) + env Cloud Run `GCS_MEDIA_BUCKET=petsfollow-media`
- Entrées registres : `projets/infra` (backup YAML, Redis, grants) + BM `PlatformApp` `petsfollow`

## Commandes

```bash
make gcp-setup         # AR + bucket médias + checklist secrets / DB / Redis
make gcp-github        # SA GitHub + WIF
make gcp-deploy        # Cloud Build → images + deploy Run
make gcp-domain        # NEG + backends + host rules + certs managés
make gcp-smoke         # smoke contre api.petsfollow.ll-it-sc.be

# legacy (bucket médias uniquement)
# make gcp-setup-media
```

Pipeline GitHub : push branche `staging` → [`.github/workflows/deploy-gcp-staging.yml`](../.github/workflows/deploy-gcp-staging.yml) (WIF).

### Seed DB staging

Le seed **n’est plus** exécuté à chaque deploy ni via Scheduler. Remise à zéro **manuelle uniquement** :

1. **Admin Pro** → tableau de bord → zone danger (saisie `RESET STAGING`) — API `POST /api/v1/admin/staging/seed` (flag `ADMIN_STAGING_SEED_ENABLED`).
2. **CLI** : `bash infra/gcp/postdeploy.sh --seed` (job Cloud Run `petsfollow-seed`).

**Conservé au reset** : tickets support (`ops.support_tickets` + replies) ; comptes admin / commercial / commercial_manager ; client staging `b.murgo1976@gmail.com` (compte + pets / messagerie / liens cabinet / billing, rebranchés sur les cabinets démo re-seedés par nom).

```bash
make gcp-delete-seed-scheduler        # retire le job Scheduler hebdo s’il existe encore
bash infra/gcp/postdeploy.sh --seed   # reset CLI (+ email si SEED_NOTIFY_STAFF)
# Annonce seule (sans truncate) :
# gcloud run jobs execute petsfollow-seed --region=europe-west9 --args=seed-notify --wait
```

### Cleanup artefacts quality (post Playwright)

Après chaque run du workflow staging, le job `cleanup-quality` (`if: always()` — même si deploy ou Playwright échouent) purge **toutes** les écritures smoke/e2e : messages (+ média `e2e media`), BP/FC/poids/labos, visites + CR e2e, salles `Box e2e` / sites `E2E Antenne`, prospects CRM, tickets support `E2E support`, lots pharmacie `E2E-*`/`S6-*` (+ DAF/BL/inventaire), users éphémères (`smoke*`, `register+`, `pw-*`, `hr-dur-*`, …) et cabinets orphelins, docs factu draft. Exclusions : `saas_master` et factures delivered/issued (Billit live). Les smoke manuels (`make gcp-smoke`, `smoke-pharmacy-s6-staging`, `postdeploy.sh`) enchaînent la même purge automatiquement. Garde-fou : `nuxtjs/tests/unit/e2e-cleanup-coverage.spec.ts` (préfixes emails e2e ↔ SQL). Manuel :

```bash
make gcp-staging-quality-cleanup
# local Docker / proxy TCP :
PF_CLEANUP_TARGET=staging DATABASE_URL=… make staging-quality-cleanup
# dry-run : CLEANUP_ARGS='--dry-run' make gcp-staging-quality-cleanup
```

Stripe Live : voir checklist [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md) + `./infra/gcp/setup-stripe-secrets.sh`.

## DNS OVH (zone `ll-it-sc.be`)

| Hôte | Type | Cible |
|------|------|-------|
| `petsfollow` | A | `34.54.99.89` |
| `api.petsfollow` | A | `34.54.99.89` |

Attendre certificats ACTIVE :

```bash
gcloud compute ssl-certificates describe petsfollow-domains-cert --global --format='yaml(managed)'
```

## Ressources LB

- URL map : `staging-premedica-care-urlmap`
- HTTPS proxy : `staging-premedica-care-proxy`
- Certs : `petsfollow-domains-cert` (petsfollow + api.petsfollow)

## Vérification Cloud Run (sans DNS)

```bash
curl -s https://petsfollow-api-a7ako2njea-od.a.run.app/health
PETSFOLLOW_API_URL=https://petsfollow-api-a7ako2njea-od.a.run.app make gcp-smoke
```

Les certificats managés restent en `PROVISIONING` tant que les enregistrements A OVH ne pointent pas vers `${LB_IP}`.

---

## Production (branche `main`)

Objectif : API/site **non seedables** pour la piste Play Production et le site `petsfollow.app` (Flutter-first : API prioritaire).

| Face | Service Cloud Run | Domaine cible |
|------|-------------------|---------------|
| Pro (Nuxt) | `petsfollow-nuxtjs-prod` | https://petsfollow.app |
| API (Go) | `petsfollow-api-prod` | https://api.petsfollow.app |

| Ressource | Valeur |
|-----------|--------|
| Cloud SQL | instance `petsfollow-db-prod` (≠ staging) · DB `petsfollow` |
| Redis | même VM `shared-redis` · URL SM `petsfollow-prod-redis-url` (`/15`) · préfixe `petsfollow-prod:` |
| Bucket médias | `petsfollow-media-prod` |
| Secrets SM | `petsfollow-prod-database-url`, `petsfollow-prod-migrate-database-url`, `petsfollow-prod-jwt-signing-key`, … |
| Cert LB | `petsfollow-prod-domains-cert` (apex + api) — quota SSL global = 10 |

### Commandes

```bash
make gcp-setup-prod    # SQL + users + secrets *-prod + bucket
make gcp-deploy-prod   # Cloud Build → migrate + petsfollow-api-prod + nuxtjs-prod
make gcp-domain-prod   # NEG / backends / host rules / cert (après DNS OVH)
```

Fichiers :

| Fichier | Rôle |
|---------|------|
| [`.github/workflows/deploy-gcp-prod.yml`](../.github/workflows/deploy-gcp-prod.yml) | CI deploy prod — **workflow_dispatch seulement** ; décommenter `push: branches: [main]` pour activer ; **smoke post-deploy** (`SMOKE_PROFILE=prod`) — pas de suite CI complète (filet = PR + staging) |
| [`infra/gcp/cloudbuild-prod.yaml`](../infra/gcp/cloudbuild-prod.yaml) | Build + deploy `APP_ENV=production`, seed off, modules tag-dev off ; refuse SQL staging |
| [`infra/gcp/setup-gcp-prod.sh`](../infra/gcp/setup-gcp-prod.sh) | Bootstrap SQL / secrets / bucket |
| [`infra/gcp/lib/gcp-env-prod.sh`](../infra/gcp/lib/gcp-env-prod.sh) | Overrides domaines / services `*-prod` / Redis |
| [`infra/play/api-bases.sh`](../infra/play/api-bases.sh) | URLs Play : staging vs `api.petsfollow.app` |
| `make play-android-bundle-prod` | AAB Play pointant l’API prod |
| `make gcp-deploy-prod` | Cloud Build prod manuel |

### DNS OVH (zone `petsfollow.app`) — manuel

Même LB que le staging : **`34.54.99.89`**.

| Hôte (OVH) | Type | Cible | Résultat |
|------------|------|-------|----------|
| `@` (ou champ vide / apex) | **A** | `34.54.99.89` | `petsfollow.app` → Nuxt prod |
| `api` | **A** | `34.54.99.89` | `api.petsfollow.app` → API prod |
| `www` (optionnel) | **CNAME** | `petsfollow.app.` | redirection / alias |

Pas d’enregistrement `media.*` : médias = GCS.

Après création DNS :

```bash
make gcp-domain-prod
# Attendre ACTIVE (souvent 15–60 min après propagation DNS) :
gcloud compute ssl-certificates describe petsfollow-prod-domains-cert --global --format='yaml(managed)'
curl -fsS https://api.petsfollow.app/health
```

### Checklist go-live

1. `make gcp-setup-prod` (une fois) — vérifier instance `petsfollow-db-prod` RUNNABLE.
2. DNS OVH ci-dessus.
3. `make gcp-deploy-prod` puis `make gcp-domain-prod`.
4. Smoke `make smoke-prod` (ou workflow job `smoke`) — health/ready/plans **sans** comptes seed ni écritures.
5. Flutter : `make play-android-bundle-prod` → Internal puis Production.
6. (Plus tard) décommenter `push: main` dans `deploy-gcp-prod.yml`.
7. Privacy / listing : garder `petsfollow.ll-it-sc.be/legal/*` tant que `/legal` n’est pas servi sur `petsfollow.app` ; ensuite aligner `LegalUrls` Flutter + Play Console.

**VAMReg / AFMPS (prod)** — même politique que staging ([39](39-VAMREG-AFMPS-READONLY.md)) :

- `pf_write_api_env_file` force `VAMREG_DRY_RUN=true` et pose `VAMREG_AFMPS_BASE_URL`.
- `pf_api_secrets` monte `petsfollow-vamreg-afmps-api-key` → `VAMREG_AFMPS_API_KEY` (pas de `VAMREG_API_KEY` write tant que dry-run).
- Au go-live prod : réutiliser le secret SM (ou clone dédié) + `./infra/gcp/setup-vamreg-afmps-secret.sh --attach-run` sur `petsfollow-api-prod` si le service existe déjà hors Cloud Build.

Staging reste sur push `staging` → [`deploy-gcp-staging.yml`](../.github/workflows/deploy-gcp-staging.yml) — **ne pas** mutualiser les services Run ni la DB.
