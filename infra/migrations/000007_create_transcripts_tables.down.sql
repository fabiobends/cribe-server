-- Drop indexes first
DROP INDEX IF EXISTS idx_speakers_transcript_id;
DROP INDEX IF EXISTS idx_chunks_transcript_id;
DROP INDEX IF EXISTS idx_transcripts_episode_id;

-- Drop tables
DROP TABLE IF EXISTS transcript_speakers;
DROP TABLE IF EXISTS transcript_chunks;
DROP TABLE IF EXISTS transcripts;

-- Drop enum type
DROP TYPE IF EXISTS transcript_status;
