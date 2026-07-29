ALTER TABLE pharmacy.stock_movements DROP COLUMN IF EXISTS inventory_session_id;
DROP TABLE IF EXISTS pharmacy.inventory_lines;
DROP TABLE IF EXISTS pharmacy.inventory_sessions;
