#!/usr/bin/env bash
# Cloud Scheduler mensuel : POST /internal/afmps-import/run (1er du mois 05:00 Brussels).
# Gate 1 seule (staging + notif ops) — gates 2–3 restent humaines sur /admin/afmps-imports.
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod …
# Usage:
#   AFMPS_IMPORT_SECRET=... ./infra/gcp/setup-afmps-import-scheduler.sh
# Prérequis : CSV pack licencié déposé sur gs://$GCS_MEDIA_BUCKET/afmps-imports/latest.csv
#
# Note : ne pas combiner day-of-month + day-of-week (OR GCP) — ex. « 1-7 * 1 » ≠ 1er lundi.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-afmps-import-secret"
JOB_NAME="${PF_SCHED_PREFIX}-afmps-import"
# 1er du mois à 05:00 Brussels (même pattern que saas-invoices).
SCHEDULE="0 5 1 * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/afmps-import/run"
# Parse + COPY ~2.7k CNK : deadline élargie (max Scheduler HTTP = 1800s).
ATTEMPT_DEADLINE="${ATTEMPT_DEADLINE:-540s}"

pf_scheduler_ensure_secret "$SECRET_NAME" "${AFMPS_IMPORT_SECRET:-}"

SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with AFMPS_IMPORT_SECRET=${SECRET_NAME}:latest if not yet wired."
echo "→ Déposer le CSV : gsutil cp pack.csv gs://\$GCS_MEDIA_BUCKET/afmps-imports/latest.csv"

HEADERS="Content-Type=application/json,X-Afmps-Import-Secret=${SECRET}"
pf_scheduler_upsert_http
