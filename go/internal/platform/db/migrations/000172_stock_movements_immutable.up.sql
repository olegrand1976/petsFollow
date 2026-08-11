-- V4.1: stock_movements append-only for petsfollow_app (INSERT/SELECT only).
-- RGPD pro anonymisation may null created_by via SECURITY DEFINER function below.
DO $$ BEGIN
    REVOKE UPDATE, DELETE ON pharmacy.stock_movements FROM petsfollow_app;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;

CREATE OR REPLACE FUNCTION pharmacy.rgpd_null_stock_movement_created_by(p_user_id UUID)
RETURNS INTEGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pharmacy, pg_temp
AS $$
DECLARE
  n INTEGER;
BEGIN
  UPDATE pharmacy.stock_movements
  SET created_by = NULL
  WHERE created_by = p_user_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  RETURN n;
END;
$$;

REVOKE ALL ON FUNCTION pharmacy.rgpd_null_stock_movement_created_by(UUID) FROM PUBLIC;
DO $$ BEGIN
    GRANT EXECUTE ON FUNCTION pharmacy.rgpd_null_stock_movement_created_by(UUID) TO petsfollow_app;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;
