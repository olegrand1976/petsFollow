-- Animal lifecycle status dates (all species).
ALTER TABLE pets.pets
    ADD COLUMN IF NOT EXISTS adopted_at DATE,
    ADD COLUMN IF NOT EXISTS sold_at DATE,
    ADD COLUMN IF NOT EXISTS deceased_at DATE;
