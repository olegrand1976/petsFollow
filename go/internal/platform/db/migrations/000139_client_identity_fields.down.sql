ALTER TABLE identity.users
  DROP COLUMN IF EXISTS national_registry_number,
  DROP COLUMN IF EXISTS address,
  DROP COLUMN IF EXISTS last_name,
  DROP COLUMN IF EXISTS first_name;
