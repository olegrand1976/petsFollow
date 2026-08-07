#!/usr/bin/env bash
# Cloud Scheduler quotidien 09:00 : POST /internal/ai-module-friction/run.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   AI_MODULE_FRICTION_SECRET=... ./infra/gcp/setup-ai-module-friction-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-ai-module-friction-secret"
JOB_NAME="${PF_SCHED_PREFIX}-ai-module-friction"
SCHEDULE="0 9 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/ai-module-friction/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${AI_MODULE_FRICTION_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with AI_MODULE_FRICTION_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Ai-Module-Friction-Secret=${SECRET}"
pf_scheduler_upsert_http
