#!/usr/bin/env bash
# Cloud Scheduler horaire : POST /internal/auth-health/run (clients non vérifiés stuck).
# Usage:
#   AUTH_HEALTH_SECRET=... ./infra/gcp/setup-auth-health-scheduler.sh
#   ./infra/gcp/setup-auth-health-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-auth-health-secret"
JOB_NAME="petsfollow-auth-health"
SCHEDULE="15 * * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/auth-health/run"

if [[ -n "${AUTH_HEALTH_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$AUTH_HEALTH_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=-
  else
    echo -n "$AUTH_HEALTH_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=-
  fi
fi

AUTH_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with AUTH_HEALTH_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Auth-Health-Secret=${AUTH_SECRET}"

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
    --attempt-deadline=120s
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
    --attempt-deadline=120s
  echo "Created scheduler job ${JOB_NAME}"
fi
