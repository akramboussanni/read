-- SQLite does not support DROP COLUMN easily; columns will remain but be ignored
-- To truly revert, recreate the table without these columns
SELECT 1;
