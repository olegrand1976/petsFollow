DROP INDEX IF EXISTS invoicing.idx_invoicing_documents_saas;
ALTER TABLE invoicing.documents DROP CONSTRAINT IF EXISTS invoicing_documents_source_check;
ALTER TABLE invoicing.documents DROP COLUMN IF EXISTS source;
