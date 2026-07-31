-- Ordonnances vétérinaires (V1 : brouillons + preview PDF ; signature / envoi phase 2).
CREATE SCHEMA IF NOT EXISTS prescriptions;

CREATE TABLE IF NOT EXISTS prescriptions.prescriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    veterinary_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE RESTRICT,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE RESTRICT,
    country_code CHAR(2) NOT NULL DEFAULT 'BE',
    date_issued TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    signature_id UUID,
    pdf_object_key TEXT,
    pdf_sha256 TEXT,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'signed', 'sent', 'archived')),
    paper_format TEXT NOT NULL DEFAULT 'A4'
        CHECK (paper_format IN ('A4', 'A5')),
    medications JSONB NOT NULL DEFAULT '[]'::jsonb,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_prescriptions_practice_status
    ON prescriptions.prescriptions (practice_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prescriptions_pet
    ON prescriptions.prescriptions (pet_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prescriptions_owner
    ON prescriptions.prescriptions (owner_id, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA prescriptions TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON prescriptions.prescriptions TO petsfollow_app;
  END IF;
END $$;
