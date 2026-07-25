CREATE SCHEMA IF NOT EXISTS platform;

ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS request_preconsult BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS visits.preconsult_tokens (
    id UUID PRIMARY KEY,
    visit_id UUID NOT NULL REFERENCES visits.visits(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_preconsult_tokens_visit ON visits.preconsult_tokens(visit_id);
CREATE INDEX IF NOT EXISTS idx_preconsult_tokens_token ON visits.preconsult_tokens(token);

CREATE TABLE IF NOT EXISTS platform.brand_assets (
    key TEXT PRIMARY KEY,
    object_key TEXT NOT NULL DEFAULT '',
    public_url TEXT NOT NULL DEFAULT '',
    store_url TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON visits.preconsult_tokens TO petsfollow_app;
    GRANT USAGE ON SCHEMA platform TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON platform.brand_assets TO petsfollow_app;
  END IF;
END $$;
