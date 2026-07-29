DELETE FROM identity.profiles WHERE role = 'dev';
UPDATE identity.users SET role = 'admin' WHERE role = 'dev';

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN (
        'vet', 'client', 'admin', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    ));

ALTER TABLE identity.profiles DROP CONSTRAINT IF EXISTS profiles_role_check;
ALTER TABLE identity.profiles
    ADD CONSTRAINT profiles_role_check
    CHECK (role IN (
        'vet', 'client', 'admin', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    ));
