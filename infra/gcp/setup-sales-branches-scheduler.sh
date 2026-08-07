#!/usr/bin/env bash
# Cloud Scheduler 10h/18h : POST /internal/sales-branches-auto/run.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   SALES_BRANCHES_AUTO_SECRET=... ./infra/gcp/setup-sales-branches-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-sales-branches-auto-secret"
JOB_NAME="${PF_SCHED_PREFIX}-sales-branches-auto"
SCHEDULE="0 10,18 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/sales-branches-auto/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${SALES_BRANCHES_AUTO_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with SALES_BRANCHES_AUTO_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Sales-Branches-Auto-Secret=${SECRET}"
pf_scheduler_upsert_http
