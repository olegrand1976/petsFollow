-- Data room access is opt-in per group (admin-only toggle) — prevents solo-group unlock.
ALTER TABLE research.groups
    ADD COLUMN IF NOT EXISTS dataroom_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_research_groups_dataroom
    ON research.groups (dataroom_enabled)
    WHERE dataroom_enabled = true;
