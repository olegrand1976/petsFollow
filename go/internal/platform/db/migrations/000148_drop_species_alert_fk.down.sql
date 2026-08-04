-- Rétablir la FK reproduit le deadlock de seed décrit dans 000146 : ce down est
-- volontairement un no-op. Le rollback fonctionnel du catalogue d'espèces est
-- porté par 000146.down.sql (restauration du CHECK dog/cat/horse).
SELECT 1;
