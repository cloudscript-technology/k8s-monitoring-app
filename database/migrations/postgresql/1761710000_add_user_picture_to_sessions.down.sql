-- Remove user picture URL from sessions table
ALTER TABLE sessions DROP COLUMN IF EXISTS user_picture;

