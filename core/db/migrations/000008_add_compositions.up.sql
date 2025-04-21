-- Create enum type for composition status
CREATE TYPE composition_status_enum AS ENUM (
    'PENDING', -- Composition created but waiting to be processed
    'PROCESSING', -- Composition is currently being processed
    'COMPLETE', -- Composition has been successfully processed
    'FAILED' -- Composition processing failed
);

-- Create compositions table with all required fields
CREATE TABLE
    compositions (
        id UUID DEFAULT uuid_generate_v1mc () PRIMARY KEY,
        art_id UUID NOT NULL REFERENCES arts (id) ON DELETE CASCADE,
        status composition_status_enum NOT NULL DEFAULT 'PENDING',
        -- Configuration settings from ThreadGenerator
        nails_quantity INTEGER NOT NULL DEFAULT 300,
        img_size INTEGER NOT NULL DEFAULT 800,
        max_paths INTEGER NOT NULL DEFAULT 10000,
        starting_nail INTEGER NOT NULL DEFAULT 0,
        minimum_difference INTEGER NOT NULL DEFAULT 10,
        brightness_factor INTEGER NOT NULL DEFAULT 50,
        image_contrast FLOAT NOT NULL DEFAULT 40.0,
        physical_radius FLOAT NOT NULL DEFAULT 609.6, -- 24 inches in mm
        -- Result fields
        preview_url TEXT,
        gcode_url TEXT,
        pathlist_url TEXT,
        thread_length INTEGER, -- in meters
        total_lines INTEGER,
        error_message TEXT,
        -- Standard timestamps
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT NOW () NOT NULL,
            updated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT NOW () NOT NULL
    );

-- Create indexes for efficient querying
CREATE INDEX idx_compositions_art_id ON compositions (art_id);

CREATE INDEX idx_compositions_status ON compositions (status);

-- Add comments for documentation
COMMENT ON TABLE compositions IS 'Thread art composition configurations and results';

COMMENT ON COLUMN compositions.status IS 'Current status of the composition processing';

COMMENT ON COLUMN compositions.nails_quantity IS 'Number of nails to use in the circle';

COMMENT ON COLUMN compositions.img_size IS 'Image size in pixels';

COMMENT ON COLUMN compositions.max_paths IS 'Maximum number of paths to generate';

COMMENT ON COLUMN compositions.starting_nail IS 'Starting nail position';

COMMENT ON COLUMN compositions.minimum_difference IS 'Minimum difference between connected nails';

COMMENT ON COLUMN compositions.brightness_factor IS 'Brightness factor for thread lines';

COMMENT ON COLUMN compositions.image_contrast IS 'Image contrast adjustment value';

COMMENT ON COLUMN compositions.physical_radius IS 'Physical radius of the final artwork in mm';

COMMENT ON COLUMN compositions.preview_url IS 'URL to the preview image of the composition result';

COMMENT ON COLUMN compositions.gcode_url IS 'URL to download the GCode file';

COMMENT ON COLUMN compositions.pathlist_url IS 'URL to download the paths list file';

COMMENT ON COLUMN compositions.thread_length IS 'Thread length in meters';

COMMENT ON COLUMN compositions.total_lines IS 'Total number of thread lines';

COMMENT ON COLUMN compositions.error_message IS 'Error message if processing failed';
