-- Grant petsfollow_app access to veille news table (omitted in 000163).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA ops TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ops.vet_news_articles TO petsfollow_app;
  END IF;
END $$;
