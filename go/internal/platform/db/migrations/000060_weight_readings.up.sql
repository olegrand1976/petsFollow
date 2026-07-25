CREATE TABLE IF NOT EXISTS pets.weight_readings (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id),
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    weight_kg NUMERIC(5,2) NOT NULL CHECK (weight_kg > 0 AND weight_kg <= 999.99),
    comment TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_weight_readings_pet_recorded
    ON pets.weight_readings (pet_id, recorded_at DESC);
