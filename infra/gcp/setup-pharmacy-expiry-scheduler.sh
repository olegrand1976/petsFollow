#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/pharmacy/expiry-run (04:00 Brussels).
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   PHARMACY_EXPIRY_SECRET=... ./infra/gcp/setup-pharmacy-expiry-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-pharmacy-expiry-secret"
JOB_NAME="${PF_SCHED_PREFIX}-pharmacy-expiry"
SCHEDULE="0 4 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/pharmacy/expiry-run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${PHARMACY_EXPIRY_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with PHARMACY_EXPIRY_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Pharmacy-Expiry-Secret=${SECRET}"
pf_scheduler_upsert_http
