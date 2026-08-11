-- Support: to_test/done statuses, status history, private attachments

ALTER TABLE ops.support_tickets
    DROP CONSTRAINT IF EXISTS support_tickets_status_check;

UPDATE ops.support_tickets
SET status = 'done'
WHERE status = 'resolved';

ALTER TABLE ops.support_tickets
    ADD CONSTRAINT support_tickets_status_check
    CHECK (status IN ('open', 'in_progress', 'to_test', 'done', 'closed'));

CREATE TABLE IF NOT EXISTS ops.support_ticket_status_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES ops.support_tickets(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL
        CHECK (to_status IN ('open', 'in_progress', 'to_test', 'done', 'closed')),
    changed_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_support_status_events_ticket
    ON ops.support_ticket_status_events (ticket_id, created_at ASC);

CREATE TABLE IF NOT EXISTS ops.support_ticket_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES ops.support_tickets(id) ON DELETE CASCADE,
    uploaded_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    file_name TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_support_attachments_ticket
    ON ops.support_ticket_attachments (ticket_id, created_at ASC);

-- Backfill: open at creation; if already moved, add a second hop to current status.
INSERT INTO ops.support_ticket_status_events (ticket_id, from_status, to_status, changed_by, created_at)
SELECT t.id, NULL, 'open', t.created_by, t.created_at
FROM ops.support_tickets t
WHERE NOT EXISTS (
    SELECT 1 FROM ops.support_ticket_status_events e WHERE e.ticket_id = t.id
);

INSERT INTO ops.support_ticket_status_events (ticket_id, from_status, to_status, changed_by, created_at)
SELECT t.id, 'open', t.status, NULL, t.updated_at
FROM ops.support_tickets t
WHERE t.status <> 'open'
  AND (
    SELECT COUNT(*) FROM ops.support_ticket_status_events e WHERE e.ticket_id = t.id
  ) = 1;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.support_ticket_status_events TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.support_ticket_attachments TO petsfollow_app;
  END IF;
END $$;
