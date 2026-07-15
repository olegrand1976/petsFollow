DROP TABLE IF EXISTS billing.stripe_events;
DROP TABLE IF EXISTS billing.pet_entitlements;
DROP TABLE IF EXISTS billing.stripe_customers;
ALTER TABLE pets.pets DROP COLUMN IF EXISTS payment_status;
DROP SCHEMA IF EXISTS billing CASCADE;
