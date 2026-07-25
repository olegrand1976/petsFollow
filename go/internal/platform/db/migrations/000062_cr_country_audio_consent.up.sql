ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS country_code CHAR(2) NOT NULL DEFAULT 'BE';

ALTER TABLE visits.visit_reports
    ADD COLUMN IF NOT EXISTS client_audio_consent_at TIMESTAMPTZ;
