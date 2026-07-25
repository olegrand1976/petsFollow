DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    REVOKE SELECT, INSERT, UPDATE, DELETE ON practice.client_import_jobs FROM petsfollow_app;
    REVOKE SELECT, INSERT, UPDATE, DELETE ON practice.client_import_rows FROM petsfollow_app;
  END IF;
END $$;
