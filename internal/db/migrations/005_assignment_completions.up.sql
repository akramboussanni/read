-- Assignment completions: separate tracking from course progression
CREATE TABLE IF NOT EXISTS assignment_completions (
    id BIGINT PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES classroom_assignments(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_id BIGINT REFERENCES quiz_attempts(id) ON DELETE SET NULL,
    score REAL DEFAULT 0,
    max_score REAL DEFAULT 0,
    percentage REAL DEFAULT 0,
    passed BOOLEAN DEFAULT FALSE,
    completed_at BIGINT NOT NULL,
    UNIQUE (assignment_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_assignment_completions_assignment ON assignment_completions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_assignment_completions_student ON assignment_completions(student_id);

-- Add assignment_id to quiz_attempts so we can trace which assignment triggered it
ALTER TABLE quiz_attempts ADD COLUMN assignment_id BIGINT REFERENCES classroom_assignments(id) ON DELETE SET NULL;
