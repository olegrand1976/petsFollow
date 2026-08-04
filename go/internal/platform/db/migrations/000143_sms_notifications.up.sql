-- SMS transactionnel (Telnyx) : pref canal opt-out + journal d'envoi.
ALTER TABLE notifications.client_preferences
    ADD COLUMN IF NOT EXISTS sms BOOLEAN NOT NULL DEFAULT TRUE;

-- Journal des SMS client. Le corps n'est jamais stocké (kind + locale suffisent
-- à retrouver le template) ; to_phone est la seule PII, effacée en cascade avec
-- le compte et purgée par le job de rétention.
CREATE TABLE IF NOT EXISTS notifications.sms_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    visit_id UUID NULL REFERENCES visits.visits(id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK (kind IN ('visit_confirmed','visit_reminder','visit_reschedule')),
    to_phone TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT 'fr',
    scheduled_for TIMESTAMPTZ NULL,
    provider_message_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('sending','sent','dry_run','skipped','error')),
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotence rappel J-1 : au plus un rappel par (visite, créneau) ; un
-- reschedule vers un nouveau créneau ré-autorise un rappel.
CREATE UNIQUE INDEX IF NOT EXISTS sms_log_reminder_once
    ON notifications.sms_log (visit_id, scheduled_for)
    WHERE kind = 'visit_reminder';

CREATE INDEX IF NOT EXISTS sms_log_user_created_idx
    ON notifications.sms_log (user_id, created_at);
