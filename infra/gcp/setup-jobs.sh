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
pf_write_api_env_file "$API_ENV_FILE" false false staging
# APP_ENV=staging obligatoire : les jobs seed / seed-mass refusent tout autre environnement.
pf_write_api_env_file "$SEED_ENV_FILE" true false staging
# Email staff après seed manuel CLI (admins / commerciaux / managers).
{
  echo "SEED_NOTIFY_STAFF: \"true\""
} >>"$SEED_ENV_FILE"

pf_seed_secrets() {
  local secrets
  secrets="$(pf_migrate_secrets)"
  if gcloud secrets versions access latest \
    --secret=petsfollow-smtp-password --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},SMTP_PASS=petsfollow-smtp-password:latest"
  fi
  printf '%s' "$secrets"
}

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
  --set-secrets="$(pf_seed_secrets)" \
  --command=/app/petsfollow-api --args=seed \
  --quiet

# Densification démo (additif, après seed). Pas de truncate. Idempotent.
gcloud run jobs deploy petsfollow-seed-mass \
  --project="$GCP_PROJECT_ID" --image="$IMAGE" --region="$GCP_RUN_REGION" \
  --service-account="$SA" \
  --memory=1Gi --cpu=1 --task-timeout=1800 --max-retries=0 \
  --set-cloudsql-instances="$CLOUDSQL_INSTANCE" \
  --vpc-connector="$CONNECTOR" --vpc-egress=private-ranges-only \
  --env-vars-file="$API_ENV_FILE" \
  --set-secrets="$(pf_seed_secrets)" \
  --command=/app/petsfollow-api --args=seed-mass \
  --quiet

echo "Jobs déployés : petsfollow-migrate, petsfollow-seed, petsfollow-seed-mass"
