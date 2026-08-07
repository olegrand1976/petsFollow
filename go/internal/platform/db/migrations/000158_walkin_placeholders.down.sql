DROP INDEX IF EXISTS pets.idx_pets_walkin_placeholder;
DROP INDEX IF EXISTS pets.idx_pets_one_walkin_per_practice;
DROP INDEX IF EXISTS identity.idx_users_walkin_placeholder;
DROP INDEX IF EXISTS identity.idx_users_one_walkin_per_practice;

ALTER TABLE pets.pets DROP COLUMN IF EXISTS is_walkin_placeholder;
ALTER TABLE identity.users DROP COLUMN IF EXISTS is_walkin_placeholder;
