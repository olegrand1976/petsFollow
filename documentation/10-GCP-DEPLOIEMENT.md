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
| `VAMREG_DRY_RUN` | `true` (défaut) | Jusqu’à P0-1 credentials |
| `PHARMACY_WORKERS_ENABLED` | `false` (défaut) | Enqueue **inline** (timeout détaché) ; `true` = Asynq + Redis |
| Secret `petsfollow-pharmacy-expiry-secret` | Branché si présent | `make gcp-pharmacy-expiry-scheduler` |
| Secret `petsfollow-vamreg-api-key` | Optionnel | → `VAMREG_API_KEY` ; inutile en dry-run |
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

Objectif : API/site **non seedables** pour la piste Play Production et le futur site `petsfollow.app`.

| Face | Service Cloud Run | Domaine cible |
|------|-------------------|---------------|
| Pro (Nuxt) | `petsfollow-nuxtjs-prod` | https://petsfollow.app |
| API (Go) | `petsfollow-api-prod` | https://api.petsfollow.app |

Fichiers déjà en repo (ne pas oublier au go-live) :

| Fichier | Rôle |
|---------|------|
| [`.github/workflows/deploy-gcp-prod.yml`](../.github/workflows/deploy-gcp-prod.yml) | CI deploy prod — **workflow_dispatch seulement** ; décommenter `push: branches: [main]` pour activer |
| [`infra/gcp/cloudbuild-prod.yaml`](../infra/gcp/cloudbuild-prod.yaml) | Build + deploy `APP_ENV=production`, seed off, modules tag-dev off ; garde-fou SQL `FIXME` / refuse staging SQL |
| [`infra/gcp/lib/gcp-env-prod.sh`](../infra/gcp/lib/gcp-env-prod.sh) | Overrides domaines / services `*-prod` / Redis DB 15 |
| [`infra/play/api-bases.sh`](../infra/play/api-bases.sh) | URLs Play : staging vs `api.petsfollow.app` |
| `make play-android-bundle-prod` | AAB Play pointant l’API prod |
| `make gcp-deploy-prod` | Cloud Build prod manuel (même config) |

### Checklist avant d’activer le push `main`

1. **Cloud SQL** prod (instance ≠ `premedica-db-staging`) + **secrets SM dédiés** (ne pas réutiliser `petsfollow-database-url` staging sans review) — remplacer `_CLOUDSQL_INSTANCE` / `FIXME` dans `cloudbuild-prod.yaml` et `gcp-env-prod.sh` ; adapter `pf_api_secrets` / noms SM si secrets séparés.
2. **Bucket GCS** médias prod (`petsfollow-media-prod` ou équivalent) + IAM SA Run.
3. **DNS** `petsfollow.app` + `api.petsfollow.app` → LB + cert managé (NEG/backends `*-prod`).
4. Smoke `https://api.petsfollow.app/health` + login démo **non-seed** (comptes réels / bootstrap one-shot).
5. Décommenter dans `deploy-gcp-prod.yml` :
   ```yaml
   push:
     branches: [main]
   ```
6. Rebuild Play : `make play-android-bundle-prod` → upload piste Production (ou Internal smoke d’abord).
7. Privacy / listing : garder `petsfollow.ll-it-sc.be/legal/*` tant que le site prod n’expose pas `/legal` ; ensuite aligner `LegalUrls` Flutter + Play Console.

Staging reste sur push `staging` → [`deploy-gcp-staging.yml`](../.github/workflows/deploy-gcp-staging.yml) — **ne pas** mutualiser les services Run ni la DB.
