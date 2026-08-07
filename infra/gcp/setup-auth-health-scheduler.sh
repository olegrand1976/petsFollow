#!/usr/bin/env bash
# Cloud Scheduler horaire : POST /internal/auth-health/run.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   AUTH_HEALTH_SECRET=... ./infra/gcp/setup-auth-health-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-auth-health-secret"
JOB_NAME="${PF_SCHED_PREFIX}-auth-health"
SCHEDULE="15 * * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/auth-health/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${AUTH_HEALTH_SECRET:-}"

AUTH_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with AUTH_HEALTH_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Auth-Health-Secret=${AUTH_SECRET}"
pf_scheduler_upsert_http
