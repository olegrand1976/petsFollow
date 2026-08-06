# Overrides production — source après / à la place de gcp-env.sh
# Usage : source infra/gcp/lib/gcp-env-prod.sh
#
# Isolation staging ↔ prod : instance SQL, secrets SM *-prod, services Run *-prod,
# bucket médias, préfixe Redis. Même projet GCP / LB / VPC connector.

# shellcheck source=gcp-env.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/gcp-env.sh"

SQL_INSTANCE="${SQL_INSTANCE_PROD:-petsfollow-db-prod}"
CLOUDSQL_INSTANCE="${CLOUDSQL_INSTANCE_PROD:-${GCP_PROJECT_ID}:${GCP_RUN_REGION}:${SQL_INSTANCE}}"
API_SERVICE="${API_SERVICE_PROD:-petsfollow-api-prod}"
FRONTEND_SERVICE="${FRONTEND_SERVICE_PROD:-petsfollow-nuxtjs-prod}"
# Orthanc prod : off par défaut (module tag dev) — pas de service dédié V1 Flutter.
ORTHANC_SERVICE="${ORTHANC_SERVICE_PROD:-petsfollow-orthanc-prod}"

DB_NAME="${DB_NAME_PROD:-petsfollow}"
DB_USER="${DB_USER_PROD:-petsfollow_app}"
MIGRATE_USER="${MIGRATE_USER_PROD:-petsfollow_migrate}"
MIGRATE_JOB="${MIGRATE_JOB_PROD:-petsfollow-migrate-prod}"

# Redis : même VM shared-redis ; DB logique 15 + préfixe clé (l’app n’utilise que Addr+prefix).
REDIS_DB="${REDIS_DB_PROD:-15}"
REDIS_KEY_PREFIX="${REDIS_KEY_PREFIX_PROD:-petsfollow-prod}"

CUSTOM_DOMAIN="${CUSTOM_DOMAIN_PROD:-petsfollow.app}"
API_CUSTOM_DOMAIN="${API_CUSTOM_DOMAIN_PROD:-api.petsfollow.app}"
PUBLIC_SITE_URL="${PUBLIC_SITE_URL_PROD:-https://${CUSTOM_DOMAIN}}"
PUBLIC_API_URL="${PUBLIC_API_URL_PROD:-https://${API_CUSTOM_DOMAIN}}"
CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS_PROD:-${PUBLIC_SITE_URL}}"

GCS_MEDIA_BUCKET="${GCS_MEDIA_BUCKET_PROD:-petsfollow-media-prod}"
# DICOM / PACS off en prod V1
GCS_DICOM_BUCKET="${GCS_DICOM_BUCKET_PROD:-petsfollow-dicom-prod}"

FRONTEND_NEG_NAME="${FRONTEND_NEG_NAME_PROD:-petsfollow-nuxtjs-prod-neg}"
FRONTEND_BACKEND_NAME="${FRONTEND_BACKEND_NAME_PROD:-petsfollow-nuxtjs-prod-backend}"
FRONTEND_PATH_MATCHER="${FRONTEND_PATH_MATCHER_PROD:-petsfollow-nuxtjs-prod}"
API_NEG_NAME="${API_NEG_NAME_PROD:-petsfollow-api-prod-neg}"
API_BACKEND_NAME="${API_BACKEND_NAME_PROD:-petsfollow-api-prod-backend}"
API_PATH_MATCHER="${API_PATH_MATCHER_PROD:-petsfollow-api-prod}"
# Un seul cert pour apex + api (quota SSL global = 10).
FRONTEND_CERT_NAME="${FRONTEND_CERT_NAME_PROD:-petsfollow-prod-domains-cert}"
API_CERT_NAME="${API_CERT_NAME_PROD:-petsfollow-prod-domains-cert}"

# Secret Manager — ne jamais réutiliser les secrets staging (database-url / jwt).
PF_SM_DATABASE_URL="${PF_SM_DATABASE_URL:-petsfollow-prod-database-url}"
PF_SM_MIGRATE_DATABASE_URL="${PF_SM_MIGRATE_DATABASE_URL:-petsfollow-prod-migrate-database-url}"
PF_SM_JWT_SIGNING_KEY="${PF_SM_JWT_SIGNING_KEY:-petsfollow-prod-jwt-signing-key}"
PF_SM_REDIS_URL="${PF_SM_REDIS_URL:-petsfollow-prod-redis-url}"
PF_SM_DB_PASSWORD="${PF_SM_DB_PASSWORD:-petsfollow-prod-db-password}"
PF_SM_MIGRATE_DB_PASSWORD="${PF_SM_MIGRATE_DB_PASSWORD:-petsfollow-prod-migrate-db-password}"
# Billit prod (isolé sandbox staging) — ne monter qu’après whitelist Access Point + contrat.
PF_SM_BILLIT_SECRETS_KEY="${PF_SM_BILLIT_SECRETS_KEY:-petsfollow-prod-billit-secrets-key}"
PF_SM_BILLIT_WEBHOOK_SECRET="${PF_SM_BILLIT_WEBHOOK_SECRET:-petsfollow-prod-billit-webhook-secret}"
PF_SM_BILLIT_MASTER_API_KEY="${PF_SM_BILLIT_MASTER_API_KEY:-petsfollow-prod-billit-master-api-key}"

PETSFOLLOW_GCP_ENV=prod
