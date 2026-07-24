# GCP — déploiement petsFollow (staging)

Projet partagé : `premedica-prod-2025` · région Run : `europe-west9` · LB : `34.54.99.89`.

| Face | Service Cloud Run | Domaine |
|------|-------------------|---------|
| Pro (Nuxt) | `petsfollow-nuxtjs` | https://petsfollow.ll-it-sc.be |
| API (Go) | `petsfollow-api` | https://api.petsfollow.ll-it-sc.be |

Infra partagée : Cloud SQL `premedica-db-staging` (DB `petsfollow`), Redis VM `shared-redis` (DB **14**), VPC connector `premedica-connector`. Pattern domaine = LB Premedica + Serverless NEG (comme Kore).

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
