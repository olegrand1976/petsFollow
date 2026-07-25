-- Rétention RGPD : trace de dernière connexion pour la purge « 3 ans d'inactivité ».
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;

-- Comptes existants : considérer maintenant comme point de départ (pas de purge rétroactive à l'aveugle).
UPDATE identity.users SET last_login_at = NOW() WHERE last_login_at IS NULL;
