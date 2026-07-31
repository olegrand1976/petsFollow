-- Pharmacie BE — dictionnaire médicaments national (CNK / AFMPS).
-- pg_trgm : autocomplete flou (opérateur %). Normalisation accents = côté app Go.
-- Ops Cloud SQL : activer l'extension pg_trgm (souvent via user cloudsqlsuperuser) avant migrate.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE SCHEMA IF NOT EXISTS pharmacy;

CREATE TABLE IF NOT EXISTS pharmacy.ref_medications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cnk TEXT NOT NULL,
    name TEXT NOT NULL,
    name_normalized TEXT NOT NULL,
    atc_code TEXT,
    pharmaceutical_form TEXT,
    pack_size TEXT,
    is_antibiotic BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    afmps_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pharmacy_ref_medications_cnk_unique UNIQUE (cnk)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_ref_med_name_trgm
    ON pharmacy.ref_medications USING gin (name_normalized gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_pharmacy_ref_med_antibiotic
    ON pharmacy.ref_medications (is_antibiotic)
    WHERE is_active;

CREATE INDEX IF NOT EXISTS idx_pharmacy_ref_med_active_cnk
    ON pharmacy.ref_medications (cnk)
    WHERE is_active;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA pharmacy TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.ref_medications TO petsfollow_app;
  END IF;
END $$;
