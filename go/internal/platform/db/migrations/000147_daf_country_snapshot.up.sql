-- Sous quel régime réglementaire un DAF a-t-il été émis ?
-- Aligné sur le précédent 000091_prescriptions.up.sql (country_code CHAR(2) DEFAULT 'BE').
ALTER TABLE pharmacy.daf_documents
    ADD COLUMN IF NOT EXISTS country_code CHAR(2) NOT NULL DEFAULT 'BE';
