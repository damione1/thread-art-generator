CREATE EXTENSION "uuid-ossp";
-- Media table
CREATE TABLE media (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  original_filename VARCHAR(255) NOT NULL,
  original_url VARCHAR(255) NOT NULL,
  thumbnail_url VARCHAR(255) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Arts table
CREATE TABLE arts (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  image_id uuid REFERENCES media(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  deleted_at TIMESTAMP WITH TIME ZONE
);

-- Generated variations
CREATE TABLE art_variations (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  art_id uuid REFERENCES arts(id),
  image_id uuid REFERENCES media(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  deleted_at TIMESTAMP WITH TIME ZONE
);
