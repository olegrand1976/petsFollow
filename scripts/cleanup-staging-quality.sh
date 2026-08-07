#!/usr/bin/env bash
# Purge les artefacts smoke / e2e quality sur une DB seedable (staging / local).
# Usage :
#   PF_CLEANUP_TARGET=staging DATABASE_URL=… ./scripts/cleanup-staging-quality.sh
#   DATABASE_URL=…premedica-db-staging… ./scripts/cleanup-staging-quality.sh
#   … ./scripts/cleanup-staging-quality.sh --dry-run
#   … ./scripts/cleanup-staging-quality.sh --skip-invoicing
#
# Cible : marqueurs smoke/e2e + users éphémères smoke*@petsfollow.test
# Ne touche PAS au graphe seed (vet.demo, client.demo, …) ni à saas_master.
# Localhost : exiger PF_CLEANUP_TARGET=staging (posé par infra/gcp/cleanup-staging-quality.sh).
set -euo pipefail

DRY_RUN=false
SKIP_INVOICING=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --skip-invoicing) SKIP_INVOICING=true ;;
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

pf_assert_staging_cleanup_url "$DATABASE_URL" "PF_ALLOW_PROD_QUALITY_CLEANUP"

run_sql() {
  local sql="$1"
  if [[ "$DRY_RUN" == true ]]; then
    echo "DRY-RUN SQL:"
    echo "$sql"
    echo "----"
    return 0
  fi
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "$sql"
}

echo "== staging quality cleanup dry_run=$DRY_RUN skip_invoicing=$SKIP_INVOICING target=${PF_CLEANUP_TARGET:-unset} =="

if [[ "$SKIP_INVOICING" != true ]]; then
  # Propager le marqueur / override vers le sous-script invoicing.
  export PF_CLEANUP_TARGET="${PF_CLEANUP_TARGET:-}"
  if [[ "${PF_ALLOW_PROD_QUALITY_CLEANUP:-}" == "1" ]]; then
    export PF_ALLOW_PROD_INVOICING_CLEANUP=1
  fi
  if [[ "$DRY_RUN" == true ]]; then
    bash "$ROOT/scripts/cleanup-staging-invoicing.sh" --dry-run
  else
    bash "$ROOT/scripts/cleanup-staging-invoicing.sh"
  fi
fi

# Preview counts (best-effort, read-only même en dry-run).
PREVIEW_SQL=$(cat <<'SQL'
SELECT
  (SELECT COUNT(*) FROM messaging.messages
    WHERE body = 'smoke test message'
       OR body LIKE 'smoke vet to client %') AS smoke_messages,
  (SELECT COUNT(*) FROM pets.blood_pressure_readings
    WHERE comment LIKE 'smoke bp %') AS smoke_bp,
  (SELECT COUNT(*) FROM heartrate.sessions
    WHERE comment = 'smoke hr comment') AS smoke_hr,
  (SELECT COUNT(*) FROM labs.panels
    WHERE lab_name LIKE 'Smoke Lab %') AS smoke_labs,
  (SELECT COUNT(*) FROM sales.prospects
    WHERE contact_email LIKE 'crm.%@petsfollow.test'
       OR contact_email LIKE 'mail.%@petsfollow.test') AS e2e_prospects,
  (SELECT COUNT(*) FROM identity.users
    WHERE email LIKE 'smoke+%@petsfollow.test'
       OR email LIKE 'smoke-client+%@petsfollow.test'
       OR email LIKE 'smoke-comm+%@petsfollow.test') AS smoke_users,
  (SELECT COUNT(*) FROM pets.pets
    WHERE name = 'SmokePet') AS smoke_pets;
SQL
)
echo "→ Preview purge targets:"
psql "$DATABASE_URL" -c "$PREVIEW_SQL" || true

