DROP TABLE IF EXISTS research.etl_watermarks;
DROP TABLE IF EXISTS research.weekly_aggregates;
DROP TABLE IF EXISTS research.anon_events;
DROP SCHEMA IF EXISTS research;

ALTER TABLE practice.practices DROP COLUMN IF EXISTS research_opt_in_by;
ALTER TABLE practice.practices DROP COLUMN IF EXISTS research_opt_in_at;
