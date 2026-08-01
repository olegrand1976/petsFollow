-- Research warehouse: opt-in practice + anonymized events / weekly aggregates.

ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS research_opt_in_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS research_opt_in_by UUID REFERENCES identity.users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_practices_research_opt_in
    ON practice.practices (research_opt_in_at)
    WHERE research_opt_in_at IS NOT NULL;

CREATE SCHEMA IF NOT EXISTS research;

CREATE TABLE IF NOT EXISTS research.anon_events (
    id UUID PRIMARY KEY,
    event_week DATE NOT NULL,
    postal_code TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    country_code TEXT NOT NULL DEFAULT 'BE',
    species TEXT NOT NULL DEFAULT 'unknown',
    age_band TEXT NOT NULL DEFAULT 'unknown'
        CHECK (age_band IN ('0-1', '1-7', '7+', 'unknown')),
    signal_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_hash TEXT NOT NULL,
    practice_id_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_hash)
);

CREATE INDEX IF NOT EXISTS idx_research_anon_week_signal
    ON research.anon_events (event_week, signal_type, country_code, postal_code, species);

CREATE INDEX IF NOT EXISTS idx_research_anon_practice_hash
    ON research.anon_events (practice_id_hash);

CREATE TABLE IF NOT EXISTS research.weekly_aggregates (
    event_week DATE NOT NULL,
    postal_code TEXT NOT NULL DEFAULT '',
    country_code TEXT NOT NULL DEFAULT 'BE',
    species TEXT NOT NULL DEFAULT 'unknown',
    signal_type TEXT NOT NULL,
    event_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_week, postal_code, country_code, species, signal_type)
);

CREATE INDEX IF NOT EXISTS idx_research_agg_signal_week
    ON research.weekly_aggregates (signal_type, event_week DESC);

CREATE TABLE IF NOT EXISTS research.etl_watermarks (
    source_key TEXT PRIMARY KEY,
    watermark TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01'::timestamptz,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA research TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON research.anon_events TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON research.weekly_aggregates TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON research.etl_watermarks TO petsfollow_app;
  END IF;
END $$;
