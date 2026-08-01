CREATE TABLE IF NOT EXISTS pets.blood_pressure_readings (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id),
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    practice_id UUID REFERENCES practice.practices(id),
    systolic_mmhg INT NOT NULL CHECK (systolic_mmhg >= 20 AND systolic_mmhg <= 400),
    diastolic_mmhg INT NOT NULL CHECK (diastolic_mmhg >= 10 AND diastolic_mmhg <= 300),
    mean_mmhg INT CHECK (mean_mmhg IS NULL OR (mean_mmhg >= 10 AND mean_mmhg <= 350)),
    method TEXT NOT NULL DEFAULT 'unknown'
        CHECK (method IN ('doppler', 'oscillometric', 'invasive', 'unknown')),
    site TEXT,
    comment TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (diastolic_mmhg <= systolic_mmhg)
);

CREATE INDEX IF NOT EXISTS idx_blood_pressure_readings_pet_recorded
    ON pets.blood_pressure_readings (pet_id, recorded_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pets.blood_pressure_readings TO petsfollow_app;
  END IF;
END $$;
