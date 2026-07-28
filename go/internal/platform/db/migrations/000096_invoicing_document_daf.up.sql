-- Traçabilité facture Billit ← DAF pharmacie.
ALTER TABLE invoicing.documents
    ADD COLUMN IF NOT EXISTS daf_id UUID REFERENCES pharmacy.daf_documents(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invoicing_documents_daf
    ON invoicing.documents (daf_id)
    WHERE daf_id IS NOT NULL;
