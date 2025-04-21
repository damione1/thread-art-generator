-- Remove index
DROP INDEX IF EXISTS idx_arts_status;

-- Remove status column
ALTER TABLE arts
DROP COLUMN IF EXISTS status;

-- Drop enum type
DROP TYPE IF EXISTS art_status_enum;
