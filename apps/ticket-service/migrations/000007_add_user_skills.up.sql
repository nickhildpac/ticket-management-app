-- Add skills column to users table as text array
ALTER TABLE users ADD COLUMN skills TEXT[] DEFAULT '{}';
