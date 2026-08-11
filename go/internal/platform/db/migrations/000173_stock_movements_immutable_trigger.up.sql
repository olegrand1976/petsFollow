-- Belt-and-suspenders: block UPDATE/DELETE even when connected as table owner (local/dev).
-- Allow only nullifying created_by (RGPD path used by rgpd_null_stock_movement_created_by).
CREATE OR REPLACE FUNCTION pharmacy.enforce_stock_movements_immutable()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'stock_movements_immutable'
      USING ERRCODE = '42501';
  END IF;
  IF TG_OP = 'UPDATE' THEN
    IF NEW.created_by IS NULL
       AND OLD.created_by IS NOT NULL
       AND NEW.id IS NOT DISTINCT FROM OLD.id
       AND NEW.practice_id IS NOT DISTINCT FROM OLD.practice_id
       AND NEW.batch_id IS NOT DISTINCT FROM OLD.batch_id
       AND NEW.delta IS NOT DISTINCT FROM OLD.delta
       AND NEW.reason IS NOT DISTINCT FROM OLD.reason
       AND NEW.reason_detail IS NOT DISTINCT FROM OLD.reason_detail
       AND NEW.daf_id IS NOT DISTINCT FROM OLD.daf_id
       AND NEW.daf_item_id IS NOT DISTINCT FROM OLD.daf_item_id
       AND NEW.delivery_note_id IS NOT DISTINCT FROM OLD.delivery_note_id
       AND NEW.inventory_session_id IS NOT DISTINCT FROM OLD.inventory_session_id
       AND NEW.created_at IS NOT DISTINCT FROM OLD.created_at
    THEN
      RETURN NEW;
    END IF;
    RAISE EXCEPTION 'stock_movements_immutable'
      USING ERRCODE = '42501';
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_stock_movements_immutable ON pharmacy.stock_movements;
CREATE TRIGGER trg_stock_movements_immutable
  BEFORE UPDATE OR DELETE ON pharmacy.stock_movements
  FOR EACH ROW
  EXECUTE PROCEDURE pharmacy.enforce_stock_movements_immutable();
