#!/usr/bin/env bash
# Cloud Scheduler toutes les 6h : POST /internal/research-etl/run.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   RESEARCH_ETL_SECRET=... RESEARCH_ANON_SALT=... ./infra/gcp/setup-research-etl-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-research-etl-secret"
JOB_NAME="${PF_SCHED_PREFIX}-research-etl"
SCHEDULE="0 */6 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/research-etl/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${RESEARCH_ETL_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with RESEARCH_ETL_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Research-Etl-Secret=${SECRET}"
pf_scheduler_upsert_http
