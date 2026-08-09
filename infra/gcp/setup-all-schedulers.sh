#!/usr/bin/env bash
# Provisionne (create/update) tous les Cloud Schedulers petsFollow.
# Staging (défaut) ou prod :
#   ./infra/gcp/setup-all-schedulers.sh
#   PETSFOLLOW_GCP_ENV=prod ./infra/gcp/setup-all-schedulers.sh
#
# Les secrets doivent déjà exister dans Secret Manager (ou être passés en env
# pour création : PRODUCT_DIGEST_SECRET, RETENTION_PURGE_SECRET, …).
# Après création des secrets prod : redeploy API (PETSFOLLOW_GCP_ENV=prod).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_LABEL="${PETSFOLLOW_GCP_ENV:-staging}"

echo "=== petsFollow schedulers — env=${ENV_LABEL} ==="

run() {
  local script="$1"
  echo ""
  echo "→ ${script}"
  bash "${SCRIPT_DIR}/${script}"
}

run setup-product-digest-scheduler.sh
run setup-product-digest-weekly-scheduler.sh
run setup-retention-scheduler.sh
run setup-auth-health-scheduler.sh
run setup-pharmacy-expiry-scheduler.sh
run setup-sales-branches-scheduler.sh
run setup-invoicing-reconcile-scheduler.sh
# setup-saas-invoices-scheduler.sh : Flux A en sommeil (INVOICING_SAAS_ENABLED off).
# L'abonnement Pro se facture hors application — lancer le script à la main si le flux est réveillé.
run setup-research-etl-scheduler.sh
run setup-vet-news-scheduler.sh
run setup-ai-module-friction-scheduler.sh
run setup-visit-reminders-scheduler.sh

echo ""
echo "=== Done (${ENV_LABEL}). Vérifier : gcloud scheduler jobs list --location=europe-west1 --filter='name~petsfollow' ==="
