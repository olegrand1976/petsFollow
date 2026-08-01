CREATE SCHEMA IF NOT EXISTS labs;

CREATE TABLE IF NOT EXISTS labs.panels (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    practice_id UUID REFERENCES practice.practices(id),
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lab_name TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    document_id UUID REFERENCES pets.documents(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lab_panels_pet_collected
    ON labs.panels (pet_id, collected_at DESC);

CREATE TABLE IF NOT EXISTS labs.panel_results (
    id UUID PRIMARY KEY,
    panel_id UUID NOT NULL REFERENCES labs.panels(id) ON DELETE CASCADE,
    analyte_code TEXT NOT NULL,
    value_num DOUBLE PRECISION,
    value_text TEXT,
    unit TEXT NOT NULL DEFAULT '',
    ref_low DOUBLE PRECISION,
    ref_high DOUBLE PRECISION,
    flag TEXT NOT NULL DEFAULT 'unknown'
        CHECK (flag IN ('low', 'normal', 'high', 'unknown')),
    UNIQUE (panel_id, analyte_code)
);

CREATE INDEX IF NOT EXISTS idx_lab_panel_results_panel
    ON labs.panel_results (panel_id);

CREATE INDEX IF NOT EXISTS idx_lab_panel_results_analyte
    ON labs.panel_results (analyte_code);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA labs TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON labs.panels TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON labs.panel_results TO petsfollow_app;
  END IF;
END $$;
