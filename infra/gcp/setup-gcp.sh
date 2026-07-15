#!/usr/bin/env bash
# Bootstrap GCP petsFollow on premedica-prod-2025
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null
echo "→ Artifact Registry repo petsfollow"
gcloud artifacts repositories describe petsfollow --location="$AR_REGION" 2>/dev/null || \
  gcloud artifacts repositories create petsfollow --repository-format=docker --location="$AR_REGION"
echo "→ Cloud SQL database petsfollow (manual if not exists on ${CLOUDSQL_INSTANCE})"
echo "→ Secrets: petsfollow-database-url, petsfollow-jwt-signing-key"
echo "Bootstrap checklist printed — configure secrets in Secret Manager before deploy."
