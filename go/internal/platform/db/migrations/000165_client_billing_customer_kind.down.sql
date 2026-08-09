ALTER TABLE identity.users
  DROP CONSTRAINT IF EXISTS users_billing_customer_kind_check;

ALTER TABLE identity.users
  DROP COLUMN IF EXISTS billing_customer_kind;
