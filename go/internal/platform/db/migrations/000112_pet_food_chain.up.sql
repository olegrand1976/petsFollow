-- Phase 4.D — statut chaîne alimentaire animal
ALTER TABLE pets.pets
    ADD COLUMN IF NOT EXISTS food_chain_status TEXT NOT NULL DEFAULT 'companion';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'pets_food_chain_status_check'
  ) THEN
    ALTER TABLE pets.pets
      ADD CONSTRAINT pets_food_chain_status_check
      CHECK (food_chain_status IN ('companion', 'food_producing', 'excluded_from_food_chain'));
  END IF;
END $$;
