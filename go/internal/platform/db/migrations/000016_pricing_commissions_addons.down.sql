DROP TABLE IF EXISTS care.competitions;
DROP TABLE IF EXISTS care.professional_contacts;

ALTER TABLE care.reminders DROP COLUMN IF EXISTS recurrence_days;
ALTER TABLE care.reminders DROP COLUMN IF EXISTS notes;
ALTER TABLE care.reminders DROP CONSTRAINT IF EXISTS reminders_type_check;
ALTER TABLE care.reminders
    ADD CONSTRAINT reminders_type_check CHECK (type IN (
        'vaccination', 'deworming', 'vet_check', 'dental',
        'farrier', 'fecal_egg', 'custom'
    ));

ALTER TABLE identity.users
    DROP COLUMN IF EXISTS payout_account_holder,
    DROP COLUMN IF EXISTS payout_bic,
    DROP COLUMN IF EXISTS payout_iban;

DROP TABLE IF EXISTS billing.commission_settings;

ALTER TABLE billing.commission_tiers DROP CONSTRAINT IF EXISTS commission_tiers_rate_bps_check;
ALTER TABLE billing.commission_tiers
    ADD CONSTRAINT commission_tiers_rate_bps_check CHECK (rate_bps >= 0 AND rate_bps <= 1500);
