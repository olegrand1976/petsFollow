DROP TABLE IF EXISTS identity.email_verification_tokens;

ALTER TABLE identity.users DROP COLUMN IF EXISTS email_verified_at;

ALTER TABLE practice.practices
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS contact_email,
    DROP COLUMN IF EXISTS address_line1,
    DROP COLUMN IF EXISTS address_line2,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS postal_code,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS profile_completed_at;
