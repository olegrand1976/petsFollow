DROP TRIGGER IF EXISTS trg_stock_movements_immutable ON pharmacy.stock_movements;
DROP FUNCTION IF EXISTS pharmacy.enforce_stock_movements_immutable();
