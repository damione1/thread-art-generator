-- Migration 000012: remove_auth0_id_column (up)

-- Remove the unique constraint on auth0_id if it exists
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_auth0_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_auth0_id_key;

-- Remove the auth0_id column from users table
ALTER TABLE users DROP COLUMN IF EXISTS auth0_id;
