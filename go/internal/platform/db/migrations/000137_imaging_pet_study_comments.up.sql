CREATE TABLE IF NOT EXISTS imaging.pet_study_comments (
    id UUID PRIMARY KEY,
    pet_study_id UUID NOT NULL REFERENCES imaging.pet_studies(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    body TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS pet_study_comments_study_created_idx
    ON imaging.pet_study_comments (pet_study_id, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    -- Append-only clinical notes: no UPDATE/DELETE from the app role.
    GRANT SELECT, INSERT ON imaging.pet_study_comments TO petsfollow_app;
  END IF;
END $$;
