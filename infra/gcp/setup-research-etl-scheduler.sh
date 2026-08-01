#!/usr/bin/env bash
# Cloud Scheduler : POST /internal/research-etl/run (ETL observatoire Research).
# Usage:
#   RESEARCH_ETL_SECRET=... RESEARCH_ANON_SALT=... ./infra/gcp/setup-research-etl-scheduler.sh
#   ./infra/gcp/setup-research-etl-scheduler.sh
#
# Crée/met à jour les secrets SM (etl + anon salt) et le job Scheduler.
# Après création : redeploy API pour monter les secrets (pf_api_secrets).
# Restreindre scheduler.jobs.get ; ne pas afficher le YAML du job (secret en clair).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

ETL_SECRET_NAME="petsfollow-research-etl-secret"
ANON_SECRET_NAME="petsfollow-research-anon-salt"
JOB_NAME="petsfollow-research-etl"
# Toutes les 6 h Europe/Brussels
SCHEDULE="0 */6 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/research-etl/run"

upsert_secret() {
  local name="$1"
  local value="$2"
  if [[ -z "$value" ]]; then
    return 0
  fi
  if gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$value" | gcloud secrets versions add "$name" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$value" | gcloud secrets create "$name" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
}

upsert_secret "$ETL_SECRET_NAME" "${RESEARCH_ETL_SECRET:-}"
upsert_secret "$ANON_SECRET_NAME" "${RESEARCH_ANON_SALT:-}"

ETL_SECRET="$(gcloud secrets versions access latest \
  --secret="$ETL_SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with RESEARCH_ETL_SECRET=${ETL_SECRET_NAME}:latest and RESEARCH_ANON_SALT=${ANON_SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Research-Etl-Secret=${ETL_SECRET}"

if gcloud scheduler jobs describe "$JOB_NAME" --location="$LOCATION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud scheduler jobs update http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$ENDPOINT" \
    --http-method=POST \
    --headers="$HEADERS" \
    --message-body='{}' \
    --attempt-deadline=540s \
    >/dev/null
  echo "Updated scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ})"
else
  gcloud scheduler jobs create http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$ENDPOINT" \
    --http-method=POST \
    --headers="$HEADERS" \
    --message-body='{}' \
    --attempt-deadline=540s \
    >/dev/null
  echo "Created scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ})"
fi
