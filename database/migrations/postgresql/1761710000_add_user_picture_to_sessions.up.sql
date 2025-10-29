-- Add user picture URL to sessions table
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS user_picture TEXT;

