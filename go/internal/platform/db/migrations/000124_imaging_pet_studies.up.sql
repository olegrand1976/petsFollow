CREATE SCHEMA IF NOT EXISTS imaging;

CREATE TABLE IF NOT EXISTS imaging.pet_studies (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    orthanc_study_id TEXT NOT NULL,
    study_instance_uid TEXT NOT NULL DEFAULT '',
    orthanc_series_id TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    modality TEXT NOT NULL DEFAULT '',
    uploaded_by_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (practice_id, orthanc_study_id)
);

CREATE INDEX IF NOT EXISTS pet_studies_pet_id_idx ON imaging.pet_studies (pet_id);
CREATE INDEX IF NOT EXISTS pet_studies_practice_id_idx ON imaging.pet_studies (practice_id);
