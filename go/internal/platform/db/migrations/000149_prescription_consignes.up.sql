-- Consignes: lien consultation + conseils/soins client (≠ ordonnance légale).
ALTER TABLE prescriptions.prescriptions
    ADD COLUMN IF NOT EXISTS visit_id UUID REFERENCES visits.visits(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS care_advice TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_prescriptions_visit
    ON prescriptions.prescriptions (visit_id)
    WHERE visit_id IS NOT NULL;
