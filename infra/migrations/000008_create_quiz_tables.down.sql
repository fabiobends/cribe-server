-- Drop tables in reverse order
DROP TABLE IF EXISTS user_answers;
DROP TABLE IF EXISTS user_quiz_sessions;
DROP TABLE IF EXISTS question_options;
DROP TABLE IF EXISTS questions;

-- Drop enum types
DROP TYPE IF EXISTS session_status;
DROP TYPE IF EXISTS question_type;
