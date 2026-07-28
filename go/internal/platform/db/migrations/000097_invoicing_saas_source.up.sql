-- Flux A: mark documents emitted under Billit master (SaaS LL-IT-SC → cabinet).
ALTER TABLE invoicing.documents
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'practice';

ALTER TABLE invoicing.documents
    DROP CONSTRAINT IF EXISTS invoicing_documents_source_check;

ALTER TABLE invoicing.documents
    ADD CONSTRAINT invoicing_documents_source_check
    CHECK (source IN ('practice', 'saas_master'));

CREATE INDEX IF NOT EXISTS idx_invoicing_documents_saas
    ON invoicing.documents (practice_id, created_at DESC)
    WHERE source = 'saas_master';
