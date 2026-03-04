-- 1. Remove duplicates keeping only the latest one per (attempt_id, question_id)
DELETE FROM user_answers a USING (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY attempt_id, question_id ORDER BY answered_at DESC, id DESC) as row_num
    FROM user_answers
) b
WHERE a.id = b.id AND b.row_num > 1;

-- 2. Add unique constraint to user_answers
ALTER TABLE user_answers ADD CONSTRAINT unique_attempt_question UNIQUE (attempt_id, question_id);
