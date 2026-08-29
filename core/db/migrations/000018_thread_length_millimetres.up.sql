-- thread_length was stored in metres (the generator divided by 1000).
-- The UI already treats the column as millimetres, so existing rows
-- displayed as 1–2 m. Convert surviving metre values to millimetres.
UPDATE compositions
SET thread_length = thread_length * 1000
WHERE thread_length IS NOT NULL
  AND thread_length > 0
  AND thread_length < 100000;

COMMENT ON COLUMN compositions.thread_length IS 'Thread length in millimetres';
