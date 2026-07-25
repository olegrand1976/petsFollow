#!/usr/bin/env bash
# Déploie les Cloud Run Jobs petsFollow (migrate, seed).
# Usage: ./infra/gcp/setup-jobs.sh [IMAGE_URI]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"
# shellcheck source=lib/deploy-run-args.sh
source "${SCRIPT_DIR}/lib/deploy-run-args.sh"

IMAGE="${1:-$(ar_image api latest)}"
SA="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
CONNECTOR="$(connector_path)"
API_ENV_FILE="$(mktemp)"
SEED_ENV_FILE="$(mktemp)"
trap 'rm -f "$API_ENV_FILE" "$SEED_ENV_FILE"' EXIT

gcloud config set project "$GCP_PROJECT_ID" >/dev/null
pf_write_api_env_file "$API_ENV_FILE" false
pf_write_api_env_file "$SEED_ENV_FILE" true

echo "=== petsFollow Cloud Run Jobs — ${GCP_PROJECT_ID} ==="
echo "Image: ${IMAGE}"

gcloud run jobs deploy petsfollow-migrate \
  --project="$GCP_PROJECT_ID" --image="$IMAGE" --region="$GCP_RUN_REGION" \
  --service-account="$SA" \
  --memory=512Mi --cpu=1 --task-timeout=600 --max-retries=1 \
  --set-cloudsql-instances="$CLOUDSQL_INSTANCE" \
  --vpc-connector="$CONNECTOR" --vpc-egress=private-ranges-only \
  --env-vars-file="$API_ENV_FILE" \
  --set-secrets="$(pf_migrate_secrets)" \
  --command=/app/petsfollow-api --args=migrate \
  --quiet

gcloud run jobs deploy petsfollow-seed \
  --project="$GCP_PROJECT_ID" --image="$IMAGE" --region="$GCP_RUN_REGION" \
  --service-account="$SA" \
  --memory=512Mi --cpu=1 --task-timeout=1800 --max-retries=0 \
  --set-cloudsql-instances="$CLOUDSQL_INSTANCE" \
  --vpc-connector="$CONNECTOR" --vpc-egress=private-ranges-only \
  --env-vars-file="$SEED_ENV_FILE" \
  --set-secrets="$(pf_migrate_secrets)" \
  --command=/app/petsfollow-api --args=seed \
  --quiet

echo "Jobs déployés : petsfollow-migrate, petsfollow-seed"
