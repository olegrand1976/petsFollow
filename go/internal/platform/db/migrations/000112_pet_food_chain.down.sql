ALTER TABLE pets.pets DROP CONSTRAINT IF EXISTS pets_food_chain_status_check;
ALTER TABLE pets.pets DROP COLUMN IF EXISTS food_chain_status;
