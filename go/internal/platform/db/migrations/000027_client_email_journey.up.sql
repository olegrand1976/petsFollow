-- Client discovery / loyalty email journey (parallel to in-app Discovery)

CREATE TABLE IF NOT EXISTS discovery.email_journey (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    anchor_at TIMESTAMPTZ NOT NULL,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'paused', 'completed'))
);

CREATE INDEX IF NOT EXISTS idx_discovery_email_journey_active
    ON discovery.email_journey (status, anchor_at)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS discovery.email_sends (
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    step_key TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'sent'
        CHECK (status IN ('sent', 'skipped')),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (user_id, step_key)
);

CREATE INDEX IF NOT EXISTS idx_discovery_email_sends_sent_at
    ON discovery.email_sends (step_key, sent_at);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON discovery.email_journey TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON discovery.email_sends TO petsfollow_app;
  END IF;
END $$;
