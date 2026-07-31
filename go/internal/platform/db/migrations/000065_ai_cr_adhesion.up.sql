-- Idempotence des emails machine d'adhésion module CR IA (1 envoi / step / cabinet).

CREATE TABLE IF NOT EXISTS practice.ai_cr_email_sends (
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    step_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'sent'
        CHECK (status IN ('sent', 'skipped')),
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (practice_id, step_key)
);

CREATE INDEX IF NOT EXISTS idx_ai_cr_email_sends_sent_at
    ON practice.ai_cr_email_sends (step_key, sent_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.ai_cr_email_sends TO petsfollow_app;
  END IF;
END $$;
