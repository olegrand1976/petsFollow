-- Proforma client validation: magic-link token + accepted status; timeline event meta for CTA.

ALTER TABLE invoicing.documents DROP CONSTRAINT IF EXISTS documents_status_check;
ALTER TABLE invoicing.documents
    ADD CONSTRAINT documents_status_check CHECK (status IN (
        'draft', 'issued', 'sending', 'delivered', 'rejected', 'cancelled', 'accepted'
    ));

ALTER TABLE invoicing.documents
    ADD COLUMN IF NOT EXISTS public_token TEXT,
    ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoicing_documents_public_token
    ON invoicing.documents (public_token)
    WHERE public_token IS NOT NULL;

ALTER TABLE pets.dossier_events
    ADD COLUMN IF NOT EXISTS meta JSONB NOT NULL DEFAULT '{}'::jsonb;
