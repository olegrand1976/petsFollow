-- Continuous improvement: mark finalized CRs as quality reference consultations.
ALTER TABLE visits.visit_reports
  ADD COLUMN IF NOT EXISTS is_reference boolean NOT NULL DEFAULT false;

-- Partial index for listing reference CRs (joined to visits.practice_id in queries).
CREATE INDEX IF NOT EXISTS visit_reports_is_reference_idx
  ON visits.visit_reports (is_reference)
  WHERE is_reference = true;
