CREATE TABLE IF NOT EXISTS sales.pitch_sim_manager_notes (
    simulation_id UUID PRIMARY KEY REFERENCES sales.pitch_simulations(id) ON DELETE CASCADE,
    manager_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    note TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pitch_sim_manager_notes_manager
    ON sales.pitch_sim_manager_notes(manager_user_id);
