DROP TABLE IF EXISTS visits.visit_reports;

ALTER TABLE visits.visits
    DROP COLUMN IF EXISTS address_text,
    DROP COLUMN IF EXISTS lat,
    DROP COLUMN IF EXISTS lng;

DROP TABLE IF EXISTS pets.pet_access;
DROP TABLE IF EXISTS practice.client_access;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_care_pro_specialty;
ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_specialty_check;
ALTER TABLE identity.users DROP COLUMN IF EXISTS professional_specialty;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('vet', 'client', 'admin', 'commercial', 'commercial_manager'));
