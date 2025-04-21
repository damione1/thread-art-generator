-- Create enum type for art status
CREATE TYPE art_status_enum AS ENUM (
    'PENDING_IMAGE', -- Art is created but image is pending upload
    'PROCESSING', -- Image is uploaded and being processed
    'COMPLETE', -- Art is complete with processed image
    'FAILED', -- Processing failed
    'ARCHIVED' -- Art is archived/hidden but not deleted
);

-- Add status column to arts table
ALTER TABLE arts
ADD COLUMN status art_status_enum NOT NULL DEFAULT 'PENDING_IMAGE';

-- Update existing records to COMPLETE if they have an image_id
UPDATE arts
SET
    status = 'COMPLETE'
WHERE
    image_id IS NOT NULL;

-- Add index for status filtering
CREATE INDEX idx_arts_status ON arts (status);

-- Add comment
COMMENT ON COLUMN arts.status IS 'Current status of the art resource';
