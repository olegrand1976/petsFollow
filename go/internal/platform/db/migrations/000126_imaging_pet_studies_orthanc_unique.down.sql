ALTER TABLE imaging.pet_studies
  DROP CONSTRAINT IF EXISTS pet_studies_orthanc_study_id_key;

ALTER TABLE imaging.pet_studies
  ADD CONSTRAINT pet_studies_practice_id_orthanc_study_id_key UNIQUE (practice_id, orthanc_study_id);
