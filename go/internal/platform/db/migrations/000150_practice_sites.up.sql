-- Multi-sites (org + sites enfants) : lieux physiques sous un practice tenant.
-- Agenda / vacances / RDV scopés par site_id ; 1 site primary par practice.

CREATE TABLE IF NOT EXISTS practice.sites (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    address_line1 TEXT NOT NULL DEFAULT '',
    address_line2 TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    postal_code TEXT NOT NULL DEFAULT '',
    country_code TEXT NOT NULL DEFAULT 'BE',
    timezone TEXT NOT NULL DEFAULT 'Europe/Brussels',
    is_primary BOOLEAN NOT NULL DEFAULT false,
    active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS sites_one_primary_per_practice
    ON practice.sites (practice_id) WHERE is_primary;

CREATE INDEX IF NOT EXISTS sites_practice_idx
    ON practice.sites (practice_id) WHERE active;

GRANT SELECT, INSERT, UPDATE, DELETE ON practice.sites TO CURRENT_USER;

-- Backfill : 1 site primary par practice (copie contact + timezone schedule si présent).
INSERT INTO practice.sites (
    id, practice_id, name, phone, address_line1, address_line2, city, postal_code,
    country_code, timezone, is_primary, active, sort_order
)
SELECT
    gen_random_uuid(),
    p.id,
    COALESCE(NULLIF(TRIM(p.name), ''), 'Principal'),
    COALESCE(p.phone, ''),
    COALESCE(p.address_line1, ''),
    COALESCE(p.address_line2, ''),
    COALESCE(p.city, ''),
    COALESCE(p.postal_code, ''),
    COALESCE(NULLIF(TRIM(p.country_code), ''), 'BE'),
    COALESCE(vs.timezone, 'Europe/Brussels'),
    true,
    true,
    0
FROM practice.practices p
LEFT JOIN practice.vet_schedule vs ON vs.practice_id = p.id
WHERE NOT EXISTS (
    SELECT 1 FROM practice.sites s WHERE s.practice_id = p.id AND s.is_primary
);

-- vet_schedule : PK → site_id, garder practice_id dénormalisé.
ALTER TABLE practice.vet_schedule
    ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES practice.sites(id) ON DELETE CASCADE;

UPDATE practice.vet_schedule vs
SET site_id = s.id
FROM practice.sites s
WHERE s.practice_id = vs.practice_id AND s.is_primary AND vs.site_id IS NULL;

-- Practices without a schedule row yet: nothing to update.
-- Drop old PK and recreate on site_id (only rows with site_id remain usable).
ALTER TABLE practice.vet_schedule DROP CONSTRAINT IF EXISTS vet_schedule_pkey;

-- Orphan schedule rows without site (should not happen after backfill) — drop them.
DELETE FROM practice.vet_schedule WHERE site_id IS NULL;

ALTER TABLE practice.vet_schedule
    ALTER COLUMN site_id SET NOT NULL;

ALTER TABLE practice.vet_schedule
    ADD PRIMARY KEY (site_id);

CREATE INDEX IF NOT EXISTS idx_vet_schedule_practice
    ON practice.vet_schedule (practice_id);

-- slots
ALTER TABLE practice.vet_schedule_slots
    ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES practice.sites(id) ON DELETE CASCADE;

UPDATE practice.vet_schedule_slots sl
SET site_id = s.id
FROM practice.sites s
WHERE s.practice_id = sl.practice_id AND s.is_primary AND sl.site_id IS NULL;

DELETE FROM practice.vet_schedule_slots WHERE site_id IS NULL;

ALTER TABLE practice.vet_schedule_slots
    ALTER COLUMN site_id SET NOT NULL;

DROP INDEX IF EXISTS practice.idx_vet_schedule_slots_practice;
CREATE INDEX IF NOT EXISTS idx_vet_schedule_slots_site
    ON practice.vet_schedule_slots (site_id);
CREATE INDEX IF NOT EXISTS idx_vet_schedule_slots_practice
    ON practice.vet_schedule_slots (practice_id);

-- vacations
ALTER TABLE practice.vet_vacations
    ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES practice.sites(id) ON DELETE CASCADE;

UPDATE practice.vet_vacations v
SET site_id = s.id
FROM practice.sites s
WHERE s.practice_id = v.practice_id AND s.is_primary AND v.site_id IS NULL;

DELETE FROM practice.vet_vacations WHERE site_id IS NULL;

ALTER TABLE practice.vet_vacations
    ALTER COLUMN site_id SET NOT NULL;

DROP INDEX IF EXISTS practice.idx_vet_vacations_practice;
CREATE INDEX IF NOT EXISTS idx_vet_vacations_site
    ON practice.vet_vacations (site_id, starts_on, ends_on);
CREATE INDEX IF NOT EXISTS idx_vet_vacations_practice
    ON practice.vet_vacations (practice_id, starts_on, ends_on);

-- visits
ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES practice.sites(id);

UPDATE visits.visits v
SET site_id = s.id
FROM practice.sites s
WHERE s.practice_id = v.practice_id AND s.is_primary AND v.site_id IS NULL;

-- Safety: any visit whose practice vanished should not remain; force NOT NULL only when all set.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM visits.visits WHERE site_id IS NULL) THEN
    RAISE EXCEPTION 'visits.visits backfill incomplete: null site_id remain';
  END IF;
END $$;

ALTER TABLE visits.visits
    ALTER COLUMN site_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_visits_site_scheduled
    ON visits.visits (site_id, scheduled_at)
    WHERE deleted_at IS NULL AND status IN ('requested', 'confirmed', 'reschedule_pending');

-- team preference
ALTER TABLE practice.team_members
    ADD COLUMN IF NOT EXISTS default_site_id UUID REFERENCES practice.sites(id) ON DELETE SET NULL;
