-- Client identity fields for Pro fiche / create (first/last name, address, NISS).
ALTER TABLE identity.users
  ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS national_registry_number TEXT NOT NULL DEFAULT '';

-- Best-effort backfill from full_name: first token → first_name, remainder → last_name.
UPDATE identity.users
SET
  first_name = CASE
    WHEN COALESCE(first_name, '') = '' AND position(' ' IN trim(full_name)) > 0
      THEN split_part(trim(full_name), ' ', 1)
    WHEN COALESCE(first_name, '') = '' THEN trim(full_name)
    ELSE first_name
  END,
  last_name = CASE
    WHEN COALESCE(last_name, '') = '' AND position(' ' IN trim(full_name)) > 0
      THEN trim(substr(trim(full_name), position(' ' IN trim(full_name)) + 1))
    ELSE last_name
  END
WHERE COALESCE(trim(full_name), '') <> '';
