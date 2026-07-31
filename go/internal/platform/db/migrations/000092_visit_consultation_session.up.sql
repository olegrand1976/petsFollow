-- Walk-in consultation sessions do not occupy client booking slots.
ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS consultation_session BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_visits_consultation_session
    ON visits.visits (practice_id, scheduled_at)
    WHERE consultation_session = false
      AND status IN ('requested', 'confirmed', 'reschedule_pending');
