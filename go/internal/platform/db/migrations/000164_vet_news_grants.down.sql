DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    REVOKE SELECT, INSERT, UPDATE, DELETE ON ops.vet_news_articles FROM petsfollow_app;
  END IF;
END $$;
