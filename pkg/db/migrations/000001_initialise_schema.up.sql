CREATE EXTENSION "uuid-ossp";

-- Media table
CREATE TABLE medias (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  original_filename VARCHAR(255) NOT NULL,
  original_url VARCHAR(255) NOT NULL,
  thumbnail_url VARCHAR(255) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User table
CREATE TABLE users (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  avatar_id uuid REFERENCES media(id),
  active BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  role VARCHAR(255) NOT NULL DEFAULT 'user'
);

CREATE INDEX users_email_idx ON users (email);

-- Arts table
CREATE TABLE arts (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  image_id uuid REFERENCES media(id),
  author_id uuid NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX arts_author_idx ON arts (author_id);

-- Generated variations
CREATE TABLE art_variations (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  art_id uuid REFERENCES arts(id),
  image_id uuid REFERENCES media(id),
  author_id uuid NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  deleted_at TIMESTAMP WITH TIME ZONE
);

-- Session table
CREATE TABLE sessions (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token varchar NOT NULL,
  user_agent varchar NOT NULL,
  client_ip varchar NOT NULL,
  is_blocked boolean NOT NULL DEFAULT FALSE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_refresh_token_idx ON sessions (refresh_token);

-- Password reset table
CREATE TABLE password_resets (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reset_token VARCHAR(255) NOT NULL,
  expiration TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '1 day',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
