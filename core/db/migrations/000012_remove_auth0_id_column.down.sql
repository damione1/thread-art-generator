-- Migration 000012: remove_auth0_id_column (down)

-- Re-add the auth0_id column to users table
ALTER TABLE users ADD COLUMN auth0_id VARCHAR(255) NOT NULL DEFAULT '';

-- Update existing records to have a dummy auth0_id value (since it's NOT NULL)
UPDATE users SET auth0_id = 'migrated_' || id WHERE auth0_id = '';

-- Re-add the unique constraint on auth0_id
ALTER TABLE users ADD CONSTRAINT unique_auth0_id UNIQUE (auth0_id);
