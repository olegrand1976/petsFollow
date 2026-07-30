-- Suggested AMM (autorisation de mise sur le marché) for DAF autofill.
ALTER TABLE pharmacy.ref_medications
    ADD COLUMN IF NOT EXISTS amm_number TEXT NOT NULL DEFAULT '';
