ALTER TABLE pets.weight_readings
    ALTER COLUMN practice_id SET NOT NULL;

ALTER TABLE heartrate.sessions
    ALTER COLUMN practice_id SET NOT NULL;
