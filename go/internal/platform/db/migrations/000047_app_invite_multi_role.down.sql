DROP TABLE IF EXISTS practice.commercial_referrals;

-- Best-effort restore of vet-only table (data from app_invite_codes where role=vet).
CREATE TABLE IF NOT EXISTS practice.vet_app_invite_codes (
    vet_user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vet_app_invite_codes_code_unique UNIQUE (code)
);

INSERT INTO practice.vet_app_invite_codes (vet_user_id, practice_id, code, created_at)
SELECT user_id, practice_id, code, created_at
FROM practice.app_invite_codes
WHERE role = 'vet' AND practice_id IS NOT NULL
ON CONFLICT (vet_user_id) DO NOTHING;

DROP TABLE IF EXISTS practice.app_invite_codes;
