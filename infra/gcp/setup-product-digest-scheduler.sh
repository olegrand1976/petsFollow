#!/usr/bin/env bash
# Configure Cloud Scheduler for daily product digest send (18:00 Europe/Brussels).
# Prerequisites: API deployed, secret petsfollow-product-digest-secret created.
#
# Usage:
#   PRODUCT_DIGEST_SECRET=... ./infra/gcp/setup-product-digest-scheduler.sh
#   # or rely on Secret Manager:
#   ./infra/gcp/setup-product-digest-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-product-digest-secret"
JOB_NAME="petsfollow-product-digest-send"
SCHEDULE="0 18 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/product-digest/run"

if [[ -n "${PRODUCT_DIGEST_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$PRODUCT_DIGEST_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=-
  else
    echo -n "$PRODUCT_DIGEST_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=-
  fi
fi

DIGEST_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

# Ensure Cloud Run API has the secret mounted.
echo "→ Reminder: redeploy API with PRODUCT_DIGEST_SECRET=${SECRET_NAME}:latest if not yet wired."

# Scheduler needs App Engine app or explicit location for some projects;
# use cloud-scheduler in the Run region.
LOCATION="${GCP_SCHEDULER_LOCATION:-${GCP_RUN_REGION}}"

HEADERS="Content-Type=application/json,X-Product-Digest-Secret=${DIGEST_SECRET}"

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

echo "Schedule: ${SCHEDULE} (${TZ}) → ${ENDPOINT}"
