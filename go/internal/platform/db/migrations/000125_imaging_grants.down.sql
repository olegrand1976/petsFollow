DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    REVOKE SELECT, INSERT, UPDATE, DELETE ON imaging.pet_studies FROM petsfollow_app;
    REVOKE USAGE ON SCHEMA imaging FROM petsfollow_app;
  END IF;
END $$;
