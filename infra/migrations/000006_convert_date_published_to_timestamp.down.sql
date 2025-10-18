-- Revert date_published from TEXT back to BIGINT
ALTER TABLE episodes ALTER COLUMN date_published TYPE BIGINT USING EXTRACT(EPOCH FROM date_published::TIMESTAMP)::BIGINT;
