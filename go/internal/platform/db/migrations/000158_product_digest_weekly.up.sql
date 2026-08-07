-- Weekly product digest send idempotence (Saturday morning rollup).

CREATE TABLE IF NOT EXISTS ops.product_digest_weekly_sends (
    week_start DATE NOT NULL,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (week_start, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ops_product_digest_weekly_sends_sent_at
    ON ops.product_digest_weekly_sends (sent_at);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.product_digest_weekly_sends TO petsfollow_app;
  END IF;
END $$;
