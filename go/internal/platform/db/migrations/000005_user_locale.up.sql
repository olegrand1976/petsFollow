ALTER TABLE identity.users
  ADD COLUMN preferred_locale TEXT NOT NULL DEFAULT 'fr'
  CHECK (preferred_locale IN ('fr', 'nl', 'en'));
