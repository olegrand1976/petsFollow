#!/usr/bin/env bash
# Supprime le Cloud Scheduler de seed staging hebdo (petsfollow-seed-weekly).
# Le reset staging passe uniquement par l’admin Pro (POST /admin/staging/seed)
# ou manuellement : bash infra/gcp/postdeploy.sh --seed
#
# Usage:
#   ./infra/gcp/delete-seed-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

JOB_NAME="petsfollow-seed-weekly"
LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"

echo "=== petsFollow — suppression seed scheduler — ${GCP_PROJECT_ID} ==="

if gcloud scheduler jobs describe "$JOB_NAME" --location="$LOCATION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud scheduler jobs delete "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --quiet
  echo "Deleted scheduler job ${JOB_NAME}"
else
  echo "Scheduler job ${JOB_NAME} absent — rien à faire."
fi

echo "Reset staging : admin Pro (zone danger) ou bash infra/gcp/postdeploy.sh --seed"
