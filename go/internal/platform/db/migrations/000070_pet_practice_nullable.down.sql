-- Re-require practice_id only when every pet already has one.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pets.pets WHERE practice_id IS NULL) THEN
        RAISE EXCEPTION 'cannot restore NOT NULL: pets with null practice_id exist';
    END IF;
END $$;

ALTER TABLE pets.pets
    ALTER COLUMN practice_id SET NOT NULL;
