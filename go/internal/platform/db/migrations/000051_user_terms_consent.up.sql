-- RGPD art. 7 (accountability) : preuve horodatée de l'acceptation CGU/privacy au register.
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ;
