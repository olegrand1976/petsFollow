-- Support bug-report tickets (Nuxt Pro + Flutter) + admin inbox replies

CREATE SCHEMA IF NOT EXISTS ops;

CREATE TABLE IF NOT EXISTS ops.support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    source TEXT NOT NULL
        CHECK (source IN ('nuxt_pro', 'flutter_client', 'flutter_pro_light')),
    subject TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    diagnostics JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_agent TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT '',
    route TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_support_tickets_status_created
    ON ops.support_tickets (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_support_tickets_created_by
    ON ops.support_tickets (created_by);

CREATE INDEX IF NOT EXISTS idx_ops_support_tickets_created_at
    ON ops.support_tickets (created_at DESC);

CREATE TABLE IF NOT EXISTS ops.support_ticket_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES ops.support_tickets(id) ON DELETE CASCADE,
    author_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_support_ticket_replies_ticket
    ON ops.support_ticket_replies (ticket_id, created_at);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA ops TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.support_tickets TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.support_ticket_replies TO petsfollow_app;
  END IF;
END $$;
