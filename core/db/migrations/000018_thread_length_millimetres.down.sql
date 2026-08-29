-- Reverse metre → millimetre conversion for values that look like millimetres.
UPDATE compositions
SET thread_length = thread_length / 1000
WHERE thread_length IS NOT NULL
  AND thread_length >= 100000;

COMMENT ON COLUMN compositions.thread_length IS 'Thread length in meters';
