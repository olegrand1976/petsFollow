-- Align ACL column default with GrantPetAccess / share API (empty → write_notes).
ALTER TABLE practice.client_access ALTER COLUMN permission SET DEFAULT 'write_notes';
ALTER TABLE pets.pet_access ALTER COLUMN permission SET DEFAULT 'write_notes';
