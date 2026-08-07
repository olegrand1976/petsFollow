ALTER TABLE identity.users
  DROP COLUMN IF EXISTS billing_vat_number,
  DROP COLUMN IF EXISTS billing_company_number,
  DROP COLUMN IF EXISTS billing_street,
  DROP COLUMN IF EXISTS billing_city,
  DROP COLUMN IF EXISTS billing_postal,
  DROP COLUMN IF EXISTS billing_country;
