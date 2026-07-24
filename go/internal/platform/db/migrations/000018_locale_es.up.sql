-- Widen preferred_locale CHECK to include Spanish (es).
-- Drop any existing CHECK on preferred_locale by definition lookup
-- (000005 used an inline unnamed constraint; PG auto-name may vary).

DO $$
DECLARE
  cname text;
BEGIN
  SELECT con.conname INTO cname
  FROM pg_constraint con
  JOIN pg_class rel ON rel.oid = con.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  WHERE nsp.nspname = 'identity'
    AND rel.relname = 'users'
    AND con.contype = 'c'
    AND pg_get_constraintdef(con.oid) ILIKE '%preferred_locale%'
  LIMIT 1;

  IF cname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE identity.users DROP CONSTRAINT %I', cname);
  END IF;
END $$;

ALTER TABLE identity.users
  ADD CONSTRAINT users_preferred_locale_check
  CHECK (preferred_locale IN ('fr', 'nl', 'en', 'es'));
