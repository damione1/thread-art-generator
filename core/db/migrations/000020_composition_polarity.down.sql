-- Migration 000020: composition_polarity (down)
ALTER TABLE compositions DROP COLUMN polarity;
