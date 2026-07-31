-- Distinguish care_pro terrain visits from cabinet (vet) bookings.
DO $$
DECLARE
  cname text;
  already boolean;
BEGIN
  SELECT EXISTS (
    SELECT 1
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
    WHERE nsp.nspname = 'visits'
      AND rel.relname = 'visits'
      AND con.contype = 'c'
      AND pg_get_constraintdef(con.oid) LIKE '%care_pro%'
  ) INTO already;

  IF already THEN
    RETURN;
  END IF;

  SELECT con.conname INTO cname
  FROM pg_constraint con
  JOIN pg_class rel ON rel.oid = con.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  WHERE nsp.nspname = 'visits'
    AND rel.relname = 'visits'
    AND con.contype = 'c'
    AND pg_get_constraintdef(con.oid) LIKE '%source%'
    AND pg_get_constraintdef(con.oid) LIKE '%client%'
  LIMIT 1;

  IF cname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE visits.visits DROP CONSTRAINT %I', cname);
  END IF;

  ALTER TABLE visits.visits
    ADD CONSTRAINT visits_visits_source_check
    CHECK (source IN ('client', 'vet', 'care_pro'));
END $$;
