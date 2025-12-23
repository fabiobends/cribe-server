-- Create enum types
CREATE TYPE question_type AS ENUM ('multiple_choice', 'true_false', 'open_ended');
CREATE TYPE session_status AS ENUM ('in_progress', 'completed', 'abandoned');

-- Questions table (episode-level, shared across all users)
CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    type question_type NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(episode_id, position)
);

-- Question options (only for multiple_choice and true_false)
CREATE TABLE IF NOT EXISTS question_options (
    id SERIAL PRIMARY KEY,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    position INTEGER NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(question_id, position)
);

-- User quiz sessions (tracks each user's attempt)
CREATE TABLE IF NOT EXISTS user_quiz_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    status session_status NOT NULL DEFAULT 'in_progress',
    total_questions INTEGER NOT NULL,
    answered_questions INTEGER NOT NULL DEFAULT 0,
    correct_answers INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User answers (user-specific responses with feedback)
CREATE TABLE IF NOT EXISTS user_answers (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES user_quiz_sessions(id) ON DELETE CASCADE,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    selected_option_id INTEGER REFERENCES question_options(id) ON DELETE SET NULL,
    text_answer TEXT,
    is_correct BOOLEAN NOT NULL,
    feedback TEXT NOT NULL,
    answered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, question_id),
    CHECK (
        (selected_option_id IS NOT NULL AND text_answer IS NULL) OR
        (selected_option_id IS NULL AND text_answer IS NOT NULL)
    )
);

-- Indexes for performance (only for tables we'll query frequently)
CREATE INDEX idx_questions_episode_id ON questions(episode_id);
CREATE INDEX idx_user_quiz_sessions_user_id ON user_quiz_sessions(user_id);
CREATE INDEX idx_user_quiz_sessions_user_episode ON user_quiz_sessions(user_id, episode_id);
CREATE INDEX idx_user_answers_session_id ON user_answers(session_id);
