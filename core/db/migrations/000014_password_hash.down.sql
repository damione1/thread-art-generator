-- Migration 000014: password_hash (down)

ALTER TABLE users
  DROP COLUMN IF EXISTS password_hash;
