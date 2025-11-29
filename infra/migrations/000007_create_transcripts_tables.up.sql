-- Create enum for transcript status
CREATE TYPE transcript_status AS ENUM ('processing', 'complete', 'failed');

-- Create transcripts table
CREATE TABLE IF NOT EXISTS transcripts (
    id SERIAL PRIMARY KEY,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    status transcript_status NOT NULL DEFAULT 'processing',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(episode_id)
);

-- Create transcript_chunks table
CREATE TABLE IF NOT EXISTS transcript_chunks (
    id SERIAL PRIMARY KEY,
    transcript_id INTEGER NOT NULL REFERENCES transcripts(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    speaker_index INTEGER NOT NULL,
    start_time FLOAT NOT NULL,
    end_time FLOAT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(transcript_id, position)
);

-- Create transcript_speakers table
CREATE TABLE IF NOT EXISTS transcript_speakers (
    id SERIAL PRIMARY KEY,
    transcript_id INTEGER NOT NULL REFERENCES transcripts(id) ON DELETE CASCADE,
    speaker_index INTEGER NOT NULL,
    speaker_name VARCHAR(255) NOT NULL DEFAULT 'Speaker',
    inferred_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(transcript_id, speaker_index)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_transcripts_episode_id ON transcripts(episode_id);
CREATE INDEX IF NOT EXISTS idx_chunks_transcript_id ON transcript_chunks(transcript_id);
CREATE INDEX IF NOT EXISTS idx_speakers_transcript_id ON transcript_speakers(transcript_id);
