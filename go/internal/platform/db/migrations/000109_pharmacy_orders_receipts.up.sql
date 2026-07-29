-- Phase 2.C / 2.D — commandes réassort + bons de livraison
CREATE TABLE IF NOT EXISTS pharmacy.suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_suppliers_practice
    ON pharmacy.suppliers (practice_id, name);

CREATE TABLE IF NOT EXISTS pharmacy.purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    supplier_id UUID REFERENCES pharmacy.suppliers(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'sent', 'cancelled')),
    to_email TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_po_practice
    ON pharmacy.purchase_orders (practice_id, created_at DESC);

CREATE TABLE IF NOT EXISTS pharmacy.purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES pharmacy.purchase_orders(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE RESTRICT,
    qty NUMERIC(12,3) NOT NULL CHECK (qty > 0),
    unit TEXT NOT NULL DEFAULT 'unit',
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_po_items_order
    ON pharmacy.purchase_order_items (order_id, sort_order);

CREATE TABLE IF NOT EXISTS pharmacy.delivery_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    supplier_id UUID REFERENCES pharmacy.suppliers(id) ON DELETE SET NULL,
    supplier_name TEXT NOT NULL DEFAULT '',
    note_number TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id, note_number)
);

CREATE TABLE IF NOT EXISTS pharmacy.delivery_note_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_note_id UUID NOT NULL REFERENCES pharmacy.delivery_notes(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE RESTRICT,
    batch_id UUID REFERENCES pharmacy.medication_batches(id) ON DELETE SET NULL,
    deposit_id UUID REFERENCES pharmacy.medication_deposits(id) ON DELETE SET NULL,
    lot_number TEXT NOT NULL,
    expires_on DATE NOT NULL,
    qty NUMERIC(12,3) NOT NULL CHECK (qty > 0),
    unit TEXT NOT NULL DEFAULT 'unit',
    sort_order INT NOT NULL DEFAULT 0
);

ALTER TABLE pharmacy.stock_movements
    ADD COLUMN IF NOT EXISTS delivery_note_id UUID REFERENCES pharmacy.delivery_notes(id) ON DELETE SET NULL;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.suppliers TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.purchase_orders TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.purchase_order_items TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.delivery_notes TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.delivery_note_items TO petsfollow_app;
  END IF;
END $$;
