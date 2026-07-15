ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS heartrate_durations_sec INTEGER[] NOT NULL DEFAULT '{60}';

UPDATE practice.practices
SET heartrate_durations_sec = '{60}'
WHERE heartrate_durations_sec IS NULL OR cardinality(heartrate_durations_sec) = 0;
