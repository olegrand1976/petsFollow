DROP INDEX IF EXISTS ops.idx_ops_support_attachments_ticket;
DROP TABLE IF EXISTS ops.support_ticket_attachments;
DROP INDEX IF EXISTS ops.idx_ops_support_status_events_ticket;
DROP TABLE IF EXISTS ops.support_ticket_status_events;

ALTER TABLE ops.support_tickets
    DROP CONSTRAINT IF EXISTS support_tickets_status_check;

UPDATE ops.support_tickets
SET status = 'resolved'
WHERE status IN ('done', 'to_test');

ALTER TABLE ops.support_tickets
    ADD CONSTRAINT support_tickets_status_check
    CHECK (status IN ('open', 'in_progress', 'resolved', 'closed'));
