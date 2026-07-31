-- Durée d'enregistrement audio du CR (secondes). Conservée après purge RGPD de l'objet audio.
ALTER TABLE visits.visit_reports
  ADD COLUMN IF NOT EXISTS audio_duration_sec INT NULL;
