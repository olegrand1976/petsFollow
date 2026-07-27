#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/retention/run (purge RGPD).
# Purge les comptes inactifs depuis 3 ans (clients effacés / pros anonymisés) et
# les partages de dossier périmés (lignes + ZIP en bucket).
# Usage:
#   RETENTION_PURGE_SECRET=... ./infra/gcp/setup-retention-scheduler.sh
#   ./infra/gcp/setup-retention-scheduler.sh
#
# Le secret est stocké dans la config du job (header X-Retention-Secret), comme
# auth-health / ai-module-friction. Restreindre scheduler.jobs.get ; ne pas
# afficher la description YAML du job (elle contient le secret en clair).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-retention-secret"
JOB_NAME="petsfollow-retention"
SCHEDULE="30 3 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/retention/run"

if [[ -n "${RETENTION_PURGE_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$RETENTION_PURGE_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$RETENTION_PURGE_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
fi

RETENTION_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with RETENTION_PURGE_SECRET=${SECRET_NAME}:latest if not yet wired."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Retention-Secret=${RETENTION_SECRET}"

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
