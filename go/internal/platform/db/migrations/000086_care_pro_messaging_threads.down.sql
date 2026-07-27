DROP INDEX IF EXISTS messaging.uq_messaging_threads_care_pro_client_general;
DROP INDEX IF EXISTS messaging.uq_messaging_threads_care_pro_client_pet;

-- Only restore NOT NULL if no NULL practice_id rows remain.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM messaging.threads WHERE practice_id IS NULL) THEN
    ALTER TABLE messaging.threads ALTER COLUMN practice_id SET NOT NULL;
  END IF;
END $$;
