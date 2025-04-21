-- Make auth0_id NOT NULL (first ensure there are no NULL values)
UPDATE users SET auth0_id = id::text WHERE auth0_id IS NULL;
ALTER TABLE users
  ALTER COLUMN auth0_id SET NOT NULL;

-- Remove password column
ALTER TABLE users
  DROP COLUMN password;

-- Drop sessions table
DROP TABLE sessions;

-- Drop password_resets table
DROP TABLE password_resets;

-- Update the DB model with SQLBoiler
-- This doesn't actually do anything in SQL, but serves as documentation
-- You'll need to run: go generate ./...
