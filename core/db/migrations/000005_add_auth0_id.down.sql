-- Remove auth0_id field from users table
ALTER TABLE users
DROP COLUMN auth0_id;
