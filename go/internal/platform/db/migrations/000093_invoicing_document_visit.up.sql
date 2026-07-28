-- Traçabilité consultation → facture Billit.
ALTER TABLE invoicing.documents
    ADD COLUMN IF NOT EXISTS visit_id UUID REFERENCES visits.visits(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invoicing_documents_visit
    ON invoicing.documents (visit_id)
    WHERE visit_id IS NOT NULL;
