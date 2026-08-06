DROP INDEX IF EXISTS visits.idx_visits_room_scheduled;
DROP INDEX IF EXISTS visits.idx_visits_assignee_scheduled;

ALTER TABLE visits.visits
    DROP COLUMN IF EXISTS room_id,
    DROP COLUMN IF EXISTS assignee_user_id;

DROP INDEX IF EXISTS practice.idx_rooms_practice;
DROP INDEX IF EXISTS practice.idx_rooms_site;
DROP INDEX IF EXISTS practice.uq_rooms_site_name_active;
DROP TABLE IF EXISTS practice.rooms;
