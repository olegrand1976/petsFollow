ALTER TABLE pharmacy.daf_items
    DROP COLUMN IF EXISTS withdrawal_meat_days,
    DROP COLUMN IF EXISTS withdrawal_milk_days,
    DROP COLUMN IF EXISTS withdrawal_eggs_days;

ALTER TABLE pharmacy.ref_medications
    DROP COLUMN IF EXISTS withdrawal_meat_days,
    DROP COLUMN IF EXISTS withdrawal_milk_days,
    DROP COLUMN IF EXISTS withdrawal_eggs_days,
    DROP COLUMN IF EXISTS food_chain_banned;
