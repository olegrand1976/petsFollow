-- Grant petsfollow_app access to enrichment schemas created in 000010.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA care TO petsfollow_app;
    GRANT USAGE ON SCHEMA visits TO petsfollow_app;
    GRANT USAGE ON SCHEMA discovery TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA care TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA visits TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA discovery TO petsfollow_app;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA care TO petsfollow_app;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA visits TO petsfollow_app;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA discovery TO petsfollow_app;
    ALTER DEFAULT PRIVILEGES IN SCHEMA care GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO petsfollow_app;
    ALTER DEFAULT PRIVILEGES IN SCHEMA visits GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO petsfollow_app;
    ALTER DEFAULT PRIVILEGES IN SCHEMA discovery GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO petsfollow_app;
  END IF;
END $$;
