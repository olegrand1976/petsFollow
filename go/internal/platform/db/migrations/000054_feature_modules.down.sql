ALTER TABLE notifications.client_preferences
    DROP COLUMN IF EXISTS module_care_plus,
    DROP COLUMN IF EXISTS module_horse,
    DROP COLUMN IF EXISTS module_kennel,
    DROP COLUMN IF EXISTS module_family;
