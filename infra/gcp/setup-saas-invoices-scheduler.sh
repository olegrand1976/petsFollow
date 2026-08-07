#!/usr/bin/env bash
# Cloud Scheduler mensuel : POST /internal/saas-invoices/run (1er du mois 06:00).
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   SAAS_INVOICES_SECRET=... ./infra/gcp/setup-saas-invoices-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-saas-invoices-secret"
JOB_NAME="${PF_SCHED_PREFIX}-saas-invoices"
SCHEDULE="0 6 1 * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/saas-invoices/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${SAAS_INVOICES_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with SAAS_INVOICES_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Saas-Invoices-Secret=${SECRET}"
pf_scheduler_upsert_http
