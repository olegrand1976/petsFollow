DROP FUNCTION IF EXISTS pharmacy.rgpd_null_stock_movement_created_by(UUID);
DO $$ BEGIN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.stock_movements TO petsfollow_app;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;
