#!/usr/bin/env bash
# Bootstrap GCP petsFollow on premedica-prod-2025
# Usage: ./infra/gcp/setup-gcp.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== petsFollow setup GCP — ${GCP_PROJECT_ID} ==="

echo "→ Artifact Registry repo ${AR_REPO}"
if ! gcloud artifacts repositories describe "$AR_REPO" \
  --location="$AR_REGION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud artifacts repositories create "$AR_REPO" \
    --repository-format=docker \
    --location="$AR_REGION" \
    --project="$GCP_PROJECT_ID" \
    --quiet
else
  echo "  Artifact Registry ${AR_REPO} existe"
fi

cat <<EOF

Checklist — avant premier deploy (make gcp-deploy) :

1. Cloud SQL (${SQL_INSTANCE})
   - Base              : ${DB_NAME}
   - Runtime user      : ${DB_USER}
   - Migrate user      : ${MIGRATE_USER}
   - Puis              : projets/infra/shared-postgres/setup-db-protection.sh

2. Redis partagé (DB ${REDIS_DB}, préfixe ${REDIS_KEY_PREFIX}:)
   - Entrée            : projets/infra/shared-redis/redis-apps.conf
   - Puis              : projets/infra/shared-redis/setup-gcp.sh
   - Secret attendu    : petsfollow-redis-url

3. Secret Manager (créer / renseigner) :
   - petsfollow-database-url
   - petsfollow-migrate-database-url
   - petsfollow-jwt-signing-key
   - petsfollow-redis-url

4. Domaines (après services Cloud Run déployés) :
   make gcp-domain
   DNS OVH A → ${LB_IP} :
     - petsfollow     → https://${CUSTOM_DOMAIN}
     - api.petsfollow → https://${API_CUSTOM_DOMAIN}

5. Registres plateforme :
   - projets/infra/database-backup-registry.yaml
   - business-management PlatformApp id=petsfollow

EOF
