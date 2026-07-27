-- Pharmacie BE — DAF (documents, séquences gapless, lignes).
CREATE TABLE IF NOT EXISTS pharmacy.daf_sequences (
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    daf_year INT NOT NULL CHECK (daf_year >= 2000 AND daf_year <= 2100),
    next_number BIGINT NOT NULL DEFAULT 1 CHECK (next_number >= 1),
    PRIMARY KEY (practice_id, daf_year)
);

CREATE TABLE IF NOT EXISTS pharmacy.daf_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    daf_year INT NOT NULL CHECK (daf_year >= 2000 AND daf_year <= 2100),
    daf_number BIGINT,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'finalized', 'cancelled')),
    client_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    pet_id UUID REFERENCES pets.pets(id) ON DELETE SET NULL,
    prescriber_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE RESTRICT,
    notes TEXT,
    issued_at TIMESTAMPTZ,
    finalized_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancel_reason TEXT,
    pdf_object_key TEXT,
    pdf_sha256 TEXT,
    has_antibiotic BOOLEAN NOT NULL DEFAULT false,
    vamreg_status TEXT NOT NULL DEFAULT 'n/a'
        CHECK (vamreg_status IN ('n/a', 'pending', 'sent', 'failed')),
    invoices_export_status TEXT NOT NULL DEFAULT 'n/a'
        CHECK (invoices_export_status IN ('n/a', 'pending', 'sent', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pharmacy_daf_number_when_finalized CHECK (
        (status = 'draft' AND daf_number IS NULL)
        OR (status IN ('finalized', 'cancelled') AND daf_number IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pharmacy_daf_number_unique
    ON pharmacy.daf_documents (practice_id, daf_year, daf_number)
    WHERE daf_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_pharmacy_daf_practice_status
    ON pharmacy.daf_documents (practice_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS pharmacy.daf_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    daf_id UUID NOT NULL REFERENCES pharmacy.daf_documents(id) ON DELETE RESTRICT,
    medication_id UUID NOT NULL REFERENCES pharmacy.ref_medications(id) ON DELETE RESTRICT,
    batch_id UUID REFERENCES pharmacy.medication_batches(id) ON DELETE RESTRICT,
    deposit_id UUID REFERENCES pharmacy.medication_deposits(id) ON DELETE SET NULL,
    qty NUMERIC(12,3) NOT NULL CHECK (qty > 0),
    unit TEXT NOT NULL DEFAULT 'unit',
    amm_number TEXT NOT NULL DEFAULT '',
    is_antibiotic BOOLEAN NOT NULL DEFAULT false,
    vamreg_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_daf_items_daf
    ON pharmacy.daf_items (daf_id, sort_order);

-- Lier les mouvements au DAF (colonne déjà présente sans FK).
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'stock_movements_daf_item_id_fkey'
  ) THEN
    ALTER TABLE pharmacy.stock_movements
      ADD CONSTRAINT stock_movements_daf_item_id_fkey
      FOREIGN KEY (daf_item_id) REFERENCES pharmacy.daf_items(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.daf_sequences TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.daf_documents TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.daf_items TO petsfollow_app;
  END IF;
END $$;
