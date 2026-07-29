-- Horse domicile (écurie / lieu de détention) — DAF / chaîne alimentaire
ALTER TABLE pets.pets
    ADD COLUMN IF NOT EXISTS domicile_location TEXT NOT NULL DEFAULT '';
