-- Migration 000017: security_hardening (down)

ALTER TABLE arts ALTER COLUMN id SET DEFAULT uuid_generate_v1mc();

ALTER TABLE email_tokens DROP CONSTRAINT IF EXISTS email_tokens_purpose_check;
ALTER TABLE email_tokens
  ADD CONSTRAINT email_tokens_purpose_check CHECK (purpose IN ('verify', 'reset'));

ALTER TABLE email_tokens DROP COLUMN IF EXISTS payload;

ALTER TABLE users DROP COLUMN IF EXISTS session_version;