PURGE_SQL=$(cat <<'SQL'
BEGIN;

-- Messages smoke (H1 + body fixe) sur threads seed.
DELETE FROM messaging.messages
WHERE body = 'smoke test message'
   OR body LIKE 'smoke vet to client %';

-- Tension / FC / labos marquées smoke (animaux seed conservés).
DELETE FROM pets.blood_pressure_readings WHERE comment LIKE 'smoke bp %';
DELETE FROM heartrate.sessions WHERE comment = 'smoke hr comment';
DELETE FROM labs.panels WHERE lab_name LIKE 'Smoke Lab %';

-- Prospects CRM créés par e2e commercial (email timestampé).
DELETE FROM sales.prospects
WHERE contact_email LIKE 'crm.%@petsfollow.test'
   OR contact_email LIKE 'mail.%@petsfollow.test';

-- Pets smoke + ledgers qui référencent le pet sans CASCADE.
DELETE FROM billing.commission_ledger cl
USING pets.pets p
WHERE cl.pet_id = p.id AND p.name = 'SmokePet';

DELETE FROM pets.pets WHERE name = 'SmokePet';

-- Ensemble unique des users éphémères smoke*.
CREATE TEMP TABLE smoke_users AS
SELECT id, practice_id
FROM identity.users
WHERE email LIKE 'smoke+%@petsfollow.test'
   OR email LIKE 'smoke-client+%@petsfollow.test'
   OR email LIKE 'smoke-comm+%@petsfollow.test';

-- Détacher références entrantes.
UPDATE identity.users u
SET assigned_commercial_id = NULL
FROM smoke_users s
WHERE u.assigned_commercial_id = s.id;

UPDATE identity.users u
SET sponsor_user_id = NULL
FROM smoke_users s
WHERE u.sponsor_user_id = s.id;

UPDATE sales.prospects p
SET referring_vet_user_id = NULL
FROM smoke_users s
WHERE p.referring_vet_user_id = s.id;

UPDATE practice.practices p
SET reference_vet_user_id = NULL
FROM smoke_users s
WHERE p.reference_vet_user_id = s.id;

UPDATE invoicing.documents d
SET created_by = NULL
FROM smoke_users s
WHERE d.created_by = s.id;

-- Artefacts liés aux smoke users (FK sans CASCADE / RESTRICT).
DELETE FROM identity.email_verification_tokens t USING smoke_users s WHERE t.user_id = s.id;
DELETE FROM identity.password_reset_tokens t USING smoke_users s WHERE t.user_id = s.id;
DELETE FROM notifications.device_tokens t USING smoke_users s WHERE t.user_id = s.id;
DELETE FROM notifications.notification_preferences np USING smoke_users s WHERE np.vet_user_id = s.id;
DELETE FROM notifications.notification_log nl USING smoke_users s WHERE nl.vet_user_id = s.id;
DELETE FROM messaging.vet_availability va USING smoke_users s WHERE va.vet_user_id = s.id;

DELETE FROM invoicing.connect_states c USING smoke_users s WHERE c.created_by = s.id;

DELETE FROM billing.commercial_bonus_awards a USING smoke_users s WHERE a.commercial_user_id = s.id;
DELETE FROM billing.commercial_payout_lines l USING smoke_users s WHERE l.commercial_user_id = s.id;
DELETE FROM billing.commercial_commission_ledger l
USING smoke_users s
WHERE l.commercial_user_id = s.id OR l.vet_user_id = s.id OR l.client_user_id = s.id;
DELETE FROM billing.commission_ledger l
USING smoke_users s
WHERE l.vet_user_id = s.id OR l.client_user_id = s.id;

DELETE FROM billing.pet_entitlements e USING smoke_users s WHERE e.owner_user_id = s.id;
DELETE FROM billing.addon_entitlements e USING smoke_users s WHERE e.owner_user_id = s.id;
DELETE FROM billing.stripe_customers c USING smoke_users s WHERE c.user_id = s.id;

DELETE FROM labs.panels p USING smoke_users s WHERE p.author_user_id = s.id;
DELETE FROM pets.blood_pressure_readings r
USING smoke_users s
WHERE r.owner_user_id = s.id OR r.author_user_id = s.id;
DELETE FROM heartrate.sessions h
USING smoke_users s
WHERE h.owner_user_id = s.id;

DELETE FROM pets.pets p USING smoke_users s WHERE p.owner_user_id = s.id;
DELETE FROM messaging.threads t
USING smoke_users s
WHERE t.client_user_id = s.id OR t.vet_user_id = s.id;
DELETE FROM practice.practice_clients pc
USING smoke_users s
WHERE pc.client_user_id = s.id OR pc.vet_user_id = s.id;
DELETE FROM practice.team_members tm USING smoke_users s WHERE tm.user_id = s.id;
DELETE FROM practice.invitations i USING smoke_users s WHERE i.vet_user_id = s.id;
DELETE FROM sales.prospects p USING smoke_users s WHERE p.commercial_user_id = s.id;

UPDATE identity.users u
SET active_profile_id = NULL, practice_id = NULL, branch_id = NULL, sponsor_user_id = NULL
FROM smoke_users s
WHERE u.id = s.id;

DELETE FROM identity.profiles p USING smoke_users s WHERE p.user_id = s.id;
DELETE FROM identity.users u USING smoke_users s WHERE u.id = s.id;

-- Cabinets créés uniquement par register smoke (plus aucun user/profil rattaché).
DELETE FROM practice.practices p
WHERE p.name = 'Smoke Practice'
  AND NOT EXISTS (SELECT 1 FROM identity.users u WHERE u.practice_id = p.id)
  AND NOT EXISTS (SELECT 1 FROM identity.profiles pr WHERE pr.practice_id = p.id);

DROP TABLE smoke_users;

COMMIT;

SELECT 'quality_cleanup_ok' AS status;
SQL
)

run_sql "$PURGE_SQL"

echo "OK cleanup-staging-quality done."
