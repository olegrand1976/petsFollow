-- Idempotent Billit webhook ingestion (same provider + order + event type).
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoicing_webhook_provider_external_event
    ON invoicing.webhook_events (provider, external_id, COALESCE(event_type, ''))
    WHERE external_id IS NOT NULL AND external_id <> '';
