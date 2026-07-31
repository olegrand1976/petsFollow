ALTER TABLE practice.practices
  DROP CONSTRAINT IF EXISTS practices_desk_idle_minutes_check;

ALTER TABLE practice.practices
  DROP COLUMN IF EXISTS desk_idle_minutes;
