#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/pharmacy/expiry-run
# Auto-quarantaine des lots périmés + digest email le lundi (décidé côté API).
# Usage:
#   PHARMACY_EXPIRY_SECRET=... ./infra/gcp/setup-pharmacy-expiry-scheduler.sh
#   ./infra/gcp/setup-pharmacy-expiry-scheduler.sh
#
# Le secret est stocké dans la config du job (header X-Pharmacy-Expiry-Secret).
# Après création du secret SM : redeploy API avec PHARMACY_EXPIRY_SECRET=…:latest.
# Restreindre scheduler.jobs.get ; ne pas afficher le YAML du job (secret en clair).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-pharmacy-expiry-secret"
JOB_NAME="petsfollow-pharmacy-expiry"
# Quotidien 04:00 Brussels (après retention 03:30) — digest lundi géré par l'API.
SCHEDULE="0 4 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/pharmacy/expiry-run"

if [[ -n "${PHARMACY_EXPIRY_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$PHARMACY_EXPIRY_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$PHARMACY_EXPIRY_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
fi

EXPIRY_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with PHARMACY_EXPIRY_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Pharmacy-Expiry-Secret=${EXPIRY_SECRET}"

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
    --attempt-deadline=300s \
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
    --attempt-deadline=300s \
    >/dev/null
  echo "Created scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ})"
fi
