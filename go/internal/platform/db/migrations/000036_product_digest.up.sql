-- Daily product digest emails (admin / commercial / commercial_manager)

CREATE SCHEMA IF NOT EXISTS ops;

CREATE TABLE IF NOT EXISTS ops.product_digests (
    digest_date DATE PRIMARY KEY,
    headline TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    headline_by_locale JSONB NOT NULL DEFAULT '{}'::jsonb,
    body_by_locale JSONB NOT NULL DEFAULT '{}'::jsonb,
    commits_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'ready', 'empty', 'sent')),
    generated_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS ops.product_digest_sends (
    digest_date DATE NOT NULL REFERENCES ops.product_digests(digest_date) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (digest_date, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ops_product_digest_sends_sent_at
    ON ops.product_digest_sends (sent_at);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA ops TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.product_digests TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.product_digest_sends TO petsfollow_app;
  END IF;
END $$;
