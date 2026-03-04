DROP TABLE IF EXISTS assignment_completions;
-- SQLite does not support DROP COLUMN; assignment_id column on quiz_attempts will remain
-- but will be unused if this migration is rolled back.
