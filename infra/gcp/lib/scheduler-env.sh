# Scheduler env bootstrap — staging (défaut) ou prod (PETSFOLLOW_GCP_ENV=prod).
# Usage (dans setup-*-scheduler.sh) :
#   source "${SCRIPT_DIR}/lib/scheduler-env.sh"
#   # définit PF_SCHED_PREFIX, PUBLIC_API_URL, GCP_PROJECT_ID, …
#   SECRET_NAME="${PF_SCHED_PREFIX}-product-digest-secret"
#   JOB_NAME="${PF_SCHED_PREFIX}-product-digest-send"
#
# Staging : petsfollow-*
# Prod    : petsfollow-prod-*  + api.petsfollow.app

_SCHED_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "${PETSFOLLOW_GCP_ENV:-}" == "prod" ]]; then
  # shellcheck source=gcp-env-prod.sh
  source "${_SCHED_LIB_DIR}/gcp-env-prod.sh"
  PF_SCHED_PREFIX="petsfollow-prod"
else
  # shellcheck source=gcp-env.sh
  source "${_SCHED_LIB_DIR}/gcp-env.sh"
  PF_SCHED_PREFIX="petsfollow"
fi

pf_scheduler_ensure_secret() {
  # $1 = SECRET_NAME, $2 = env var value (optional — skip create/add if empty)
  local secret_name="$1"
  local secret_value="${2:-}"
  if [[ -z "$secret_value" ]]; then
    return 0
  fi
  if gcloud secrets describe "$secret_name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo -n "$secret_value" | gcloud secrets versions add "$secret_name" \
      --project="$GCP_PROJECT_ID" --data-file=- >/dev/null
  else
    echo -n "$secret_value" | gcloud secrets create "$secret_name" \
      --project="$GCP_PROJECT_ID" --replication-policy=automatic --data-file=- >/dev/null
  fi
}

pf_scheduler_upsert_http() {
  # Args via env: JOB_NAME SCHEDULE TZ ENDPOINT HEADERS [ATTEMPT_DEADLINE]
  # stdout/stderr of gcloud are discarded: job YAML embeds secret headers.
  local location="${GCP_SCHEDULER_LOCATION:-europe-west1}"
  local deadline="${ATTEMPT_DEADLINE:-120s}"
  if gcloud scheduler jobs describe "$JOB_NAME" --location="$location" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud scheduler jobs update http "$JOB_NAME" \
      --project="$GCP_PROJECT_ID" \
      --location="$location" \
      --schedule="$SCHEDULE" \
      --time-zone="$TZ" \
      --uri="$ENDPOINT" \
      --http-method=POST \
      --headers="$HEADERS" \
      --message-body='{}' \
      --attempt-deadline="$deadline" \
      >/dev/null
    echo "Updated scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ}) → ${ENDPOINT}"
  else
    gcloud scheduler jobs create http "$JOB_NAME" \
      --project="$GCP_PROJECT_ID" \
      --location="$location" \
      --schedule="$SCHEDULE" \
      --time-zone="$TZ" \
      --uri="$ENDPOINT" \
      --http-method=POST \
      --headers="$HEADERS" \
      --message-body='{}' \
      --attempt-deadline="$deadline" \
      >/dev/null
    echo "Created scheduler job ${JOB_NAME} (${SCHEDULE} ${TZ}) → ${ENDPOINT}"
  fi
}
