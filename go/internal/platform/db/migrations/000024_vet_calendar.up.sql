-- Vet calendar: schedule slots, vacations, visit reschedule fields, email pref.

CREATE TABLE IF NOT EXISTS practice.vet_schedule (
    practice_id UUID PRIMARY KEY REFERENCES practice.practices(id) ON DELETE CASCADE,
    client_booking_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    slot_duration_minutes INT NOT NULL DEFAULT 30
        CHECK (slot_duration_minutes IN (15, 30, 60)),
    vacations_declared_year INT NULL,
    timezone TEXT NOT NULL DEFAULT 'Europe/Brussels',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS practice.vet_schedule_slots (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    CHECK (start_time < end_time)
);

CREATE INDEX IF NOT EXISTS idx_vet_schedule_slots_practice
    ON practice.vet_schedule_slots (practice_id);

CREATE TABLE IF NOT EXISTS practice.vet_vacations (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    starts_on DATE NOT NULL,
    ends_on DATE NOT NULL,
    label TEXT,
    CHECK (starts_on <= ends_on)
);

CREATE INDEX IF NOT EXISTS idx_vet_vacations_practice
    ON practice.vet_vacations (practice_id, starts_on, ends_on);

ALTER TABLE visits.visits
    DROP CONSTRAINT IF EXISTS visits_status_check;

ALTER TABLE visits.visits
    ADD CONSTRAINT visits_status_check
    CHECK (status IN ('requested', 'confirmed', 'done', 'cancelled', 'reschedule_pending'));

ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS duration_minutes INT NULL
        CHECK (duration_minutes IS NULL OR duration_minutes IN (15, 30, 60)),
    ADD COLUMN IF NOT EXISTS proposed_scheduled_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS pending_action_by TEXT NULL
        CHECK (pending_action_by IS NULL OR pending_action_by IN ('vet', 'client'));

ALTER TABLE notifications.notification_preferences
    ADD COLUMN IF NOT EXISTS email_on_visit_request BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE visits.visits
SET pending_action_by = 'vet'
WHERE status = 'requested' AND pending_action_by IS NULL AND source = 'client';

UPDATE visits.visits
SET pending_action_by = 'client'
WHERE status = 'requested' AND pending_action_by IS NULL AND source = 'vet';

GRANT SELECT, INSERT, UPDATE, DELETE ON practice.vet_schedule TO CURRENT_USER;
GRANT SELECT, INSERT, UPDATE, DELETE ON practice.vet_schedule_slots TO CURRENT_USER;
GRANT SELECT, INSERT, UPDATE, DELETE ON practice.vet_vacations TO CURRENT_USER;
