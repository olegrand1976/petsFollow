-- Pharmacie BE — stocks multi-dépôts, lots FEFO, mouvements, settings péremption.
CREATE TABLE IF NOT EXISTS pharmacy.practice_settings (
    practice_id UUID PRIMARY KEY REFERENCES practice.practices(id) ON DELETE CASCADE,
    warn_soon_days INT NOT NULL DEFAULT 90 CHECK (warn_soon_days > 0),
    warn_return_days INT NOT NULL DEFAULT 60 CHECK (warn_return_days > 0),
    warn_critical_days INT NOT NULL DEFAULT 30 CHECK (warn_critical_days > 0),
    receipt_warn_days INT NOT NULL DEFAULT 30 CHECK (receipt_warn_days >= 0),
    allow_expired_receipt BOOLEAN NOT NULL DEFAULT false,
    block_expired_on_daf BOOLEAN NOT NULL DEFAULT true,
    block_expired_on_adjust_out BOOLEAN NOT NULL DEFAULT true,
    auto_quarantine_expired BOOLEAN NOT NULL DEFAULT true,
    expiry_digest_enabled BOOLEAN NOT NULL DEFAULT true,
    expiry_digest_weekday INT NOT NULL DEFAULT 1 CHECK (expiry_digest_weekday BETWEEN 1 AND 7),
    notify_on_auto_quarantine BOOLEAN NOT NULL DEFAULT true,
    digest_user_ids UUID[] NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pharmacy_settings_bands CHECK (
        warn_critical_days <= warn_return_days AND warn_return_days <= warn_soon_days
    )
);

CREATE TABLE IF NOT EXISTS pharmacy.medication_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id, code)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pharmacy_deposit_one_default
    ON pharmacy.medication_deposits (practice_id)
    WHERE is_default;

CREATE TABLE IF NOT EXISTS pharmacy.medication_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    deposit_id UUID NOT NULL REFERENCES pharmacy.medication_deposits(id) ON DELETE RESTRICT,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE RESTRICT,
    lot_number TEXT NOT NULL,
    expires_on DATE NOT NULL,
    qty_on_hand NUMERIC(12,3) NOT NULL CHECK (qty_on_hand >= 0),
    unit TEXT NOT NULL DEFAULT 'unit',
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'quarantine', 'wasted')),
    quarantined_at TIMESTAMPTZ,
    quarantine_reason TEXT,
    wasted_at TIMESTAMPTZ,
    waste_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id, deposit_id, medication_id, lot_number, expires_on)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_batches_fefo
    ON pharmacy.medication_batches (practice_id, medication_id, deposit_id, expires_on ASC)
    WHERE qty_on_hand > 0 AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_pharmacy_batches_expiry
    ON pharmacy.medication_batches (practice_id, expires_on ASC)
    WHERE qty_on_hand > 0 AND status IN ('active', 'quarantine');

CREATE TABLE IF NOT EXISTS pharmacy.stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES pharmacy.medication_batches(id) ON DELETE RESTRICT,
    delta NUMERIC(12,3) NOT NULL,
    reason TEXT NOT NULL
        CHECK (reason IN ('receipt', 'daf', 'adjust', 'waste', 'daf_cancel', 'quarantine', 'unquarantine')),
    reason_detail TEXT,
    daf_item_id UUID,
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_movements_practice_created
    ON pharmacy.stock_movements (practice_id, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.practice_settings TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.medication_deposits TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.medication_batches TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.stock_movements TO petsfollow_app;
  END IF;
END $$;
