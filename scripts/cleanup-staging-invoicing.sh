#!/usr/bin/env bash
# Nettoie les documents de test Billit (practice) sur une DB seedable.
# Usage :
#   DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh
#   DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh --connections
#   DATABASE_URL=… ./scripts/cleanup-staging-invoicing.sh --dry-run
#
# Cible : cabinets liés à des users *@petsfollow.test
# Ne touche pas aux docs source=saas_master.
set -euo pipefail

DRY_RUN=false
RESET_CONNECTIONS=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --connections) RESET_CONNECTIONS=true ;;
    -h|--help)
      sed -n '2,12p' "$0"
      exit 0
      ;;
  esac
done

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL required" >&2
  exit 1
fi

# Safety: allowlist staging Cloud SQL instance / local proxy only.
# Do NOT match GCP project ids like premedica-prod-2025 (substring -prod).
# Override (ops only): PF_ALLOW_PROD_INVOICING_CLEANUP=1
pf_invoicing_cleanup_url_allowed() {
  local u="$1"
  # Explicit prod instance → never allow without override (checked by caller).
  if [[ "$u" == *"petsfollow-db-prod"* ]]; then
    return 1
  fi
  # Staging Cloud SQL instance name (socket or DSN).
  if [[ "$u" == *"premedica-db-staging"* ]]; then
    return 0
  fi
  # cloud-sql-proxy / local rewrite (no instance name in URL).
  if [[ "$u" == *"@127.0.0.1:"* || "$u" == *"@localhost:"* ]]; then
    return 0
  fi
  return 1
}

if [[ "${PF_ALLOW_PROD_INVOICING_CLEANUP:-}" != "1" ]] && ! pf_invoicing_cleanup_url_allowed "$DATABASE_URL"; then
  echo "Refusing DATABASE_URL outside staging allowlist (premedica-db-staging | localhost/127.0.0.1)." >&2
  echo "Override only if intentional: PF_ALLOW_PROD_INVOICING_CLEANUP=1" >&2
  exit 1
fi

run_sql() {
  local sql="$1"
  if [[ "$DRY_RUN" == true ]]; then
    echo "DRY-RUN SQL:"
    echo "$sql"
    return 0
  fi
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "$sql"
}

echo "== invoicing cleanup (seed *@petsfollow.test) dry_run=$DRY_RUN connections=$RESET_CONNECTIONS =="

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
