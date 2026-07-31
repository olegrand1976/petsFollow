DROP TABLE IF EXISTS pharmacy.medication_practice_attrs;

DROP INDEX IF EXISTS pharmacy.uq_pharmacy_inv_open_practice;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_inv_open_practice
    ON pharmacy.inventory_sessions (practice_id)
    WHERE status = 'open' AND deposit_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_inv_open_deposit
    ON pharmacy.inventory_sessions (practice_id, deposit_id)
    WHERE status = 'open' AND deposit_id IS NOT NULL;
