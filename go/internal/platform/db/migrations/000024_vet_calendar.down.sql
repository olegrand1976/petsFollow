ALTER TABLE notifications.notification_preferences
    DROP COLUMN IF EXISTS email_on_visit_request;

ALTER TABLE visits.visits
    DROP COLUMN IF EXISTS pending_action_by,
    DROP COLUMN IF EXISTS proposed_scheduled_at,
    DROP COLUMN IF EXISTS duration_minutes;

ALTER TABLE visits.visits
    DROP CONSTRAINT IF EXISTS visits_status_check;

ALTER TABLE visits.visits
    ADD CONSTRAINT visits_status_check
    CHECK (status IN ('requested', 'confirmed', 'done', 'cancelled'));

DROP TABLE IF EXISTS practice.vet_vacations;
DROP TABLE IF EXISTS practice.vet_schedule_slots;
DROP TABLE IF EXISTS practice.vet_schedule;
