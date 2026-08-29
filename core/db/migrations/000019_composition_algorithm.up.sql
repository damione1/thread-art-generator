-- Migration 000019: composition_algorithm (up)
ALTER TABLE compositions
    ADD COLUMN algorithm INTEGER NOT NULL DEFAULT 1;

COMMENT ON COLUMN compositions.algorithm IS '1=Vrellis consecutive darkness, 2=L2 residual (StringArt/Birsak)';
