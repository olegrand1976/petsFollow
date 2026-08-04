DROP INDEX IF EXISTS notifications.sms_inbound_user_created_idx;
DROP INDEX IF EXISTS notifications.sms_inbound_provider_message_once;
DROP TABLE IF EXISTS notifications.sms_inbound;
DROP INDEX IF EXISTS notifications.sms_log_provider_message_idx;
ALTER TABLE notifications.sms_log
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS delivery_error,
    DROP COLUMN IF EXISTS delivery_status;
