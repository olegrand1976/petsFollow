-- Care_pro messaging: person-scoped threads (no cabinet practice_id).
-- Distinct from practice-scoped unique indexes (000076).

ALTER TABLE messaging.threads
    ALTER COLUMN practice_id DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_messaging_threads_care_pro_client_pet
    ON messaging.threads (vet_user_id, client_user_id, pet_id)
    WHERE practice_id IS NULL AND pet_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_messaging_threads_care_pro_client_general
    ON messaging.threads (vet_user_id, client_user_id)
    WHERE practice_id IS NULL AND pet_id IS NULL;
