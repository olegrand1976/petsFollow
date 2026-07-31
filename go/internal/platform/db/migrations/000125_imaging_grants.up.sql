-- Grants omitted from 000124 (schema imaging created without petsfollow_app USAGE).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA imaging TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON imaging.pet_studies TO petsfollow_app;
  END IF;
END $$;
