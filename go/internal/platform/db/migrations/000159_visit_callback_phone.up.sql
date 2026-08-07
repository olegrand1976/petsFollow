-- Temporary callback phone for walk-in / Nouveau client bookings (visit-scoped).
ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS callback_phone TEXT NOT NULL DEFAULT '';
