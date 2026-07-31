-- Lien traçabilité consultation (visite) → DAF.
ALTER TABLE pharmacy.daf_documents
    ADD COLUMN IF NOT EXISTS visit_id UUID REFERENCES visits.visits(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_pharmacy_daf_visit
    ON pharmacy.daf_documents (visit_id)
    WHERE visit_id IS NOT NULL;
