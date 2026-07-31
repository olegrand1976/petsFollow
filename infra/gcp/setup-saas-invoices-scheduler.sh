#!/usr/bin/env bash
# Cloud Scheduler mensuel : POST /internal/saas-invoices/run (Flux A brouillons C1).
# Usage:
#   SAAS_INVOICES_SECRET=... ./infra/gcp/setup-saas-invoices-scheduler.sh
#
# Le secret est stocké dans la config du job (header X-Saas-Invoices-Secret).
# Après création du secret SM : redeploy API avec SAAS_INVOICES_SECRET=…:latest.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-saas-invoices-secret"
JOB_NAME="petsfollow-saas-invoices"
# 1er du mois 06:00 Brussels — draft only (API boucle batches limit=50 ; opted-in only) ; envoi Peppol manuel admin.
SCHEDULE="0 6 1 * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/saas-invoices/run"

if [[ -n "${SAAS_INVOICES_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$SAAS_INVOICES_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$SAAS_INVOICES_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
fi

SAAS_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with SAAS_INVOICES_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Saas-Invoices-Secret=${SAAS_SECRET}"

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
