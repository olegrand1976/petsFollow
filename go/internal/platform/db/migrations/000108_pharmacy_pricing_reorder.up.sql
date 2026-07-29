-- Phase 2.A / 2.B — catalogue prix practice + seuils réassort
CREATE TABLE IF NOT EXISTS pharmacy.medication_prices (
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE CASCADE,
    purchase_price_cents INT NOT NULL DEFAULT 0 CHECK (purchase_price_cents >= 0),
    sell_price_cents INT NOT NULL DEFAULT 0 CHECK (sell_price_cents >= 0),
    vat_percent NUMERIC(5,2) NOT NULL DEFAULT 21.00 CHECK (vat_percent >= 0 AND vat_percent <= 100),
    currency TEXT NOT NULL DEFAULT 'EUR',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    PRIMARY KEY (practice_id, medication_id)
);

CREATE TABLE IF NOT EXISTS pharmacy.reorder_thresholds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE CASCADE,
    deposit_id UUID REFERENCES pharmacy.medication_deposits(id) ON DELETE CASCADE,
    min_qty NUMERIC(12,3) NOT NULL DEFAULT 1 CHECK (min_qty >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pharmacy_reorder_deposit
    ON pharmacy.reorder_thresholds (practice_id, medication_id, deposit_id)
    WHERE deposit_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_pharmacy_reorder_practice_wide
    ON pharmacy.reorder_thresholds (practice_id, medication_id)
    WHERE deposit_id IS NULL;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.medication_prices TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.reorder_thresholds TO petsfollow_app;
  END IF;
END $$;
