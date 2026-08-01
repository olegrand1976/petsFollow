DROP INDEX IF EXISTS research.idx_research_groups_dataroom;
ALTER TABLE research.groups DROP COLUMN IF EXISTS dataroom_enabled;
