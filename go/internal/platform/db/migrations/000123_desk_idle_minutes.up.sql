-- Idle lock timeout for shared-desk workstation (minutes). Default 2.
ALTER TABLE practice.practices
  ADD COLUMN IF NOT EXISTS desk_idle_minutes INTEGER NOT NULL DEFAULT 2;

ALTER TABLE practice.practices
  DROP CONSTRAINT IF EXISTS practices_desk_idle_minutes_check;

ALTER TABLE practice.practices
  ADD CONSTRAINT practices_desk_idle_minutes_check
  CHECK (desk_idle_minutes IN (1, 2, 5, 10, 15, 30));
