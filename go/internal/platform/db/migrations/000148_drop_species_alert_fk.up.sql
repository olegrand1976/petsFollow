-- 000146 a été poussée avec une FK heartrate.species_alert_deltas(species)
-- → pets.species(code). Elle introduit une dépendance de verrous entre les
-- schémas heartrate et pets qui fait deadlocker le seed (écritures concurrentes
-- par profil), d'où son retrait de 000146. Mais golang-migrate ne rejoue pas une
-- version déjà appliquée : les bases qui ont reçu 000146 avec la FK (staging,
-- dev locaux) la conservent. Cette migration les fait converger.
--
-- L'intégrité reste portée par le CRUD admin des espèces, qui n'autorise que la
-- désactivation — jamais la suppression.
ALTER TABLE heartrate.species_alert_deltas
    DROP CONSTRAINT IF EXISTS species_alert_deltas_species_fkey;

-- Idempotent : sur une base fraîche (000146 sans FK) le DROP ne fait rien.
