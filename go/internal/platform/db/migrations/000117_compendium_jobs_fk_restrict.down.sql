ALTER TABLE pharmacy.compendium_import_jobs
  DROP CONSTRAINT IF EXISTS compendium_import_jobs_created_by_admin_id_fkey;

ALTER TABLE pharmacy.compendium_import_jobs
  ADD CONSTRAINT compendium_import_jobs_created_by_admin_id_fkey
    FOREIGN KEY (created_by_admin_id) REFERENCES identity.users(id) ON DELETE CASCADE;
