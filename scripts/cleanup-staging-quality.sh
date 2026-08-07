#!/usr/bin/env bash
# Purge les artefacts smoke / e2e quality sur une DB seedable (staging / local).
# Usage :
#   PF_CLEANUP_TARGET=staging DATABASE_URL=… ./scripts/cleanup-staging-quality.sh
#   DATABASE_URL=…premedica-db-staging… ./scripts/cleanup-staging-quality.sh
#   … ./scripts/cleanup-staging-quality.sh --dry-run
#   … ./scripts/cleanup-staging-quality.sh --skip-invoicing
#
# Cible : tous les marqueurs smoke + e2e Playwright (messages, mesures, visites/CR,
# salles/sites, prospects, tickets support, pharmacie E2E-*/S6-*) et les users
# éphémères smoke*/e2e (uniqueE2EEmail + hr-dur-*) @petsfollow.test.
# Ne touche PAS au graphe seed (vet.demo, client.demo, …) ni à saas_master,
# ni aux factures delivered/issued (Billit live).
# Garde-fou : nuxtjs/tests/unit/e2e-cleanup-coverage.spec.ts verrouille la
# couverture des préfixes email e2e — le mettre à jour avec tout nouveau motif.
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
       OR body LIKE 'smoke vet to client %'
       OR body LIKE 'e2e media %') AS smoke_e2e_messages,
  (SELECT COUNT(*) FROM pets.blood_pressure_readings
    WHERE comment LIKE 'smoke bp %'
       OR comment LIKE 'e2e bp %') AS smoke_e2e_bp,
  (SELECT COUNT(*) FROM heartrate.sessions
    WHERE comment = 'smoke hr comment'
       OR comment LIKE 'e2e comment %') AS smoke_e2e_hr,
  (SELECT COUNT(*) FROM pets.weight_readings
    WHERE comment LIKE 'e2e weight %') AS e2e_weights,
  (SELECT COUNT(*) FROM labs.panels
    WHERE lab_name LIKE 'Smoke Lab %'
       OR lab_name LIKE 'E2E Lab %') AS smoke_e2e_labs,
  (SELECT COUNT(*) FROM visits.visits v
    WHERE v.notes LIKE 'e2e %' OR v.notes LIKE 'E2E %'
       OR EXISTS (
         SELECT 1 FROM visits.visit_reports r
         WHERE r.visit_id = v.id
           AND (r.body_text LIKE 'E2E %' OR r.body_text LIKE 'e2e %'
                OR r.body_text LIKE 'Desk resume CR %'
                OR r.transcript_text LIKE 'E2E %' OR r.transcript_text LIKE 'e2e %')
       )) AS e2e_visits,
  (SELECT COUNT(*) FROM practice.rooms WHERE name LIKE 'Box e2e %') AS e2e_rooms,
  (SELECT COUNT(*) FROM practice.sites WHERE name LIKE 'E2E Antenne %') AS e2e_sites,
  (SELECT COUNT(*) FROM sales.prospects
    WHERE contact_email LIKE 'crm.%@petsfollow.test'
       OR contact_email LIKE 'mail.%@petsfollow.test'
       OR practice_name LIKE 'Prospect E2E %'
       OR practice_name LIKE 'Prospect CRM %'
       OR practice_name LIKE 'Prospect Mail %') AS e2e_prospects,
  (SELECT COUNT(*) FROM ops.support_tickets
    WHERE subject LIKE 'E2E support %') AS e2e_support_tickets,
  (SELECT COUNT(*) FROM pharmacy.medication_batches
    WHERE lot_number LIKE 'E2E-%' OR lot_number LIKE 'S6-%') AS e2e_pharmacy_batches,
  (SELECT COUNT(*) FROM identity.users
    WHERE email LIKE 'smoke+%@petsfollow.test'
       OR email LIKE 'smoke-client+%@petsfollow.test'
       OR email LIKE 'smoke-comm+%@petsfollow.test'
       OR email LIKE 'e2e+%@petsfollow.test'
       OR email LIKE 'register+%@petsfollow.test'
       OR email LIKE 'unverified-login+%@petsfollow.test'
       OR email LIKE 'mismatch+%@petsfollow.test'
       OR email LIKE 'pw-vet+%@petsfollow.test'
       OR email LIKE 'pw-commercial+%@petsfollow.test'
       OR email LIKE 'hr-dur-vet-%@petsfollow.test'
       OR email LIKE 'hr-dur-client-%@petsfollow.test') AS ephemeral_users,
  (SELECT COUNT(*) FROM pets.pets
    WHERE name IN ('SmokePet', 'HR Dur Pet')) AS smoke_e2e_pets;
