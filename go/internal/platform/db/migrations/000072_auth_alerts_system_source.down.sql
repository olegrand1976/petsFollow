DROP TABLE IF EXISTS ops.auth_alerts;

DO $$
DECLARE
  cname text;
  def text;
  is_owner boolean;
BEGIN
  SELECT (c.relowner = (SELECT oid FROM pg_roles WHERE rolname = current_user))
  INTO is_owner
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'ops' AND c.relname = 'support_tickets';

  IF NOT COALESCE(is_owner, false) THEN
    RETURN;
  END IF;

  SELECT con.conname, pg_get_constraintdef(con.oid)
  INTO cname, def
  FROM pg_constraint con
  JOIN pg_class rel ON rel.oid = con.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  WHERE nsp.nspname = 'ops'
    AND rel.relname = 'support_tickets'
    AND con.contype = 'c'
    AND pg_get_constraintdef(con.oid) ILIKE '%source%'
  LIMIT 1;

  IF cname IS NULL THEN
    RETURN;
  END IF;

  EXECUTE format('ALTER TABLE ops.support_tickets DROP CONSTRAINT %I', cname);
  ALTER TABLE ops.support_tickets ADD CONSTRAINT support_tickets_source_check
    CHECK (source IN ('nuxt_pro', 'flutter_client', 'flutter_pro_light'));
END $$;
