#!/usr/bin/env bash
# Crée le bucket GCS médias (avatars / photos animaux) + IAM.
# Usage: ./infra/gcp/setup-gcs-media.sh
#
# Note GCP : les conditions IAM sur allUsers sont refusées
# (PublicResourceAllowConditionCheck). La protection PHI visit-reports/
# repose donc sur l’app : Upload sans URL publique + stream auth + clés UUID.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

BUCKET="${GCS_MEDIA_BUCKET:-petsfollow-media}"
LOCATION="${GCS_MEDIA_LOCATION:-${GCP_RUN_REGION}}"
SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== petsFollow GCS media — gs://${BUCKET} (${LOCATION}) ==="

if gcloud storage buckets describe "gs://${BUCKET}" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "  Bucket gs://${BUCKET} existe déjà"
else
  gcloud storage buckets create "gs://${BUCKET}" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --uniform-bucket-level-access \
    --no-public-access-prevention
  echo "  Bucket gs://${BUCKET} créé"
fi

echo "→ IAM objectAdmin pour ${SA_EMAIL}"
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin" \
  --quiet >/dev/null

# Avatars / photos : lecture publique (URLs storage.googleapis.com).
# GCP interdit une condition allUsers qui exclurait visit-reports/ —
# les objets PHI n’exposent pas d’URL publique (media.Upload → "").
echo "→ IAM objectViewer public (allUsers) pour médias non-PHI"
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member="allUsers" \
  --role="roles/storage.objectViewer" \
  --quiet >/dev/null 2>&1 || true

echo ""
echo "Configurer Cloud Run API :"
echo "  GCS_MEDIA_BUCKET=${BUCKET}"
echo "Done. PHI visit-reports : pas d’URL publique (API Open auth uniquement)."
echo "Avertissement : un objet visit-reports/* reste lisible si son chemin UUID est connu"
echo "  (UBLA + allUsers). Mitigation : clés aléatoires + pas d’URL retournée + purge finalize."
