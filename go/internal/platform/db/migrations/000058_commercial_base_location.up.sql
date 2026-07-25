-- Base location for commercials (nearby discovery at signup without invite).
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS base_lat DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS base_lng DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS base_city TEXT,
    ADD COLUMN IF NOT EXISTS base_postal_code TEXT;

CREATE INDEX IF NOT EXISTS idx_users_commercial_base_coords
    ON identity.users (base_lat, base_lng)
    WHERE role IN ('commercial', 'commercial_manager')
      AND base_lat IS NOT NULL
      AND base_lng IS NOT NULL;
