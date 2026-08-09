-- Tarif de l'acte par type de RDV : sert à pré-remplir la facture Billit en fin
-- de consultation (BIL-9). Le produit n'a pas de catalogue de prestations ; le
-- type de rendez-vous est le seul ancrage existant du prix d'une consultation.
-- 0 = non tarifé : aucune ligne n'est proposée, le véto saisit comme aujourd'hui.

ALTER TABLE practice.visit_types
    ADD COLUMN IF NOT EXISTS price_excl_cents INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS vat_percent NUMERIC(5,2) NOT NULL DEFAULT 21;

ALTER TABLE practice.visit_types
    DROP CONSTRAINT IF EXISTS visit_types_price_excl_cents_check;

ALTER TABLE practice.visit_types
    ADD CONSTRAINT visit_types_price_excl_cents_check
    CHECK (price_excl_cents >= 0 AND price_excl_cents <= 100000000);

ALTER TABLE practice.visit_types
    DROP CONSTRAINT IF EXISTS visit_types_vat_percent_check;

-- Mêmes taux que le catalogue TVA de la facturation (0 / 6 / 12 / 21 / 22).
ALTER TABLE practice.visit_types
    ADD CONSTRAINT visit_types_vat_percent_check
    CHECK (vat_percent >= 0 AND vat_percent <= 100);
