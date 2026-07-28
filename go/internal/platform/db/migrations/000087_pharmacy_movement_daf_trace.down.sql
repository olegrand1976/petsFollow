DROP INDEX IF EXISTS pharmacy.idx_pharmacy_movements_daf;
ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS pharmacy_movements_daf_trace;
ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS stock_movements_daf_item_id_fkey;
ALTER TABLE pharmacy.stock_movements
  ADD CONSTRAINT stock_movements_daf_item_id_fkey
  FOREIGN KEY (daf_item_id) REFERENCES pharmacy.daf_items(id) ON DELETE SET NULL;
ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS stock_movements_daf_id_fkey;
ALTER TABLE pharmacy.stock_movements DROP COLUMN IF EXISTS daf_id;
