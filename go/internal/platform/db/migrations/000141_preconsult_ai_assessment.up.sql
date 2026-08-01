ALTER TABLE visits.preconsult_intakes
    ADD COLUMN IF NOT EXISTS ai_urgency TEXT
        CHECK (ai_urgency IS NULL OR ai_urgency IN ('green', 'orange', 'red')),
    ADD COLUMN IF NOT EXISTS ai_summary TEXT,
    ADD COLUMN IF NOT EXISTS ai_assessed_at TIMESTAMPTZ;
