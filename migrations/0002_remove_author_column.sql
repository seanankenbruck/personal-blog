-- Remove the not-null constraint from the author column
ALTER TABLE posts ALTER COLUMN author DROP NOT NULL;