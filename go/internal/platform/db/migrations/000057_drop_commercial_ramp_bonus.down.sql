-- Irreversible content delete of ramp awards; restore CHECK to allow commercial_ramp for schema symmetry only.

ALTER TABLE billing.commercial_bonus_awards
    DROP CONSTRAINT IF EXISTS commercial_bonus_awards_bonus_code_check;

ALTER TABLE billing.commercial_bonus_awards
    ADD CONSTRAINT commercial_bonus_awards_bonus_code_check
    CHECK (bonus_code IN ('commercial_ramp', 'commercial_mix'));

ALTER TABLE billing.commercial_bonus_awards
    DROP CONSTRAINT IF EXISTS commercial_bonus_ramp_fields;

ALTER TABLE billing.commercial_bonus_awards
    ADD CONSTRAINT commercial_bonus_ramp_fields CHECK (
        bonus_code <> 'commercial_ramp' OR vet_user_id IS NOT NULL
    );
