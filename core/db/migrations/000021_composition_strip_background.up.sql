-- Migration 000021: composition_strip_background (up)
ALTER TABLE compositions
    ADD COLUMN strip_background BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN compositions.strip_background IS 'If true, the worker runs rembg and composites the subject on white before solving';
