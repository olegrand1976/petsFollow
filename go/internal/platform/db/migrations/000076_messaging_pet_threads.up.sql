-- Pet-scoped messaging threads: one conversation per (practice, client, pet).
-- Keep a single legacy general thread (pet_id IS NULL) per (practice, client).

ALTER TABLE messaging.threads
    DROP CONSTRAINT IF EXISTS threads_practice_id_client_user_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_messaging_threads_practice_client_pet
    ON messaging.threads (practice_id, client_user_id, pet_id)
    WHERE pet_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_messaging_threads_practice_client_general
    ON messaging.threads (practice_id, client_user_id)
    WHERE pet_id IS NULL;
