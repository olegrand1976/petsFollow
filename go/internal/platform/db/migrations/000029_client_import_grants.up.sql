-- Grants for client import tables (000028 omitted petsfollow_app).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.client_import_jobs TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.client_import_rows TO petsfollow_app;
  END IF;
END $$;
