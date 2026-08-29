-- Migration 000019: composition_algorithm (down)
ALTER TABLE compositions DROP COLUMN algorithm;
