DELETE FROM heartrate.species_alert_deltas WHERE species NOT IN ('dog', 'cat', 'horse');

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'species_alert_deltas_species_check'
  ) THEN
    ALTER TABLE heartrate.species_alert_deltas
      ADD CONSTRAINT species_alert_deltas_species_check
      CHECK (species IN ('dog', 'cat', 'horse'));
  END IF;
END $$;

DROP INDEX IF EXISTS pets.idx_species_rules_country;
DROP TABLE IF EXISTS pets.species_country_rules;
DROP TABLE IF EXISTS pets.species;
