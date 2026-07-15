# Defaults GCP — petsFollow (premedica-prod-2025)

GCP_PROJECT_ID="${GCP_PROJECT_ID:-${PROJECT_ID:-premedica-prod-2025}}"
GCP_RUN_REGION="${GCP_RUN_REGION:-europe-west9}"
GCP_AR_REGION="${GCP_AR_REGION:-europe-west1}"
AR_REGION="${AR_REGION:-${GCP_AR_REGION}}"
VPC_CONNECTOR="${VPC_CONNECTOR:-premedica-connector}"
CLOUDSQL_INSTANCE="${CLOUDSQL_INSTANCE:-premedica-prod-2025:europe-west9:premedica-db-staging}"
SQL_INSTANCE="${SQL_INSTANCE:-premedica-db-staging}"
AR_REPO="${AR_REPO:-petsfollow}"
SERVICE_ACCOUNT="${SERVICE_ACCOUNT:-petsfollow-run}"
API_SERVICE="${API_SERVICE:-petsfollow-api}"
FRONTEND_SERVICE="${FRONTEND_SERVICE:-petsfollow-nuxtjs}"
DB_NAME="${DB_NAME:-petsfollow}"
DB_USER="${DB_USER:-petsfollow_app}"
MIGRATE_USER="${MIGRATE_USER:-petsfollow_migrate}"
REDIS_DB="${REDIS_DB:-14}"
REDIS_KEY_PREFIX="${REDIS_KEY_PREFIX:-petsfollow}"

CUSTOM_DOMAIN="${CUSTOM_DOMAIN:-petsfollow.ll-it-sc.be}"
API_CUSTOM_DOMAIN="${API_CUSTOM_DOMAIN:-api.petsfollow.ll-it-sc.be}"
PUBLIC_SITE_URL="${PUBLIC_SITE_URL:-https://${CUSTOM_DOMAIN}}"
PUBLIC_API_URL="${PUBLIC_API_URL:-https://${API_CUSTOM_DOMAIN}}"

LB_IP="${LB_IP:-34.54.99.89}"
URL_MAP="${URL_MAP:-staging-premedica-care-urlmap}"
HTTPS_PROXY="${HTTPS_PROXY:-staging-premedica-care-proxy}"

FRONTEND_NEG_NAME="${FRONTEND_NEG_NAME:-petsfollow-nuxtjs-neg}"
FRONTEND_BACKEND_NAME="${FRONTEND_BACKEND_NAME:-petsfollow-nuxtjs-backend}"
FRONTEND_PATH_MATCHER="${FRONTEND_PATH_MATCHER:-petsfollow-nuxtjs}"
FRONTEND_CERT_NAME="${FRONTEND_CERT_NAME:-petsfollow-ll-it-sc-cert}"

API_NEG_NAME="${API_NEG_NAME:-petsfollow-api-neg}"
API_BACKEND_NAME="${API_BACKEND_NAME:-petsfollow-api-backend}"
API_PATH_MATCHER="${API_PATH_MATCHER:-petsfollow-api}"
API_CERT_NAME="${API_CERT_NAME:-petsfollow-api-ll-it-sc-cert}"

GITHUB_DEPLOY_SA="${GITHUB_DEPLOY_SA:-github-petsfollow-deploy}"
WIF_POOL_ID="${WIF_POOL_ID:-github-pool}"
WIF_PROVIDER_ID="${WIF_PROVIDER_ID:-github-provider}"
PROJECT_NUMBER="${PROJECT_NUMBER:-237481297060}"
GCP_WORKLOAD_IDENTITY_PROVIDER="${GCP_WORKLOAD_IDENTITY_PROVIDER:-projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${WIF_POOL_ID}/providers/${WIF_PROVIDER_ID}}"

connector_path() {
  echo "projects/${GCP_PROJECT_ID}/locations/${GCP_RUN_REGION}/connectors/${VPC_CONNECTOR}"
}

ar_image() {
  local name="$1"
  local tag="${2:-latest}"
  echo "${GCP_AR_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/${AR_REPO}/${name}:${tag}"
}

api_run_url() {
  gcloud run services describe "$API_SERVICE" \
    --region="${GCP_RUN_REGION}" --project="${GCP_PROJECT_ID}" \
    --format='value(status.url)' 2>/dev/null || true
}

frontend_run_url() {
  gcloud run services describe "$FRONTEND_SERVICE" \
    --region="${GCP_RUN_REGION}" --project="${GCP_PROJECT_ID}" \
    --format='value(status.url)' 2>/dev/null || true
}
