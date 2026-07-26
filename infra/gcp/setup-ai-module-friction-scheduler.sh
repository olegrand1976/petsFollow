#!/usr/bin/env bash
# Configure Cloud Scheduler for daily AI CR adhesion drip + friction alerts (09:00 Europe/Brussels).
# Prerequisites: API deployed, secret petsfollow-ai-module-friction-secret created (or set env).
#
# Usage:
#   AI_MODULE_FRICTION_SECRET=... ./infra/gcp/setup-ai-module-friction-scheduler.sh
#   ./infra/gcp/setup-ai-module-friction-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-ai-module-friction-secret"
JOB_NAME="petsfollow-ai-module-friction"
SCHEDULE="0 9 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/ai-module-friction/run"

if [[ -n "${AI_MODULE_FRICTION_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$AI_MODULE_FRICTION_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=-
  else
    echo -n "$AI_MODULE_FRICTION_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=-
  fi
fi

FRICTION_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Reminder: redeploy API with AI_MODULE_FRICTION_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"

HEADERS="Content-Type=application/json,X-Ai-Module-Friction-Secret=${FRICTION_SECRET}"

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
    --attempt-deadline=180s
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
    --attempt-deadline=180s
  echo "Created scheduler job ${JOB_NAME}"
fi

echo "Schedule: ${SCHEDULE} (${TZ}) → ${ENDPOINT}"
echo "Job runs adhesion drip (J0–J90) + friction alerts + trial expiry."
