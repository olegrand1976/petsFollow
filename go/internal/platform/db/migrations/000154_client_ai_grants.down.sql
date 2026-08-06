DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    REVOKE SELECT, INSERT, UPDATE, DELETE ON visits.visit_report_explanations FROM petsfollow_app;
    REVOKE SELECT, INSERT, UPDATE, DELETE ON client_ai.usage_events FROM petsfollow_app;
    REVOKE SELECT, INSERT, UPDATE, DELETE ON client_ai.triage_messages FROM petsfollow_app;
    REVOKE SELECT, INSERT, UPDATE, DELETE ON client_ai.triage_sessions FROM petsfollow_app;
    REVOKE USAGE ON SCHEMA client_ai FROM petsfollow_app;
  END IF;
END $$;
