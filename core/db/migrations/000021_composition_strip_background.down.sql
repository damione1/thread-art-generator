-- Migration 000021: composition_strip_background (down)
ALTER TABLE compositions DROP COLUMN strip_background;
