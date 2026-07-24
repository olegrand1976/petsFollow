#!/usr/bin/env bash
# Après migrate : grants DML (exécuté en tant que petsfollow_migrate, owner des objets).
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/gcp-env.sh"

TMP="$(mktemp)"
cat >"$TMP" <<'EOSQL'
GRANT CONNECT ON DATABASE petsfollow TO petsfollow_app;
GRANT USAGE ON SCHEMA public TO petsfollow_app;
GRANT USAGE ON SCHEMA identity TO petsfollow_app;
GRANT USAGE ON SCHEMA practice TO petsfollow_app;
GRANT USAGE ON SCHEMA pets TO petsfollow_app;
GRANT USAGE ON SCHEMA heartrate TO petsfollow_app;
GRANT USAGE ON SCHEMA messaging TO petsfollow_app;
GRANT USAGE ON SCHEMA notifications TO petsfollow_app;
GRANT USAGE ON SCHEMA billing TO petsfollow_app;
GRANT USAGE ON SCHEMA care TO petsfollow_app;
GRANT USAGE ON SCHEMA visits TO petsfollow_app;
GRANT USAGE ON SCHEMA discovery TO petsfollow_app;
GRANT USAGE ON SCHEMA sales TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA identity TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA practice TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA pets TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA heartrate TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA messaging TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA notifications TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA billing TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA care TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA visits TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA discovery TO petsfollow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA sales TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA identity TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA practice TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA pets TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA heartrate TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA messaging TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA notifications TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA billing TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA care TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA visits TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA discovery TO petsfollow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA sales TO petsfollow_app;
-- schema_migrations (public) : lecture seule utile
GRANT SELECT ON TABLE public.schema_migrations TO petsfollow_app;
EOSQL

GCS_URI="gs://premedica-prod-2025-db-exports/admin/pg-petsfollow-app-grants-$(date +%Y%m%d%H%M%S).sql"
gcloud storage cp "$TMP" "$GCS_URI" --quiet
gcloud sql import sql "$SQL_INSTANCE" "$GCS_URI" \
  --database="$DB_NAME" --user="$MIGRATE_USER" --project="$GCP_PROJECT_ID" --quiet
rm -f "$TMP"
echo "→ grants app OK (via ${MIGRATE_USER})"
