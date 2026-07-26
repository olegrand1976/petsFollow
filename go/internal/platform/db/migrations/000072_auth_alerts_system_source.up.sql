-- System support tickets (ops auth alerts) + dedup table for ALERT/URGENT signals.
-- Widening source CHECK needs table ownership (Cloud Run: petsfollow-migrate-database-url).

DO $$
DECLARE
  cname text;
  already boolean;
  is_owner boolean;
BEGIN
  SELECT EXISTS (
    SELECT 1
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
    WHERE nsp.nspname = 'ops'
      AND rel.relname = 'support_tickets'
      AND con.contype = 'c'
      AND pg_get_constraintdef(con.oid) LIKE '%''system''%'
  ) INTO already;

  IF already THEN
    RETURN;
  END IF;

  SELECT (c.relowner = (SELECT oid FROM pg_roles WHERE rolname = current_user))
  INTO is_owner
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'ops' AND c.relname = 'support_tickets';

  IF NOT COALESCE(is_owner, false) THEN
    RAISE EXCEPTION
      '000072: widen ops.support_tickets.source CHECK requires table owner (current_user=%). Use petsfollow-migrate-database-url.',
      current_user;
  END IF;

  SELECT con.conname INTO cname
  FROM pg_constraint con
  JOIN pg_class rel ON rel.oid = con.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  WHERE nsp.nspname = 'ops'
    AND rel.relname = 'support_tickets'
    AND con.contype = 'c'
    AND pg_get_constraintdef(con.oid) ILIKE '%source%'
  LIMIT 1;

  IF cname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE ops.support_tickets DROP CONSTRAINT %I', cname);
  END IF;

  ALTER TABLE ops.support_tickets ADD CONSTRAINT support_tickets_source_check
    CHECK (source IN ('nuxt_pro', 'flutter_client', 'flutter_pro_light', 'system'));
END $$;

CREATE TABLE IF NOT EXISTS ops.auth_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    ticket_id UUID REFERENCES ops.support_tickets(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_auth_alerts_kind_fp_created
    ON ops.auth_alerts (kind, fingerprint, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    BEGIN
      GRANT SELECT, INSERT, UPDATE, DELETE ON ops.auth_alerts TO petsfollow_app;
    EXCEPTION WHEN insufficient_privilege THEN
      NULL;
    END;
  END IF;
END $$;
