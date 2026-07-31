-- Role DEV (support IT applicatif) — ops sans billing / seed.

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN (
        'vet', 'client', 'admin', 'dev', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    ));

ALTER TABLE identity.profiles DROP CONSTRAINT IF EXISTS profiles_role_check;
ALTER TABLE identity.profiles
    ADD CONSTRAINT profiles_role_check
    CHECK (role IN (
        'vet', 'client', 'admin', 'dev', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    ));
