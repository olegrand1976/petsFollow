-- Allow pets without a linked practice (vet linked after registration).
ALTER TABLE pets.pets
    ALTER COLUMN practice_id DROP NOT NULL;
