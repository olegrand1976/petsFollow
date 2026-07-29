-- Phase 4.C — temps d'attente structurés (référentiel + snapshot DAF)
ALTER TABLE pharmacy.ref_medications
    ADD COLUMN IF NOT EXISTS withdrawal_meat_days INT,
    ADD COLUMN IF NOT EXISTS withdrawal_milk_days INT,
    ADD COLUMN IF NOT EXISTS withdrawal_eggs_days INT,
    ADD COLUMN IF NOT EXISTS food_chain_banned BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE pharmacy.daf_items
    ADD COLUMN IF NOT EXISTS withdrawal_meat_days INT,
    ADD COLUMN IF NOT EXISTS withdrawal_milk_days INT,
    ADD COLUMN IF NOT EXISTS withdrawal_eggs_days INT;
