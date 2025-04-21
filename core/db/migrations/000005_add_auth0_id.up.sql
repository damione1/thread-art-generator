-- Add auth0_id field to users table
ALTER TABLE users
ADD COLUMN auth0_id VARCHAR(255) UNIQUE;

-- Update the DB model with SQLBoiler
-- This doesn't actually do anything in SQL, but serves as documentation
-- You'll need to run: go generate ./...
