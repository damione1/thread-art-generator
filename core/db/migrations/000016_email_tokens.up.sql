-- Migration 000016: email_tokens (up)
-- Opaque hashed tokens for account verification and password reset.

CREATE TABLE email_tokens (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose VARCHAR(32) NOT NULL,
  token_hash VARCHAR(64) NOT NULL,
  expiration TIMESTAMP WITH TIME ZONE NOT NULL,
  used_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  CONSTRAINT email_tokens_purpose_check CHECK (purpose IN ('verify', 'reset'))
);

CREATE UNIQUE INDEX email_tokens_token_hash_idx ON email_tokens (token_hash);
CREATE INDEX email_tokens_user_purpose_idx ON email_tokens (user_id, purpose);
