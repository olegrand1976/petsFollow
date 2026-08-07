-- Coordonnées fiscales client (Peppol / facturation) — compte identity.users.
-- L'adresse libre `address` reste pour la fiche générale.
ALTER TABLE identity.users
  ADD COLUMN IF NOT EXISTS billing_vat_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS billing_company_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS billing_street TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS billing_city TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS billing_postal TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS billing_country TEXT NOT NULL DEFAULT '';
