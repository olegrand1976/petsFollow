CREATE TABLE IF NOT EXISTS practice.eid_readings (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    tool TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT false,
    fields_read TEXT[] NOT NULL DEFAULT '{}',
    niss_hash TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS eid_readings_practice_created_idx
    ON practice.eid_readings (practice_id, created_at DESC);
