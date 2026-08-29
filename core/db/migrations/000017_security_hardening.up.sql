-- Migration 000017: security_hardening (up)
-- Session epoch, email-change token payload, UUID v4 art ids.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS session_version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE email_tokens
  ADD COLUMN IF NOT EXISTS payload TEXT NOT NULL DEFAULT '';

ALTER TABLE email_tokens DROP CONSTRAINT IF EXISTS email_tokens_purpose_check;
ALTER TABLE email_tokens
  ADD CONSTRAINT email_tokens_purpose_check CHECK (purpose IN ('verify', 'reset', 'email_change'));

ALTER TABLE arts ALTER COLUMN id SET DEFAULT uuid_generate_v4();
