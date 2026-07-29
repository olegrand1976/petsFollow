# Overrides GCP petsFollow — environnement **production** (branche main).
# Source AFTER infra/gcp/lib/gcp-env.sh :
#   source infra/gcp/lib/gcp-env.sh
#   source infra/gcp/lib/gcp-env-prod.sh
#
# Domaines brand petsfollow.app (support@petsfollow.app).
# À compléter avant le 1er deploy prod : Cloud SQL, secrets SM, LB/DNS.

CUSTOM_DOMAIN="${CUSTOM_DOMAIN:-petsfollow.app}"
API_CUSTOM_DOMAIN="${API_CUSTOM_DOMAIN:-api.petsfollow.app}"
PUBLIC_SITE_URL="${PUBLIC_SITE_URL:-https://${CUSTOM_DOMAIN}}"
PUBLIC_API_URL="${PUBLIC_API_URL:-https://${API_CUSTOM_DOMAIN}}"

# Services Cloud Run dédiés (ne pas écraser le staging petsfollow-api / petsfollow-nuxtjs).
API_SERVICE="${API_SERVICE:-petsfollow-api-prod}"
FRONTEND_SERVICE="${FRONTEND_SERVICE:-petsfollow-nuxtjs-prod}"

# TODO avant activation : instance SQL + DB prod (ne pas réutiliser premedica-db-staging).
# Exemple : premedica-prod-2025:europe-west9:petsfollow-db-prod
CLOUDSQL_INSTANCE="${CLOUDSQL_INSTANCE:-premedica-prod-2025:europe-west9:FIXME-petsfollow-db-prod}"
SQL_INSTANCE="${SQL_INSTANCE:-FIXME-petsfollow-db-prod}"
DB_NAME="${DB_NAME:-petsfollow}"

# Redis : DB/prefix distincts du staging (14 / petsfollow).
REDIS_DB="${REDIS_DB:-15}"
REDIS_KEY_PREFIX="${REDIS_KEY_PREFIX:-petsfollow_prod}"

# Médias : bucket dédié recommandé (éviter mélange PHI staging/prod).
GCS_MEDIA_BUCKET="${GCS_MEDIA_BUCKET:-petsfollow-media-prod}"

# Cert / NEG : noms distincts du staging (à créer via setup-custom-domain adapté).
FRONTEND_NEG_NAME="${FRONTEND_NEG_NAME:-petsfollow-nuxtjs-prod-neg}"
FRONTEND_BACKEND_NAME="${FRONTEND_BACKEND_NAME:-petsfollow-nuxtjs-prod-backend}"
FRONTEND_PATH_MATCHER="${FRONTEND_PATH_MATCHER:-petsfollow-nuxtjs-prod}"
FRONTEND_CERT_NAME="${FRONTEND_CERT_NAME:-petsfollow-app-domains-cert}"
API_NEG_NAME="${API_NEG_NAME:-petsfollow-api-prod-neg}"
API_BACKEND_NAME="${API_BACKEND_NAME:-petsfollow-api-prod-backend}"
API_PATH_MATCHER="${API_PATH_MATCHER:-petsfollow-api-prod}"
API_CERT_NAME="${API_CERT_NAME:-petsfollow-app-domains-cert}"
