ALTER TABLE practice.practices
  DROP CONSTRAINT IF EXISTS practices_animal_scope_check;

ALTER TABLE practice.practices
  DROP COLUMN IF EXISTS animal_scope;
