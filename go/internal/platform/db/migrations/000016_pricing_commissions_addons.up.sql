-- Commission settings (commercial flat rate), payout profile, care+ / horse privileges.

-- 1. Raise vet tier cap (admin may set up to 50%).
ALTER TABLE billing.commission_tiers DROP CONSTRAINT IF EXISTS commission_tiers_rate_bps_check;
ALTER TABLE billing.commission_tiers
    ADD CONSTRAINT commission_tiers_rate_bps_check CHECK (rate_bps >= 0 AND rate_bps <= 5000);

-- Align progressive ladder: 5% → 8% → 10% → 12% (max).
UPDATE billing.commission_tiers SET rate_bps = 1000 WHERE min_clients = 15 AND rate_bps = 1200;
UPDATE billing.commission_tiers SET rate_bps = 1200 WHERE min_clients = 40 AND rate_bps = 1500;

-- 2. Commercial rate settings (single-row).
CREATE TABLE IF NOT EXISTS billing.commission_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    commercial_rate_bps INT NOT NULL DEFAULT 1200
        CHECK (commercial_rate_bps >= 0 AND commercial_rate_bps <= 5000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO billing.commission_settings (id, commercial_rate_bps)
VALUES (1, 1200)
ON CONFLICT (id) DO NOTHING;

-- 3. Payout bank details on commercial users.
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS payout_iban TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payout_bic TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payout_account_holder TEXT NOT NULL DEFAULT '';

-- 4. Care+ reminder fields + medication type.
ALTER TABLE care.reminders DROP CONSTRAINT IF EXISTS reminders_type_check;
ALTER TABLE care.reminders
    ADD CONSTRAINT reminders_type_check CHECK (type IN (
        'vaccination', 'deworming', 'vet_check', 'dental',
        'farrier', 'fecal_egg', 'custom', 'medication'
    ));
ALTER TABLE care.reminders ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';
ALTER TABLE care.reminders ADD COLUMN IF NOT EXISTS recurrence_days INT;

-- 5. Horse pack: professional contacts + competitions.
CREATE TABLE IF NOT EXISTS care.professional_contacts (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT '',
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_care_contacts_pet ON care.professional_contacts(pet_id);

CREATE TABLE IF NOT EXISTS care.competitions (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    event_date DATE NOT NULL,
    title TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    discipline TEXT NOT NULL DEFAULT '',
    result TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_care_competitions_pet ON care.competitions(pet_id);

-- 6. Grants for runtime role when present.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON billing.commission_settings TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON care.professional_contacts TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON care.competitions TO petsfollow_app;
  END IF;
END $$;
