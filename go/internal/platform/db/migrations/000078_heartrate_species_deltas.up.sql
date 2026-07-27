CREATE TABLE IF NOT EXISTS heartrate.species_alert_deltas (
    species TEXT PRIMARY KEY CHECK (species IN ('dog', 'cat', 'horse')),
    delta_bpm INT NOT NULL CHECK (delta_bpm > 0 AND delta_bpm <= 200),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO heartrate.species_alert_deltas (species, delta_bpm)
VALUES
    ('dog', 30),
    ('cat', 30),
    ('horse', 30)
ON CONFLICT (species) DO NOTHING;
