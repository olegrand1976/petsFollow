-- Drop commercial_ramp SPIFF; keep commercial_mix only.

DELETE FROM billing.commercial_bonus_awards WHERE bonus_code = 'commercial_ramp';

ALTER TABLE billing.commercial_bonus_awards
    DROP CONSTRAINT IF EXISTS commercial_bonus_awards_bonus_code_check;

ALTER TABLE billing.commercial_bonus_awards
    ADD CONSTRAINT commercial_bonus_awards_bonus_code_check
    CHECK (bonus_code IN ('commercial_mix'));

ALTER TABLE billing.commercial_bonus_awards
    DROP CONSTRAINT IF EXISTS commercial_bonus_ramp_fields;

-- Pitch script: remove ramp talking point from money step.
UPDATE sales.pitch_scripts SET
    steps_json = jsonb_set(
        steps_json,
        '{2}',
        '{
          "id":"money",
          "title":"Commission véto",
          "talkingPoints":["Même plafond avec ou sans commercial","Steer triennial ~9,4 €","Activation pets payants = revenu"],
          "exampleLine":"Sur le triennial, jusqu’à environ 9,4 € pour vous."
        }'::jsonb
    ),
    updated_at = NOW()
WHERE id = 'a0000000-0000-4000-8000-000000000001'
  AND steps_json->2->>'id' = 'money';
