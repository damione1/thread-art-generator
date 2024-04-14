CREATE TYPE role_enum AS ENUM ('user', 'admin', 'super_admin');

ALTER TABLE users DROP COLUMN role;
ALTER TABLE users ADD COLUMN role role_enum NOT NULL DEFAULT 'user';
