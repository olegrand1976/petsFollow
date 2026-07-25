-- Configurable commission rates per professional profile (care_pro specialties start at 0%).
CREATE TABLE IF NOT EXISTS billing.profile_commission_rates (
    profile_key TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    rate_bps INT NOT NULL DEFAULT 0 CHECK (rate_bps >= 0 AND rate_bps <= 10000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO billing.profile_commission_rates (profile_key, label, rate_bps) VALUES
    ('care_pro.vet_light', 'Care pro — Véto light', 0),
    ('care_pro.farrier', 'Care pro — Maréchal', 0),
    ('care_pro.physio', 'Care pro — Physio', 0),
    ('care_pro.behaviorist', 'Care pro — Comportementaliste', 0)
ON CONFLICT (profile_key) DO NOTHING;
