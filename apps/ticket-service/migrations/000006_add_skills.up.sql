-- Add skills column to tickets table as text array
ALTER TABLE tickets ADD COLUMN skills TEXT[] DEFAULT '{}';
