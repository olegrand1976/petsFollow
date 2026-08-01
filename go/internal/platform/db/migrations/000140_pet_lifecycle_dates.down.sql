ALTER TABLE pets.pets
    DROP COLUMN IF EXISTS deceased_at,
    DROP COLUMN IF EXISTS sold_at,
    DROP COLUMN IF EXISTS adopted_at;
