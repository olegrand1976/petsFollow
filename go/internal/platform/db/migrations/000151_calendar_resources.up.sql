-- Calendar resources (model 2): work rooms under sites + visit assignee/room.

CREATE TABLE IF NOT EXISTS practice.rooms (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES practice.sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rooms_name_not_blank CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_site_name_active
    ON practice.rooms (site_id, lower(trim(name)))
    WHERE active;

CREATE INDEX IF NOT EXISTS idx_rooms_site ON practice.rooms(site_id) WHERE active;
CREATE INDEX IF NOT EXISTS idx_rooms_practice ON practice.rooms(practice_id);

ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS assignee_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS room_id UUID REFERENCES practice.rooms(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_visits_assignee_scheduled
    ON visits.visits (practice_id, assignee_user_id, scheduled_at)
    WHERE assignee_user_id IS NOT NULL
      AND deleted_at IS NULL
      AND status IN ('requested', 'confirmed', 'reschedule_pending');

CREATE INDEX IF NOT EXISTS idx_visits_room_scheduled
    ON visits.visits (room_id, scheduled_at)
    WHERE room_id IS NOT NULL
      AND deleted_at IS NULL
      AND status IN ('requested', 'confirmed', 'reschedule_pending');
