-- Lot number must be non-empty after trim (regulatory batch identity).
UPDATE pharmacy.medication_batches
SET lot_number = TRIM(lot_number)
WHERE lot_number <> TRIM(lot_number);

-- Blank lots that only satisfied NOT NULL: placeholder for migrate, NOT a real AFMPS lot.
-- Ops: inventory any UNKNOWN-PRE-088 rows after migrate and re-label or waste them.
UPDATE pharmacy.medication_batches
SET lot_number = 'UNKNOWN-PRE-088'
WHERE TRIM(lot_number) = '';

ALTER TABLE pharmacy.medication_batches DROP CONSTRAINT IF EXISTS pharmacy_batches_lot_nonempty;
ALTER TABLE pharmacy.medication_batches
  ADD CONSTRAINT pharmacy_batches_lot_nonempty CHECK (length(trim(lot_number)) > 0);
