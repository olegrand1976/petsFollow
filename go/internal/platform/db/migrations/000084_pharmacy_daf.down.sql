DROP INDEX IF EXISTS pharmacy.idx_pharmacy_daf_items_daf;
ALTER TABLE pharmacy.stock_movements DROP CONSTRAINT IF EXISTS stock_movements_daf_item_id_fkey;
DROP TABLE IF EXISTS pharmacy.daf_items;
DROP INDEX IF EXISTS pharmacy.idx_pharmacy_daf_practice_status;
DROP INDEX IF EXISTS pharmacy.idx_pharmacy_daf_number_unique;
DROP TABLE IF EXISTS pharmacy.daf_documents;
DROP TABLE IF EXISTS pharmacy.daf_sequences;
