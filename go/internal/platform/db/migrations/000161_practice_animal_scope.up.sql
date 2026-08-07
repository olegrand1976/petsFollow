-- Practice clinical focus for AFSCA newsletter filtering (BE dashboard).
ALTER TABLE practice.practices
  ADD COLUMN IF NOT EXISTS animal_scope TEXT NOT NULL DEFAULT 'both';

ALTER TABLE practice.practices
  DROP CONSTRAINT IF EXISTS practices_animal_scope_check;

ALTER TABLE practice.practices
  ADD CONSTRAINT practices_animal_scope_check
  CHECK (animal_scope IN ('small', 'large', 'both'));
