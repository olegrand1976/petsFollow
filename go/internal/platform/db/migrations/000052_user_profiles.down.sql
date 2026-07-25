UPDATE identity.users SET active_profile_id = NULL;
ALTER TABLE identity.users DROP COLUMN IF EXISTS active_profile_id;
DROP TABLE IF EXISTS identity.profiles;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('vet', 'client', 'admin', 'commercial', 'commercial_manager', 'care_pro'));
