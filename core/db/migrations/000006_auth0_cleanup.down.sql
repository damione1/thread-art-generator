-- Re-create password_resets table
CREATE TABLE
    password_resets (
        id UUID DEFAULT uuid_generate_v1mc () PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        reset_token VARCHAR(255) NOT NULL,
        expiration TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT NOW () + INTERVAL '1 day',
            created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT NOW ()
    );

-- Re-create sessions table
CREATE TABLE
    sessions (
        id UUID DEFAULT uuid_generate_v1mc () PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        refresh_token varchar NOT NULL,
        user_agent varchar NOT NULL,
        client_ip varchar NOT NULL,
        is_blocked boolean NOT NULL DEFAULT FALSE,
        expires_at timestamptz NOT NULL,
        created_at timestamptz NOT NULL DEFAULT now ()
    );

CREATE INDEX sessions_refresh_token_idx ON sessions (refresh_token);

-- Add password column back
ALTER TABLE users
ADD COLUMN password VARCHAR(255);

-- Make auth0_id nullable again
ALTER TABLE users
ALTER COLUMN auth0_id
DROP NOT NULL;
