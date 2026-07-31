-- Phase 2.E — inventaire annuel (sessions + lignes snapshot)
CREATE TABLE IF NOT EXISTS pharmacy.inventory_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    deposit_id UUID REFERENCES pharmacy.medication_deposits(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closed', 'cancelled')),
    notes TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    closed_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_inv_sessions_practice
    ON pharmacy.inventory_sessions (practice_id, created_at DESC);

-- At most one open session per practice (+ optional deposit scope).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_inv_open_practice
    ON pharmacy.inventory_sessions (practice_id)
    WHERE status = 'open' AND deposit_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacy_inv_open_deposit
    ON pharmacy.inventory_sessions (practice_id, deposit_id)
    WHERE status = 'open' AND deposit_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS pharmacy.inventory_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES pharmacy.inventory_sessions(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES pharmacy.medication_batches(id) ON DELETE RESTRICT,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE RESTRICT,
    deposit_id UUID REFERENCES pharmacy.medication_deposits(id) ON DELETE SET NULL,
    lot_number TEXT NOT NULL,
    expires_on DATE NOT NULL,
    batch_status TEXT NOT NULL DEFAULT 'active',
    system_qty NUMERIC(12,3) NOT NULL,
    counted_qty NUMERIC(12,3),
    adjusted BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE (session_id, batch_id)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_inv_lines_session
    ON pharmacy.inventory_lines (session_id, sort_order);

ALTER TABLE pharmacy.stock_movements
    ADD COLUMN IF NOT EXISTS inventory_session_id UUID REFERENCES pharmacy.inventory_sessions(id) ON DELETE SET NULL;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.inventory_sessions TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.inventory_lines TO petsfollow_app;
  END IF;
END $$;
