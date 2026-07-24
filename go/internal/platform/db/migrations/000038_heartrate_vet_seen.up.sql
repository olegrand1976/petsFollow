-- Track whether a validated heart-rate session has been seen by the practice vet.

ALTER TABLE heartrate.sessions
  ADD COLUMN IF NOT EXISTS vet_seen_at TIMESTAMPTZ NULL;

-- Existing validated sessions are already known to the vet — do not flood unread badges.
UPDATE heartrate.sessions
SET vet_seen_at = COALESCE(validated_at, NOW())
WHERE status = 'validated'
  AND vet_seen_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_heartrate_unread_practice
  ON heartrate.sessions (practice_id)
  WHERE status = 'validated' AND vet_seen_at IS NULL;
