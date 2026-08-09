-- Type de client facturé (Billit) : particulier = ni TVA ni Peppol, envoi email.
-- Défaut vide plutôt que 'individual' : une fiche antérieure resterait sinon
-- marquée particulier alors qu'elle porte un n° de TVA, et son avoir partirait
-- hors Peppol sans TVA. Vide = déduction par identifiant fiscal (comportement actuel).
ALTER TABLE identity.users
  ADD COLUMN IF NOT EXISTS billing_customer_kind TEXT NOT NULL DEFAULT '';

ALTER TABLE identity.users
  DROP CONSTRAINT IF EXISTS users_billing_customer_kind_check;

ALTER TABLE identity.users
  ADD CONSTRAINT users_billing_customer_kind_check
  CHECK (billing_customer_kind IN ('', 'individual', 'business'));
