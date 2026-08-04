#!/usr/bin/env bash
# Cloud Scheduler quotidien : POST /internal/visit-reminders/run (rappel J-1).
# Envoie push + SMS aux clients dont un RDV confirmé tombe dans la fenêtre
# now+1h .. now+VISIT_REMINDER_LOOKAHEAD_HOURS (défaut 30h). À 17:00 Europe/Brussels,
# cela couvre exactement les RDV du lendemain ; l'idempotence est garantie côté DB
# (index unique sms_log_reminder_once sur visite+créneau), donc un run rejoué ou
# manqué reste sans effet de bord.
# Usage:
#   VISIT_REMINDERS_SECRET=... ./infra/gcp/setup-visit-reminders-scheduler.sh
#   ./infra/gcp/setup-visit-reminders-scheduler.sh
#
# Le secret est stocké dans la config du job (header X-Visit-Reminders-Secret),
# comme retention / auth-health. Restreindre scheduler.jobs.get ; ne pas afficher
# la description YAML du job (elle contient le secret en clair).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SECRET_NAME="petsfollow-visit-reminders-secret"
JOB_NAME="petsfollow-visit-reminders"
SCHEDULE="0 17 * * *"
TZ="Europe/Brussels"
API_URL="${PUBLIC_API_URL%/}"
ENDPOINT="${API_URL}/api/v1/internal/visit-reminders/run"

if [[ -n "${VISIT_REMINDERS_SECRET:-}" ]]; then
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$VISIT_REMINDERS_SECRET" | gcloud secrets versions add "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$VISIT_REMINDERS_SECRET" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
fi

REMINDERS_SECRET="$(gcloud secrets versions access latest \
  --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID")"

echo "→ Redeploy API with VISIT_REMINDERS_SECRET=${SECRET_NAME}:latest if not yet wired."
echo "→ Reminder SMS only leave the system when SMS_ENABLED=true AND SMS_DRY_RUN=false."

LOCATION="${GCP_SCHEDULER_LOCATION:-europe-west1}"
HEADERS="Content-Type=application/json,X-Visit-Reminders-Secret=${REMINDERS_SECRET}"

if gcloud scheduler jobs describe "$JOB_NAME" --location="$LOCATION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud scheduler jobs update http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$ENDPOINT" \
    --http-method=POST \
    --headers="$HEADERS" \
    --message-body='{}' \
    --attempt-deadline=300s \
    >/dev/null
  echo "Updated scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ})"
else
  gcloud scheduler jobs create http "$JOB_NAME" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --schedule="$SCHEDULE" \
    --time-zone="$TZ" \
    --uri="$ENDPOINT" \
    --http-method=POST \
    --headers="$HEADERS" \
    --message-body='{}' \
    --attempt-deadline=300s \
    >/dev/null
  echo "Created scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ})"
fi
