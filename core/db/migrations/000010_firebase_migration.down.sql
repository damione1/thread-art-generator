-- Migration 000010: firebase_migration (down)

-- Remove index on firebase_uid
DROP INDEX IF EXISTS idx_users_firebase_uid;

-- Remove firebase_uid column from users table
ALTER TABLE users DROP COLUMN IF EXISTS firebase_uid;
