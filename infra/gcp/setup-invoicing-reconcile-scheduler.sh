#!/usr/bin/env bash
# Cloud Scheduler horaire : POST /internal/invoicing-reconcile/run.
# Relit chez Billit les factures restées en `sending` (webhook manqué) — sans ce
# rattrapage elles restent en vol, quota consommé, jusqu'au rejet à 7 jours.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   INVOICING_RECONCILE_SECRET=... ./infra/gcp/setup-invoicing-reconcile-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-invoicing-reconcile-secret"
JOB_NAME="${PF_SCHED_PREFIX}-invoicing-reconcile"
SCHEDULE="20 * * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/invoicing-reconcile/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${INVOICING_RECONCILE_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with INVOICING_RECONCILE_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Invoicing-Reconcile-Secret=${SECRET}"
pf_scheduler_upsert_http
