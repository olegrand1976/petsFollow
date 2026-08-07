#!/usr/bin/env bash
# Cloud Scheduler : digest produit quotidien (18:00 Europe/Brussels).
# Staging (défaut) ou prod : PETSFOLLOW_GCP_ENV=prod ./infra/gcp/setup-product-digest-scheduler.sh
#
# Usage:
#   PRODUCT_DIGEST_SECRET=... ./infra/gcp/setup-product-digest-scheduler.sh
#   PETSFOLLOW_GCP_ENV=prod PRODUCT_DIGEST_SECRET=... ./infra/gcp/setup-product-digest-scheduler.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="${PF_SCHED_PREFIX}-product-digest-secret"
JOB_NAME="${PF_SCHED_PREFIX}-product-digest-send"
SCHEDULE="0 18 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/product-digest/run"

pf_scheduler_ensure_secret "$SECRET_NAME" "${PRODUCT_DIGEST_SECRET:-}"

DIGEST_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Reminder: redeploy API with PRODUCT_DIGEST_SECRET=${SECRET_NAME}:latest if not yet wired."

HEADERS="Content-Type=application/json,X-Product-Digest-Secret=${DIGEST_SECRET}"
pf_scheduler_upsert_http
