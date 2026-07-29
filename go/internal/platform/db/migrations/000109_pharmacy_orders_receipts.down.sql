ALTER TABLE pharmacy.stock_movements DROP COLUMN IF EXISTS delivery_note_id;
DROP TABLE IF EXISTS pharmacy.delivery_note_items;
DROP TABLE IF EXISTS pharmacy.delivery_notes;
DROP TABLE IF EXISTS pharmacy.purchase_order_items;
DROP TABLE IF EXISTS pharmacy.purchase_orders;
DROP TABLE IF EXISTS pharmacy.suppliers;
