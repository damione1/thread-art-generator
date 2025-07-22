-- Migration 000010: firebase_migration (up)

-- Add firebase_uid column to users table
ALTER TABLE users ADD COLUMN firebase_uid TEXT UNIQUE;

-- Create index on firebase_uid for performance
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);

-- Note: We're keeping auth0_id column temporarily during migration
-- It will be removed in a future migration after data is migrated
