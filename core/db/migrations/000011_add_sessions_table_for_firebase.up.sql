-- Migration 000011: add_sessions_table_for_firebase (up)

-- Create sessions table for alexedwards/scs session management
-- This table is required for the PostgreSQL session store used with Firebase authentication
CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

-- Create index on expiry for efficient cleanup of expired sessions
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- Add comment to document the table's purpose
COMMENT ON TABLE sessions IS 'Session storage for Firebase authentication using alexedwards/scs';
COMMENT ON COLUMN sessions.token IS 'Unique session token/ID';
COMMENT ON COLUMN sessions.data IS 'Serialized session data in binary format';
COMMENT ON COLUMN sessions.expiry IS 'Session expiration timestamp with timezone';
