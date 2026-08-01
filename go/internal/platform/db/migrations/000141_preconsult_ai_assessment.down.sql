ALTER TABLE visits.preconsult_intakes
    DROP COLUMN IF EXISTS ai_assessed_at,
    DROP COLUMN IF EXISTS ai_summary,
    DROP COLUMN IF EXISTS ai_urgency;
