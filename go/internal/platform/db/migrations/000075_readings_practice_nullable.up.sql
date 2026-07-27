-- Allow heart-rate and weight readings before a vet/practice is linked.
ALTER TABLE heartrate.sessions
    ALTER COLUMN practice_id DROP NOT NULL;

ALTER TABLE pets.weight_readings
    ALTER COLUMN practice_id DROP NOT NULL;
