-- Grants omitted from 000138 (schema client_ai + visit_report_explanations without petsfollow_app).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA client_ai TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON client_ai.triage_sessions TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON client_ai.triage_messages TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON client_ai.usage_events TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON visits.visit_report_explanations TO petsfollow_app;
  END IF;
END $$;
