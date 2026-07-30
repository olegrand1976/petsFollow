-- created_by_admin_id must not CASCADE from identity.users (rétention 5 ans / TestPharmacyPlanCoverage 4F).
ALTER TABLE pharmacy.compendium_import_jobs
  DROP CONSTRAINT IF EXISTS compendium_import_jobs_created_by_admin_id_fkey;

ALTER TABLE pharmacy.compendium_import_jobs
  ADD CONSTRAINT compendium_import_jobs_created_by_admin_id_fkey
    FOREIGN KEY (created_by_admin_id) REFERENCES identity.users(id) ON DELETE RESTRICT;
