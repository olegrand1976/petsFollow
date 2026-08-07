#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/visit-reminders/run (rappel J-1, 17:00 Brussels).
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   VISIT_REMINDERS_SECRET=... ./infra/gcp/setup-visit-reminders-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-visit-reminders-secret"
JOB_NAME="${PF_SCHED_PREFIX}-visit-reminders"
SCHEDULE="0 17 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/visit-reminders/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${VISIT_REMINDERS_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with VISIT_REMINDERS_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Visit-Reminders-Secret=${SECRET}"
pf_scheduler_upsert_http
