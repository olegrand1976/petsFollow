-- Reverse multi-sites: collapse to practice-scoped schedule (dev only; fails if >1 site/practice).

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM practice.sites
    GROUP BY practice_id HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION 'cannot down-migrate practice_sites: practices with multiple sites exist';
  END IF;
END $$;

ALTER TABLE practice.team_members DROP COLUMN IF EXISTS default_site_id;

DROP INDEX IF EXISTS visits.idx_visits_site_scheduled;
ALTER TABLE visits.visits DROP COLUMN IF EXISTS site_id;

DROP INDEX IF EXISTS practice.idx_vet_vacations_site;
ALTER TABLE practice.vet_vacations DROP COLUMN IF EXISTS site_id;
DROP INDEX IF EXISTS practice.idx_vet_vacations_practice;
CREATE INDEX IF NOT EXISTS idx_vet_vacations_practice
    ON practice.vet_vacations (practice_id, starts_on, ends_on);

DROP INDEX IF EXISTS practice.idx_vet_schedule_slots_site;
ALTER TABLE practice.vet_schedule_slots DROP COLUMN IF EXISTS site_id;

ALTER TABLE practice.vet_schedule DROP CONSTRAINT IF EXISTS vet_schedule_pkey;
ALTER TABLE practice.vet_schedule DROP COLUMN IF EXISTS site_id;
ALTER TABLE practice.vet_schedule ADD PRIMARY KEY (practice_id);
DROP INDEX IF EXISTS practice.idx_vet_schedule_practice;

DROP INDEX IF EXISTS practice.sites_practice_idx;
DROP INDEX IF EXISTS practice.sites_one_primary_per_practice;
DROP TABLE IF EXISTS practice.sites;
