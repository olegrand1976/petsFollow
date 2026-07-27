#!/usr/bin/env bash
# Cloud Scheduler bi-quotidien : POST /internal/sales-branches-auto/run
# Crée une branche pour chaque commercial sans branche et non rattaché à un peer,
# puis envoie un email de félicitations (10h et 18h Europe/Brussels).
# Usage:
#   SALES_BRANCHES_AUTO_SECRET=... ./infra/gcp/setup-sales-branches-scheduler.sh
#   ./infra/gcp/setup-sales-branches-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-sales-branches-auto-secret"
JOB_NAME="petsfollow-sales-branches-auto"
SCHEDULE="0 10,18 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/sales-branches-auto/run"

if [[ -n "${SALES_BRANCHES_AUTO_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$SALES_BRANCHES_AUTO_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=-
  else
    echo -n "$SALES_BRANCHES_AUTO_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=-
  fi
fi

AUTO_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with SALES_BRANCHES_AUTO_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Sales-Branches-Auto-Secret=${AUTO_SECRET}"

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
    --attempt-deadline=300s
  echo "Updated scheduler job ${JOB_NAME}"
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
    --attempt-deadline=300s
  echo "Created scheduler job ${JOB_NAME}"
fi
