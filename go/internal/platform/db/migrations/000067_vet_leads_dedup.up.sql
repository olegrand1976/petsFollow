-- Dédup leads « nouveau véto » : un seul lead par (client, email).
DELETE FROM practice.vet_leads a
    USING practice.vet_leads b
WHERE a.client_user_id = b.client_user_id
  AND a.email = b.email
  AND a.ctid < b.ctid;

CREATE UNIQUE INDEX IF NOT EXISTS idx_vet_leads_client_email
    ON practice.vet_leads (client_user_id, email);
