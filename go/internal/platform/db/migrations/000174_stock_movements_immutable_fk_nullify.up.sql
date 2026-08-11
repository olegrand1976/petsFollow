-- Allow FK ON DELETE SET NULL (created_by, delivery_note_id, inventory_session_id)
-- while still forbidding mutation of business columns. Future ALTER on stock_movements:
-- extend the "business columns" equality checks below.
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
    -- Business / audit columns: never change after insert.
    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.practice_id IS DISTINCT FROM OLD.practice_id
       OR NEW.batch_id IS DISTINCT FROM OLD.batch_id
       OR NEW.delta IS DISTINCT FROM OLD.delta
       OR NEW.reason IS DISTINCT FROM OLD.reason
       OR NEW.reason_detail IS DISTINCT FROM OLD.reason_detail
       OR NEW.daf_id IS DISTINCT FROM OLD.daf_id
       OR NEW.daf_item_id IS DISTINCT FROM OLD.daf_item_id
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
    THEN
      RAISE EXCEPTION 'stock_movements_immutable'
        USING ERRCODE = '42501';
    END IF;
    -- Soft FK nullifies only (RGPD + ON DELETE SET NULL). Re-assign to another UUID forbidden.
    IF NEW.created_by IS DISTINCT FROM OLD.created_by AND NEW.created_by IS NOT NULL THEN
      RAISE EXCEPTION 'stock_movements_immutable'
        USING ERRCODE = '42501';
    END IF;
    IF NEW.delivery_note_id IS DISTINCT FROM OLD.delivery_note_id AND NEW.delivery_note_id IS NOT NULL THEN
      RAISE EXCEPTION 'stock_movements_immutable'
        USING ERRCODE = '42501';
    END IF;
    IF NEW.inventory_session_id IS DISTINCT FROM OLD.inventory_session_id AND NEW.inventory_session_id IS NOT NULL THEN
      RAISE EXCEPTION 'stock_movements_immutable'
        USING ERRCODE = '42501';
    END IF;
    RETURN NEW;
  END IF;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION pharmacy.enforce_stock_movements_immutable() IS
  'Append-only stock_movements: block DELETE and business UPDATEs; allow nullify of created_by / delivery_note_id / inventory_session_id (FK ON DELETE SET NULL + RGPD). Hard DELETE of a practice still cascades here — ops must soft-delete or drop trigger for wipe.';
