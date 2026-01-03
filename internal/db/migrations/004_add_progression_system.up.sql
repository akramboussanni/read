-- Add user progression tracking
CREATE TABLE user_progression (
    user_id BIGINT PRIMARY KEY,
    current_level INT NOT NULL DEFAULT 1,
    unlocked_quiz_ids TEXT, -- JSON array of unlocked quiz IDs
    last_completed_quiz_id BIGINT,
    total_quizzes_completed INT NOT NULL DEFAULT 0,
    total_coins_earned INT NOT NULL DEFAULT 0,
    streak_days INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Add level/order to quizzes for progression system
ALTER TABLE quizzes ADD COLUMN level_order INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN prerequisite_quiz_id BIGINT;
ALTER TABLE quizzes ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT 1;

-- Add indexes for progression queries
CREATE INDEX idx_user_progression_user ON user_progression(user_id);
CREATE INDEX idx_quizzes_level_order ON quizzes(level_order, is_system);
CREATE INDEX idx_quizzes_public ON quizzes(is_public, is_active);
CREATE INDEX idx_quizzes_creator ON quizzes(created_by, is_active);

-- Track quiz attempts for history
ALTER TABLE quiz_attempts ADD COLUMN coins_earned INT DEFAULT 0;

CREATE INDEX idx_quiz_attempts_completed ON quiz_attempts(user_id, completed_at);
