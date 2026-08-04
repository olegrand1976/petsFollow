-- Webhooks Telnyx : rapports de livraison (DLR) sur les SMS sortants
-- et SMS entrants (dont les STOP, qui coupent le canal).

-- DLR : statut rapporté par l'opérateur, distinct de notre statut d'envoi.
ALTER TABLE notifications.sms_log
    ADD COLUMN IF NOT EXISTS delivery_status TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS delivery_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ NULL;

-- Clé de rapprochement webhook → ligne de log.
CREATE INDEX IF NOT EXISTS sms_log_provider_message_idx
    ON notifications.sms_log (provider_message_id)
    WHERE provider_message_id <> '';

-- SMS entrants. user_id est NULL quand le numéro ne correspond à aucun compte
-- (le message est conservé pour le support, jamais rattaché de force).
CREATE TABLE IF NOT EXISTS notifications.sms_inbound (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES identity.users(id) ON DELETE SET NULL,
    from_phone TEXT NOT NULL,
    to_phone TEXT NOT NULL DEFAULT '',
    -- body tronqué : nécessaire au support (« j'ai répondu STOP »), jamais exploité au-delà.
    body TEXT NOT NULL DEFAULT '',
    command TEXT NOT NULL DEFAULT 'other' CHECK (command IN ('stop','start','other')),
    provider_message_id TEXT NOT NULL DEFAULT '',
    via_failover BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotence : Telnyx peut relivrer le même événement (retry, puis failover).
CREATE UNIQUE INDEX IF NOT EXISTS sms_inbound_provider_message_once
    ON notifications.sms_inbound (provider_message_id)
    WHERE provider_message_id <> '';

CREATE INDEX IF NOT EXISTS sms_inbound_user_created_idx
    ON notifications.sms_inbound (user_id, created_at);
