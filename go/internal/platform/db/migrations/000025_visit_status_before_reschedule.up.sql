ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS status_before_reschedule TEXT
        CHECK (status_before_reschedule IS NULL OR status_before_reschedule IN ('requested', 'confirmed', 'done', 'cancelled', 'reschedule_pending'));
