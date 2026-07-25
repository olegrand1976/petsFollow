DROP INDEX IF EXISTS identity.idx_users_commercial_base_coords;

ALTER TABLE identity.users
    DROP COLUMN IF EXISTS base_lat,
    DROP COLUMN IF EXISTS base_lng,
    DROP COLUMN IF EXISTS base_city,
    DROP COLUMN IF EXISTS base_postal_code;
