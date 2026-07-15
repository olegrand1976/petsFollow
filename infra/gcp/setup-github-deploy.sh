#!/usr/bin/env bash
# Workload Identity Federation GitHub Actions → GCP petsFollow.
# Usage: ./infra/gcp/setup-github-deploy.sh [olegrand1976/petsFollow]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

GITHUB_REPO="${1:-olegrand1976/petsFollow}"
SA_EMAIL="${GITHUB_DEPLOY_SA}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
PROJECT_NUMBER="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')"
CB_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== GitHub Actions → GCP petsFollow (${GCP_PROJECT_ID}) ==="
echo "  Repo : ${GITHUB_REPO}"
echo ""

if ! gcloud iam service-accounts describe "$SA_EMAIL" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ Service account ${SA_EMAIL}"
  gcloud iam service-accounts create "$GITHUB_DEPLOY_SA" \
    --project="$GCP_PROJECT_ID" \
    --display-name="GitHub Actions deploy petsFollow"
else
  echo "Service account ${SA_EMAIL} existe déjà"
fi

echo "→ IAM ${SA_EMAIL}"
for role in \
  roles/cloudbuild.builds.editor \
  roles/storage.admin \
  roles/run.admin \
  roles/cloudsql.admin \
  roles/artifactregistry.reader \
  roles/artifactregistry.writer \
  roles/secretmanager.secretAccessor \
  roles/iam.serviceAccountUser \
  roles/serviceusage.serviceUsageConsumer; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="$role" \
    --quiet >/dev/null 2>&1 || true
done

gcloud iam service-accounts add-iam-policy-binding "$CB_SA" \
  --project="$GCP_PROJECT_ID" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser" \
  --quiet >/dev/null 2>&1 || true

RUN_SA="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
if gcloud iam service-accounts describe "$RUN_SA" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud iam service-accounts add-iam-policy-binding "$RUN_SA" \
    --project="$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/iam.serviceAccountUser" \
    --quiet >/dev/null 2>&1 || true
fi

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$GCP_PROJECT_ID" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser" \
  --quiet >/dev/null 2>&1 || true

if ! gcloud iam workload-identity-pools describe "$WIF_POOL_ID" \
  --location=global --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "ERREUR: pool WIF ${WIF_POOL_ID} introuvable." >&2
  exit 1
fi

WIF_PROVIDER="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${WIF_POOL_ID}/providers/${WIF_PROVIDER_ID}"

CURRENT_CONDITION="$(gcloud iam workload-identity-pools providers describe "$WIF_PROVIDER_ID" \
  --workload-identity-pool="$WIF_POOL_ID" \
  --location=global --project="$GCP_PROJECT_ID" \
  --format='value(attributeCondition)' 2>/dev/null || true)"
REPO_ASSERTION="assertion.repository=='${GITHUB_REPO}'"
if [[ -n "$CURRENT_CONDITION" && "$CURRENT_CONDITION" != *"$GITHUB_REPO"* ]]; then
  echo "→ Mise à jour attributeCondition WIF (+ ${GITHUB_REPO})"
  gcloud iam workload-identity-pools providers update-oidc "$WIF_PROVIDER_ID" \
    --workload-identity-pool="$WIF_POOL_ID" \
    --location=global --project="$GCP_PROJECT_ID" \
    --attribute-condition="${CURRENT_CONDITION} || ${REPO_ASSERTION}"
elif [[ -z "$CURRENT_CONDITION" ]]; then
  echo "→ attributeCondition WIF vide — pose ${GITHUB_REPO}"
  gcloud iam workload-identity-pools providers update-oidc "$WIF_PROVIDER_ID" \
    --workload-identity-pool="$WIF_POOL_ID" \
    --location=global --project="$GCP_PROJECT_ID" \
    --attribute-condition="${REPO_ASSERTION}"
else
  echo "  Repo déjà dans attributeCondition WIF"
fi

echo "→ IAM bucket Cloud Build source (${GCP_PROJECT_ID}_cloudbuild)"
CB_BUCKET="gs://${GCP_PROJECT_ID}_cloudbuild"
if gcloud storage buckets describe "$CB_BUCKET" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud storage buckets add-iam-policy-binding "$CB_BUCKET" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/storage.objectAdmin" --quiet >/dev/null 2>&1 || true
  gcloud storage buckets add-iam-policy-binding "$CB_BUCKET" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/storage.legacyBucketWriter" --quiet >/dev/null 2>&1 || true
fi

echo "→ Binding WIF → SA (repo ${GITHUB_REPO})"
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$GCP_PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${WIF_POOL_ID}/attribute.repository/${GITHUB_REPO}" \
  --quiet 2>/dev/null || true

cat <<EOF

OK — valeurs workflow (hardcodées dans deploy-gcp-staging.yml) :

  GCP_WORKLOAD_IDENTITY_PROVIDER
    ${WIF_PROVIDER}

  GCP_SERVICE_ACCOUNT
    ${SA_EMAIL}

EOF
