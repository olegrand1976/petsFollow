#!/usr/bin/env bash
# Après migrate : grants DML sur tous les schémas petsFollow pour petsfollow_app.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/gcp-env.sh"

SCHEMAS=(public identity practice pets heartrate messaging notifications billing)
TMP="$(mktemp)"
{
  echo "GRANT CONNECT ON DATABASE ${DB_NAME} TO ${DB_USER};"
  for s in "${SCHEMAS[@]}"; do
    cat <<EOSQL
GRANT USAGE ON SCHEMA ${s} TO ${DB_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ${s} TO ${DB_USER};
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ${s} TO ${DB_USER};
EOSQL
  done
} >"$TMP"

GCS_URI="gs://premedica-prod-2025-db-exports/admin/pg-petsfollow-app-grants-$(date +%Y%m%d%H%M%S).sql"
gcloud storage cp "$TMP" "$GCS_URI" --quiet
gcloud sql import sql "$SQL_INSTANCE" "$GCS_URI" \
  --database="$DB_NAME" --project="$GCP_PROJECT_ID" --quiet
rm -f "$TMP"
echo "→ grants app OK"
