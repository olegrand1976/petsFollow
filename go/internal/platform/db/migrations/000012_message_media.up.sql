ALTER TABLE messaging.messages
    ADD COLUMN IF NOT EXISTS media_url TEXT,
    ADD COLUMN IF NOT EXISTS media_type TEXT
        CHECK (media_type IS NULL OR media_type IN ('image', 'video'));
