DROP INDEX IF EXISTS invoicing.idx_invoicing_documents_daf;
ALTER TABLE invoicing.documents DROP COLUMN IF EXISTS daf_id;
