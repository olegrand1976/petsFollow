DROP INDEX IF EXISTS prescriptions.idx_prescriptions_visit;
ALTER TABLE prescriptions.prescriptions
    DROP COLUMN IF EXISTS care_advice,
    DROP COLUMN IF EXISTS visit_id;
