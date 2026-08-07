#!/usr/bin/env bash
# Cloud Scheduler quotidien 06:00 Europe/Brussels : POST /internal/vet-news/run.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   VET_NEWS_SECRET=... ./infra/gcp/setup-vet-news-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-vet-news-secret"
JOB_NAME="${PF_SCHED_PREFIX}-vet-news"
SCHEDULE="0 6 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/vet-news/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${VET_NEWS_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with VET_NEWS_SECRET=${SECRET_NAME}:latest (+ VET_NEWS_ENABLED=true) if not yet wired."

HEADERS="Content-Type=application/json,X-Vet-News-Secret=${SECRET}"
pf_scheduler_upsert_http
