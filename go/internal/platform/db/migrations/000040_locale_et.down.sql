DO $$
DECLARE
  cname text;
BEGIN
  -- Reset any et users before narrowing the CHECK
  UPDATE identity.users SET preferred_locale = 'en' WHERE preferred_locale = 'et';

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
