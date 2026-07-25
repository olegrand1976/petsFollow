ALTER TABLE sales.pitch_simulations
    ADD COLUMN IF NOT EXISTS feedback_skipped BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_pitch_sims_feedback_skipped_day
    ON sales.pitch_simulations (user_id, ended_at)
    WHERE feedback_skipped = true;
