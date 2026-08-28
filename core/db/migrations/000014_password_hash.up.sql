-- Migration 000014: password_hash (up)
-- Cookie email/password login. Column was dropped in 000006 (Auth0).
-- Nullable so Firebase dual-run users keep working until Phase E is fully cut.

ALTER TABLE users
  ADD COLUMN password_hash TEXT;

COMMENT ON COLUMN users.password_hash IS 'bcrypt hash for SCS cookie login. NULL for federated-only users.';
