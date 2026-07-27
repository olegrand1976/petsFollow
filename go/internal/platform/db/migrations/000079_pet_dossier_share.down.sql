DROP TABLE IF EXISTS pets.dossier_share_tokens;

ALTER TABLE identity.users
  DROP COLUMN IF EXISTS contact_phone;
