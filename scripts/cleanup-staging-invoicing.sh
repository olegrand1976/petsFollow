#!/usr/bin/env bash
# Nettoie les documents de test Billit (practice) sur une DB seedable.
# Usage :
#   PF_CLEANUP_TARGET=staging DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh
#   DATABASE_URL=…premedica-db-staging… ./scripts/cleanup-staging-invoicing.sh
#   DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh --connections
#   DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh --dry-run
#
# Cible : cabinets liés à des users *@petsfollow.test
# Ne touche pas aux docs source=saas_master.
# Localhost : exiger PF_CLEANUP_TARGET=staging (éviter footgun proxy→prod).
set -euo pipefail

DRY_RUN=false
RESET_CONNECTIONS=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --connections) RESET_CONNECTIONS=true ;;
    -h|--help)
      sed -n '2,14p' "$0"
      exit 0
      ;;
  esac
done

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib/pf-staging-cleanup-guard.sh
source "$ROOT/scripts/lib/pf-staging-cleanup-guard.sh"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL required" >&2
  exit 1
fi

pf_assert_staging_cleanup_url "$DATABASE_URL" "PF_ALLOW_PROD_INVOICING_CLEANUP"

run_sql() {
  local sql="$1"
  if [[ "$DRY_RUN" == true ]]; then
    echo "DRY-RUN SQL:"
    echo "$sql"
    return 0
  fi
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "$sql"
}

echo "== invoicing cleanup (seed *@petsfollow.test) dry_run=$DRY_RUN connections=$RESET_CONNECTIONS target=${PF_CLEANUP_TARGET:-unset} =="

COUNT_SQL=$(cat <<'SQL'
SELECT COUNT(*) AS practice_docs_to_purge
FROM invoicing.documents d
WHERE COALESCE(d.source, 'practice') = 'practice'
  AND d.status IN ('draft', 'rejected', 'sending')
  AND d.practice_id IN (
    SELECT DISTINCT COALESCE(u.practice_id, p.id)
    FROM identity.users u
    LEFT JOIN practice.practices p ON p.reference_vet_user_id = u.id
    WHERE u.email LIKE '%@petsfollow.test'
      AND COALESCE(u.practice_id, p.id) IS NOT NULL
  );
SQL
)
echo "→ Preview purge targets:"
psql "$DATABASE_URL" -c "$COUNT_SQL" || true

PURGE_SQL=$(cat <<'SQL'
WITH seed_practices AS (
  SELECT DISTINCT COALESCE(u.practice_id, p.id) AS id
  FROM identity.users u
  LEFT JOIN practice.practices p ON p.reference_vet_user_id = u.id
  WHERE u.email LIKE '%@petsfollow.test'
    AND COALESCE(u.practice_id, p.id) IS NOT NULL
),
del_lines AS (
  DELETE FROM invoicing.document_lines l
  USING invoicing.documents d
  WHERE l.document_id = d.id
    AND d.practice_id IN (SELECT id FROM seed_practices)
    AND COALESCE(d.source, 'practice') = 'practice'
    AND d.status IN ('draft', 'rejected', 'sending')
  RETURNING l.document_id
),
del_docs AS (
  DELETE FROM invoicing.documents d
  WHERE d.practice_id IN (SELECT id FROM seed_practices)
    AND COALESCE(d.source, 'practice') = 'practice'
    AND d.status IN ('draft', 'rejected', 'sending')
  RETURNING d.id
)
SELECT (SELECT COUNT(*) FROM del_docs) AS deleted_documents;
SQL
)

run_sql "$PURGE_SQL"

if [[ "$RESET_CONNECTIONS" == true ]]; then
  CONN_SQL=$(cat <<'SQL'
WITH seed_practices AS (
  SELECT DISTINCT COALESCE(u.practice_id, p.id) AS id
  FROM identity.users u
  LEFT JOIN practice.practices p ON p.reference_vet_user_id = u.id
  WHERE u.email LIKE '%@petsfollow.test'
    AND COALESCE(u.practice_id, p.id) IS NOT NULL
)
DELETE FROM invoicing.practice_connections
WHERE practice_id IN (SELECT id FROM seed_practices);
SQL
)
  echo "→ Reset practice_connections (force re-complete Billit):"
  run_sql "$CONN_SQL"
fi

echo "OK cleanup done."
echo "Si secrets mismatch : reconnecter via /invoicing (complete PartyID + ApiKey sandbox)."