SQL
)
echo "→ Preview purge targets:"
psql "$DATABASE_URL" -c "$PREVIEW_SQL" || true

PURGE_SQL=$(cat <<'SQL'
BEGIN;

-- Messages smoke (H1 + body fixe) + média e2e marqué, sur threads seed.
DELETE FROM messaging.messages
WHERE body = 'smoke test message'
   OR body LIKE 'smoke vet to client %'
   OR body LIKE 'e2e media %';

-- Tension / FC / poids / labos marqués smoke ou e2e (animaux seed conservés).
DELETE FROM pets.blood_pressure_readings
WHERE comment LIKE 'smoke bp %' OR comment LIKE 'e2e bp %';
DELETE FROM heartrate.sessions
WHERE comment = 'smoke hr comment' OR comment LIKE 'e2e comment %';
DELETE FROM pets.weight_readings WHERE comment LIKE 'e2e weight %';
DELETE FROM labs.panels
WHERE lab_name LIKE 'Smoke Lab %' OR lab_name LIKE 'E2E Lab %';

-- Visites + CR e2e : notes marquées (créées via API) ou CR marqué (walk-in UI).
-- Toutes les FK entrantes de visits.visits sont CASCADE ou SET NULL.
DELETE FROM visits.visits v
WHERE v.notes LIKE 'e2e %'
   OR v.notes LIKE 'E2E %'
   OR EXISTS (
     SELECT 1 FROM visits.visit_reports r
     WHERE r.visit_id = v.id
       AND (r.body_text LIKE 'E2E %' OR r.body_text LIKE 'e2e %'
            OR r.body_text LIKE 'Desk resume CR %'
            OR r.transcript_text LIKE 'E2E %' OR r.transcript_text LIKE 'e2e %')
   );

-- Salles e2e (Box e2e) : RDV rattachés puis salles.
DELETE FROM visits.visits v
USING practice.rooms r
WHERE v.room_id = r.id AND r.name LIKE 'Box e2e %';
DELETE FROM practice.rooms WHERE name LIKE 'Box e2e %';

-- Sites e2e (E2E Antenne) : RDV rattachés puis site (rooms/schedule CASCADE).
DELETE FROM visits.visits v
USING practice.sites s
WHERE v.site_id = s.id AND s.name LIKE 'E2E Antenne %';
DELETE FROM practice.sites WHERE name LIKE 'E2E Antenne %';

-- Prospects CRM créés par e2e commercial (email timestampé ou nom marqué).
-- Events / activités / emails prospect : FK CASCADE.
DELETE FROM sales.prospects
WHERE contact_email LIKE 'crm.%@petsfollow.test'
   OR contact_email LIKE 'mail.%@petsfollow.test'
   OR practice_name LIKE 'Prospect E2E %'
   OR practice_name LIKE 'Prospect CRM %'
   OR practice_name LIKE 'Prospect Mail %';

-- Tickets support e2e (replies CASCADE).
DELETE FROM ops.support_tickets WHERE subject LIKE 'E2E support %';

-- Pharmacie : lots e2e (@pharmacy) + smoke S6, DAF, BL, inventaire.
CREATE TEMP TABLE e2e_batches AS
SELECT id FROM pharmacy.medication_batches
WHERE lot_number LIKE 'E2E-%' OR lot_number LIKE 'S6-%';

CREATE TEMP TABLE e2e_dafs AS
SELECT DISTINCT d.id
FROM pharmacy.daf_documents d
JOIN pharmacy.daf_items i ON i.daf_id = d.id
WHERE i.batch_id IN (SELECT id FROM e2e_batches)
   OR i.amm_number LIKE 'BE-S6-%'
   OR i.amm_number LIKE 'BE-E2E%';

DELETE FROM pharmacy.stock_movements m
WHERE m.batch_id IN (SELECT id FROM e2e_batches)
   OR m.daf_item_id IN (
     SELECT i.id FROM pharmacy.daf_items i WHERE i.daf_id IN (SELECT id FROM e2e_dafs)
   );

-- Sessions touchées capturées avant la purge des lignes : après, le lien est perdu.
CREATE TEMP TABLE e2e_inventory_sessions AS
SELECT DISTINCT session_id AS id
FROM pharmacy.inventory_lines
WHERE batch_id IN (SELECT id FROM e2e_batches);

DELETE FROM pharmacy.inventory_lines WHERE batch_id IN (SELECT id FROM e2e_batches);

-- Sessions e2e fermées/annulées devenues vides (une session encore utilisée reste).
DELETE FROM pharmacy.inventory_sessions s
WHERE s.id IN (SELECT id FROM e2e_inventory_sessions)
  AND s.status IN ('closed', 'cancelled')
  AND NOT EXISTS (SELECT 1 FROM pharmacy.inventory_lines l WHERE l.session_id = s.id);

DELETE FROM pharmacy.daf_items WHERE daf_id IN (SELECT id FROM e2e_dafs);
DELETE FROM pharmacy.daf_documents WHERE id IN (SELECT id FROM e2e_dafs);

DELETE FROM pharmacy.delivery_note_items
WHERE batch_id IN (SELECT id FROM e2e_batches)
   OR lot_number LIKE 'E2E-%' OR lot_number LIKE 'S6-%';
DELETE FROM pharmacy.delivery_notes dn
WHERE dn.note_number LIKE 'E2E-BL-%'
  AND NOT EXISTS (SELECT 1 FROM pharmacy.delivery_note_items i WHERE i.delivery_note_id = dn.id);

DELETE FROM pharmacy.medication_batches WHERE id IN (SELECT id FROM e2e_batches);

DROP TABLE e2e_batches;
DROP TABLE e2e_dafs;
DROP TABLE e2e_inventory_sessions;

-- Pets smoke/e2e + ledgers qui référencent le pet sans CASCADE.
DELETE FROM billing.commission_ledger cl
USING pets.pets p
WHERE cl.pet_id = p.id AND p.name IN ('SmokePet', 'HR Dur Pet');

DELETE FROM pets.pets WHERE name IN ('SmokePet', 'HR Dur Pet');

-- Ensemble unique des users éphémères smoke* + e2e (uniqueE2EEmail + hr-dur).
CREATE TEMP TABLE ephemeral_users AS
SELECT id, practice_id
FROM identity.users
WHERE email LIKE 'smoke+%@petsfollow.test'
   OR email LIKE 'smoke-client+%@petsfollow.test'
   OR email LIKE 'smoke-comm+%@petsfollow.test'
   OR email LIKE 'e2e+%@petsfollow.test'
   OR email LIKE 'register+%@petsfollow.test'
   OR email LIKE 'unverified-login+%@petsfollow.test'
   OR email LIKE 'mismatch+%@petsfollow.test'
   OR email LIKE 'pw-vet+%@petsfollow.test'
   OR email LIKE 'pw-commercial+%@petsfollow.test'
   OR email LIKE 'hr-dur-vet-%@petsfollow.test'
   OR email LIKE 'hr-dur-client-%@petsfollow.test';

-- Détacher références entrantes.
UPDATE identity.users u
SET assigned_commercial_id = NULL
FROM ephemeral_users s
WHERE u.assigned_commercial_id = s.id;

UPDATE identity.users u
SET sponsor_user_id = NULL
FROM ephemeral_users s
WHERE u.sponsor_user_id = s.id;

UPDATE identity.users u
SET manager_user_id = NULL
FROM ephemeral_users s
WHERE u.manager_user_id = s.id;

UPDATE sales.prospects p
SET referring_vet_user_id = NULL
FROM ephemeral_users s
WHERE p.referring_vet_user_id = s.id;

UPDATE sales.prospects p
SET converted_vet_user_id = NULL
FROM ephemeral_users s
WHERE p.converted_vet_user_id = s.id;

UPDATE practice.practices p
SET reference_vet_user_id = NULL
FROM ephemeral_users s
WHERE p.reference_vet_user_id = s.id;

UPDATE invoicing.documents d
SET created_by = NULL
FROM ephemeral_users s
WHERE d.created_by = s.id;

-- Artefacts liés aux smoke users (FK sans CASCADE / RESTRICT).
DELETE FROM identity.email_verification_tokens t USING ephemeral_users s WHERE t.user_id = s.id;
DELETE FROM identity.password_reset_tokens t USING ephemeral_users s WHERE t.user_id = s.id;
DELETE FROM notifications.device_tokens t USING ephemeral_users s WHERE t.user_id = s.id;
DELETE FROM notifications.notification_preferences np USING ephemeral_users s WHERE np.vet_user_id = s.id;
DELETE FROM notifications.notification_log nl USING ephemeral_users s WHERE nl.vet_user_id = s.id;
DELETE FROM messaging.vet_availability va USING ephemeral_users s WHERE va.vet_user_id = s.id;

DELETE FROM invoicing.connect_states c USING ephemeral_users s WHERE c.created_by = s.id;

DELETE FROM billing.commercial_bonus_awards a USING ephemeral_users s WHERE a.commercial_user_id = s.id;
DELETE FROM billing.commercial_payout_lines l USING ephemeral_users s WHERE l.commercial_user_id = s.id;
DELETE FROM billing.payout_lines l USING ephemeral_users s WHERE l.vet_user_id = s.id;
DELETE FROM billing.commercial_commission_ledger l
USING ephemeral_users s
WHERE l.commercial_user_id = s.id OR l.vet_user_id = s.id OR l.client_user_id = s.id;
DELETE FROM billing.commission_ledger l
USING ephemeral_users s
WHERE l.vet_user_id = s.id OR l.client_user_id = s.id;

DELETE FROM billing.pet_entitlements e USING ephemeral_users s WHERE e.owner_user_id = s.id;
DELETE FROM billing.addon_entitlements e USING ephemeral_users s WHERE e.owner_user_id = s.id;
DELETE FROM billing.stripe_customers c USING ephemeral_users s WHERE c.user_id = s.id;

DELETE FROM labs.panels p USING ephemeral_users s WHERE p.author_user_id = s.id;
DELETE FROM pets.blood_pressure_readings r
USING ephemeral_users s
WHERE r.owner_user_id = s.id OR r.author_user_id = s.id;
DELETE FROM pets.weight_readings r
USING ephemeral_users s
WHERE r.owner_user_id = s.id OR r.author_user_id = s.id;
DELETE FROM heartrate.sessions h
USING ephemeral_users s
WHERE h.owner_user_id = s.id;

DELETE FROM visits.visit_reports vr USING ephemeral_users s WHERE vr.author_user_id = s.id;
DELETE FROM pets.dossier_events de USING ephemeral_users s WHERE de.author_user_id = s.id;
DELETE FROM pets.documents d USING ephemeral_users s WHERE d.uploaded_by_user_id = s.id;
DELETE FROM imaging.pet_study_comments c USING ephemeral_users s WHERE c.author_user_id = s.id;
DELETE FROM practice.client_access ca
USING ephemeral_users s
WHERE ca.granted_by_user_id = s.id;
DELETE FROM pets.pet_access pa
USING ephemeral_users s
WHERE pa.granted_by_user_id = s.id;

DELETE FROM pets.pets p USING ephemeral_users s WHERE p.owner_user_id = s.id;
DELETE FROM messaging.messages m USING ephemeral_users s WHERE m.sender_user_id = s.id;
DELETE FROM messaging.threads t
USING ephemeral_users s
WHERE t.client_user_id = s.id OR t.vet_user_id = s.id;
DELETE FROM practice.client_vet_link_requests lr
USING ephemeral_users s
WHERE lr.vet_user_id = s.id OR lr.client_user_id = s.id;
DELETE FROM practice.practice_clients pc
USING ephemeral_users s
WHERE pc.client_user_id = s.id OR pc.vet_user_id = s.id;
DELETE FROM practice.team_members tm USING ephemeral_users s WHERE tm.user_id = s.id;
DELETE FROM practice.invitations i USING ephemeral_users s WHERE i.vet_user_id = s.id;
DELETE FROM sales.prospects p USING ephemeral_users s WHERE p.commercial_user_id = s.id;

UPDATE identity.users u
SET active_profile_id = NULL, practice_id = NULL, branch_id = NULL, sponsor_user_id = NULL
FROM ephemeral_users s
WHERE u.id = s.id;

DELETE FROM identity.profiles p USING ephemeral_users s WHERE p.user_id = s.id;
DELETE FROM identity.users u USING ephemeral_users s WHERE u.id = s.id;

-- Cabinets créés uniquement par register smoke/e2e (plus aucun user/profil rattaché).
DELETE FROM practice.practices p
WHERE p.name IN ('Smoke Practice', 'Cabinet E2E', 'Cabinet Unverified', 'Cabinet HR Dur')
  AND NOT EXISTS (SELECT 1 FROM identity.users u WHERE u.practice_id = p.id)
  AND NOT EXISTS (SELECT 1 FROM identity.profiles pr WHERE pr.practice_id = p.id);

DROP TABLE ephemeral_users;

COMMIT;

SELECT 'quality_cleanup_ok' AS status;
SQL
)

run_sql "$PURGE_SQL"

echo "OK cleanup-staging-quality done."
