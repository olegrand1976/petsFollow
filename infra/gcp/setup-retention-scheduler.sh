#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/retention/run (purge RGPD).
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   RETENTION_PURGE_SECRET=... ./infra/gcp/setup-retention-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-retention-secret"
JOB_NAME="${PF_SCHED_PREFIX}-retention"
SCHEDULE="30 3 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/retention/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${RETENTION_PURGE_SECRET:-}"

RETENTION_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with RETENTION_PURGE_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Retention-Secret=${RETENTION_SECRET}"
pf_scheduler_upsert_http
