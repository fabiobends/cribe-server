-- Convert date_published from BIGINT to TEXT with ISO 8601 format
ALTER TABLE episodes ALTER COLUMN date_published TYPE TEXT
USING to_char(to_timestamp(date_published), 'YYYY-MM-DD"T"HH24:MI:SS"Z"');
