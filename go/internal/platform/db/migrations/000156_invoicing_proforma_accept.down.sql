ALTER TABLE pets.dossier_events DROP COLUMN IF EXISTS meta;

DROP INDEX IF EXISTS invoicing.idx_invoicing_documents_public_token;

ALTER TABLE invoicing.documents
    DROP COLUMN IF EXISTS accepted_at,
    DROP COLUMN IF EXISTS token_expires_at,
    DROP COLUMN IF EXISTS public_token;

ALTER TABLE invoicing.documents DROP CONSTRAINT IF EXISTS documents_status_check;
ALTER TABLE invoicing.documents
    ADD CONSTRAINT documents_status_check CHECK (status IN (
        'draft', 'issued', 'sending', 'delivered', 'rejected', 'cancelled'
    ));
