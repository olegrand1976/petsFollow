-- Unique protocol name per practice (case-insensitive).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_clinical_protocols_practice_name
    ON pharmacy.clinical_protocols (practice_id, lower(name));
