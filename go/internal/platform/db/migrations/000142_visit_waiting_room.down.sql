DROP INDEX IF EXISTS notifications.idx_notification_log_vet_unread;

ALTER TABLE notifications.notification_log
  DROP COLUMN IF EXISTS read_at;

ALTER TABLE visits.visits
  DROP COLUMN IF EXISTS waiting_room_at;
