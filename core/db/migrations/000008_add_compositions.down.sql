-- Remove indexes
DROP INDEX IF EXISTS idx_compositions_art_id;

DROP INDEX IF EXISTS idx_compositions_status;

-- Drop compositions table
DROP TABLE IF EXISTS compositions;

-- Drop enum type
DROP TYPE IF EXISTS composition_status_enum;
