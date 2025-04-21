-- Migration 000009: spring_clean (up)
ALTER TABLE users ADD CONSTRAINT unique_auth0_id UNIQUE (auth0_id);

ALTER TABLE users
ALTER COLUMN email
DROP NOT NULL;

ALTER TABLE users
ALTER COLUMN email
SET DEFAULT NULL;
