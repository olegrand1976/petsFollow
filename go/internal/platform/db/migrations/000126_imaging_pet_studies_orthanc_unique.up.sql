-- Global uniqueness: one Orthanc study may bind to at most one practice (anti IDOR merge).
ALTER TABLE imaging.pet_studies
  DROP CONSTRAINT IF EXISTS pet_studies_practice_id_orthanc_study_id_key;

ALTER TABLE imaging.pet_studies
  ADD CONSTRAINT pet_studies_orthanc_study_id_key UNIQUE (orthanc_study_id);
