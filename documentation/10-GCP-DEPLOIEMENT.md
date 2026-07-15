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
- Entrées registres : `projets/infra` (backup YAML, Redis, grants) + BM `PlatformApp` `petsfollow`

## Commandes

```bash
make gcp-setup    # AR + checklist secrets / DB / Redis
make gcp-deploy   # Cloud Build → images + deploy Run
make gcp-domain   # NEG + backends + host rules + certs managés
make gcp-smoke    # smoke contre api.petsfollow.ll-it-sc.be
```

Pipeline GitHub : push branche `staging` → [`.github/workflows/deploy-gcp-staging.yml`](../.github/workflows/deploy-gcp-staging.yml) (WIF).

## DNS OVH (zone `ll-it-sc.be`)

| Hôte | Type | Cible |
|------|------|-------|
| `petsfollow` | A | `34.54.99.89` |
| `api.petsfollow` | A | `34.54.99.89` |

Attendre certificats ACTIVE :

```bash
gcloud compute ssl-certificates describe petsfollow-ll-it-sc-cert --global --format='yaml(managed)'
gcloud compute ssl-certificates describe petsfollow-api-ll-it-sc-cert --global --format='yaml(managed)'
```

## Ressources LB

- URL map : `staging-premedica-care-urlmap`
- HTTPS proxy : `staging-premedica-care-proxy`
- Certs : `petsfollow-ll-it-sc-cert`, `petsfollow-api-ll-it-sc-cert`
