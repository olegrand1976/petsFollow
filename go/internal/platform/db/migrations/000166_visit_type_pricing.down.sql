ALTER TABLE practice.visit_types
    DROP CONSTRAINT IF EXISTS visit_types_vat_percent_check;

ALTER TABLE practice.visit_types
    DROP CONSTRAINT IF EXISTS visit_types_price_excl_cents_check;

ALTER TABLE practice.visit_types
    DROP COLUMN IF EXISTS vat_percent,
    DROP COLUMN IF EXISTS price_excl_cents;
