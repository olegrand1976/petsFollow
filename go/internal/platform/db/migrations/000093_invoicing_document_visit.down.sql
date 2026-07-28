DROP INDEX IF EXISTS invoicing.idx_invoicing_documents_visit;
ALTER TABLE invoicing.documents DROP COLUMN IF EXISTS visit_id;
