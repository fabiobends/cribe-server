CREATE TABLE IF NOT EXISTS podcasts (
    id SERIAL PRIMARY KEY,
    author_name TEXT NOT NULL,
    name TEXT NOT NULL,
    image_url TEXT,
    description TEXT,
    external_id TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on external_id for fast lookups
CREATE INDEX IF NOT EXISTS idx_podcasts_external_id ON podcasts(external_id);
