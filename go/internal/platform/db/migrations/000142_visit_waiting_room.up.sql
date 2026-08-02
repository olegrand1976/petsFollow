-- Desk: client in waiting room + mark desk alerts read in Pro topbar.
ALTER TABLE visits.visits
  ADD COLUMN IF NOT EXISTS waiting_room_at TIMESTAMPTZ;

ALTER TABLE notifications.notification_log
  ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_notification_log_vet_unread
  ON notifications.notification_log (vet_user_id, created_at DESC)
  WHERE read_at IS NULL;
