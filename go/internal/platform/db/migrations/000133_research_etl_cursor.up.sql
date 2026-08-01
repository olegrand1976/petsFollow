-- Composite ETL cursor (ts + id) to avoid skipping events at LIMIT boundaries.
ALTER TABLE research.etl_watermarks
    ADD COLUMN IF NOT EXISTS watermark_id TEXT NOT NULL DEFAULT '';
