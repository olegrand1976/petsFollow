-- Traçabilité réglementaire : tout mouvement DAF / DAF cancel doit pointer le document + la ligne.
ALTER TABLE pharmacy.stock_movements
    ADD COLUMN IF NOT EXISTS daf_id UUID;

-- Backfill depuis les lignes DAF déjà liées.
UPDATE pharmacy.stock_movements m
SET daf_id = i.daf_id
FROM pharmacy.daf_items i
WHERE m.daf_item_id = i.id
  AND m.daf_id IS NULL;

-- Orphelins (sortie DAF sans lien) : reclasse en adjust pour conserver l'audit (pas de DELETE).
UPDATE pharmacy.stock_movements
SET reason = 'adjust',
    reason_detail = TRIM(BOTH ' ' FROM CONCAT(COALESCE(reason_detail, ''), ' orphan_pre_087_daf_trace')),
    daf_id = NULL,
    daf_item_id = NULL
WHERE reason IN ('daf', 'daf_cancel')
  AND (daf_item_id IS NULL OR daf_id IS NULL);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'stock_movements_daf_id_fkey'
  ) THEN
    ALTER TABLE pharmacy.stock_movements
      ADD CONSTRAINT stock_movements_daf_id_fkey
      FOREIGN KEY (daf_id) REFERENCES pharmacy.daf_documents(id) ON DELETE RESTRICT;
  END IF;
END $$;

-- Remplacer SET NULL par RESTRICT : un DELETE item ne doit pas casser le CHECK métier.
ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS stock_movements_daf_item_id_fkey;
ALTER TABLE pharmacy.stock_movements
  ADD CONSTRAINT stock_movements_daf_item_id_fkey
  FOREIGN KEY (daf_item_id) REFERENCES pharmacy.daf_items(id) ON DELETE RESTRICT;

ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS pharmacy_movements_daf_trace;
ALTER TABLE pharmacy.stock_movements
  ADD CONSTRAINT pharmacy_movements_daf_trace CHECK (
    (reason NOT IN ('daf', 'daf_cancel'))
    OR (daf_id IS NOT NULL AND daf_item_id IS NOT NULL)
  );

CREATE INDEX IF NOT EXISTS idx_pharmacy_movements_daf
    ON pharmacy.stock_movements (practice_id, daf_id)
    WHERE daf_id IS NOT NULL;
