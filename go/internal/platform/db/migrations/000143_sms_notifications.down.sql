DROP INDEX IF EXISTS notifications.sms_log_user_created_idx;
DROP INDEX IF EXISTS notifications.sms_log_reminder_once;
DROP TABLE IF EXISTS notifications.sms_log;
ALTER TABLE notifications.client_preferences DROP COLUMN IF EXISTS sms;
