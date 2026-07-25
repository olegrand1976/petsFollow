-- Revert monthly; fails if any monthly entitlements exist.
ALTER TABLE billing.pet_entitlements
    DROP CONSTRAINT IF EXISTS pet_entitlements_plan_code_check;

ALTER TABLE billing.pet_entitlements
    ADD CONSTRAINT pet_entitlements_plan_code_check
    CHECK (plan_code IN ('annual', 'triennial', 'quinquennial'));
