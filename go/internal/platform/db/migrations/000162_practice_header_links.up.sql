-- Practice header quick-links prefs (catalog toggles + custom URLs).
ALTER TABLE practice.practices
  ADD COLUMN IF NOT EXISTS header_links JSONB NOT NULL DEFAULT '{}'::jsonb;
