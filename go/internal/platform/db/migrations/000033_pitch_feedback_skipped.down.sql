DROP INDEX IF EXISTS sales.idx_pitch_sims_feedback_skipped_day;
ALTER TABLE sales.pitch_simulations DROP COLUMN IF EXISTS feedback_skipped;
