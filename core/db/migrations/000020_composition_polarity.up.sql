-- Migration 000020: composition_polarity (up)
ALTER TABLE compositions
    ADD COLUMN polarity INTEGER NOT NULL DEFAULT 1;

COMMENT ON COLUMN compositions.polarity IS '1=dark thread on light canvas, 2=light thread on dark canvas';
