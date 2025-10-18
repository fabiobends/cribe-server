CREATE TABLE IF NOT EXISTS episodes (
    id SERIAL PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    podcast_id INTEGER NOT NULL REFERENCES podcasts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    audio_url TEXT NOT NULL,
    image_url TEXT,
    date_published BIGINT NOT NULL,
    duration INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on external_id for fast lookups
CREATE INDEX IF NOT EXISTS idx_episodes_external_id ON episodes(external_id);

-- Create index on podcast_id for fast lookups of episodes by podcast
CREATE INDEX IF NOT EXISTS idx_episodes_podcast_id ON episodes(podcast_id);
