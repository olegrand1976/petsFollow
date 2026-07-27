#!/usr/bin/env bash
# Cloud Scheduler : seed staging chaque dimanche 08:00 Europe/Brussels.
# Prérequis : job Cloud Run petsfollow-seed déployé (setup-jobs.sh / postdeploy).
#
# Usage:
#   ./infra/gcp/setup-seed-scheduler.sh
#   SKIP_SEED_NOTIFY_ONCE=1 ./infra/gcp/setup-seed-scheduler.sh   # pas d'annonce immédiate
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

JOB_NAME="petsfollow-seed-weekly"
SCHEDULE="${SEED_SCHEDULE:-0 8 * * 0}"
TZ="Europe/Brussels"
LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
URI="https://run.googleapis.com/v2/projects/${GCP_PROJECT_ID}/locations/${GCP_RUN_REGION}/jobs/petsfollow-seed:run"

echo "=== petsFollow seed scheduler — ${GCP_PROJECT_ID} ==="

# Cloud Scheduler doit pouvoir impersonner le SA OAuth.
PROJECT_NUMBER="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')"
SCHEDULER_SA="service-${PROJECT_NUMBER}@gcp-sa-cloudscheduler.iam.gserviceaccount.com"
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$GCP_PROJECT_ID" \
  --member="serviceAccount:${SCHEDULER_SA}" \
  --role="roles/iam.serviceAccountUser" \
  --quiet >/dev/null || true

gcloud run jobs add-iam-policy-binding petsfollow-seed \
  --project="$GCP_PROJECT_ID" \
  --region="$GCP_RUN_REGION" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.invoker" \
  --quiet >/dev/null || true

if gcloud scheduler jobs describe "$JOB_NAME" --location="$LOCATION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud scheduler jobs update http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$URI" \
    --http-method=POST \
    --oauth-service-account-email="$SA_EMAIL" \
    --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform" \
    --attempt-deadline=1800s
  echo "Updated scheduler job ${JOB_NAME}"
else
  gcloud scheduler jobs create http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$URI" \
    --http-method=POST \
    --oauth-service-account-email="$SA_EMAIL" \
    --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform" \
    --attempt-deadline=1800s
  echo "Created scheduler job ${JOB_NAME}"
fi

echo "Schedule: ${SCHEDULE} (${TZ}) → ${URI}"

if [[ "${SKIP_SEED_NOTIFY_ONCE:-}" != "1" ]]; then
  echo "→ Annonce immédiate (seed-notify, sans truncate DB)"
  # Nécessite une image API déjà déployée avec la commande seed-notify.
  gcloud run jobs execute petsfollow-seed \
    --project="$GCP_PROJECT_ID" \
    --region="$GCP_RUN_REGION" \
    --args=seed-notify \
    --wait --quiet || {
      echo "⚠ seed-notify a échoué — déployer d’abord l’API (push staging), puis relancer ce script." >&2
      exit 1
    }
  echo "Annonce email envoyée (admins / commerciaux / managers + OPS_NOTIFY_EMAIL)."
fi
