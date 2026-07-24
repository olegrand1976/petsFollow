-- Optional client comment attached when validating a heart-rate reading.

ALTER TABLE heartrate.sessions
  ADD COLUMN IF NOT EXISTS comment TEXT NULL;
