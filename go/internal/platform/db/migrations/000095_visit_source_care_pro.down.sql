DO $$
DECLARE
  cname text;
BEGIN
  SELECT con.conname INTO cname
  FROM pg_constraint con
  JOIN pg_class rel ON rel.oid = con.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  WHERE nsp.nspname = 'visits'
    AND rel.relname = 'visits'
    AND con.contype = 'c'
    AND pg_get_constraintdef(con.oid) LIKE '%care_pro%'
  LIMIT 1;

  IF cname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE visits.visits DROP CONSTRAINT %I', cname);
  END IF;

  -- Revert to original CHECK (client | vet). Rows with care_pro must be rewritten first.
  UPDATE visits.visits SET source = 'vet' WHERE source = 'care_pro';

  ALTER TABLE visits.visits
    ADD CONSTRAINT visits_visits_source_check
    CHECK (source IN ('client', 'vet'));
END $$;
