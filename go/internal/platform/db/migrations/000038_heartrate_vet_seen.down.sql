DROP INDEX IF EXISTS heartrate.idx_heartrate_unread_practice;

ALTER TABLE heartrate.sessions
  DROP COLUMN IF EXISTS vet_seen_at;
