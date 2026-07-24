ALTER TABLE messaging.messages
    DROP COLUMN IF EXISTS media_type,
    DROP COLUMN IF EXISTS media_url;
