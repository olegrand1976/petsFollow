-- Review fixes: unique open inventory per practice; practice-scoped medication attrs (withdrawal / food-chain).

-- Cancel duplicate open sessions (keep oldest per practice) before tightening the unique index.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY practice_id ORDER BY created_at ASC, id ASC) AS rn
    FROM pharmacy.inventory_sessions
    WHERE status = 'open'
)
UPDATE pharmacy.inventory_sessions s
SET status = 'cancelled', closed_at = now()
FROM ranked r
WHERE s.id = r.id AND r.rn > 1;

DROP INDEX IF EXISTS pharmacy.uq_pharmacy_inv_open_practice;
DROP INDEX IF EXISTS pharmacy.uq_pharmacy_inv_open_deposit;

-- At most one open inventory session per practice (any deposit scope).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_inv_open_practice
    ON pharmacy.inventory_sessions (practice_id)
    WHERE status = 'open';

CREATE TABLE IF NOT EXISTS pharmacy.medication_practice_attrs (
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE CASCADE,
    withdrawal_meat_days INT,
    withdrawal_milk_days INT,
    withdrawal_eggs_days INT,
    food_chain_banned BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (practice_id, medication_id)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_med_practice_attrs_med
    ON pharmacy.medication_practice_attrs (medication_id);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.medication_practice_attrs TO petsfollow_app;
  END IF;
END $$;
