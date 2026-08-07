-- Révocation des JWT : le refresh compare la version portée par le token à
-- celle de l'utilisateur. Incrémentée au reset de mot de passe et au logout.
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 1;
