-- Drop unused identity column leftover. Users are Postgres UUIDs.
ALTER TABLE users DROP COLUMN IF EXISTS firebase_uid;
