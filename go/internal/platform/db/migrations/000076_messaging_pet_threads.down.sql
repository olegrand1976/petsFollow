DROP INDEX IF EXISTS messaging.uq_messaging_threads_practice_client_general;
DROP INDEX IF EXISTS messaging.uq_messaging_threads_practice_client_pet;

-- Restore legacy uniqueness (may fail if pet-scoped duplicates exist).
ALTER TABLE messaging.threads
    ADD CONSTRAINT threads_practice_id_client_user_id_key UNIQUE (practice_id, client_user_id);
