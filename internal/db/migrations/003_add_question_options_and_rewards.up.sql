-- Add question options table for MCQ answers
CREATE TABLE question_options (
    id BIGINT PRIMARY KEY,
    question_id BIGINT NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT 0,
    is_auto_generated BOOLEAN NOT NULL DEFAULT 0,
    generation_rule VARCHAR(100),
    display_order INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
);

CREATE INDEX idx_question_options_question ON question_options(question_id);

-- Add coin rewards to quizzes
ALTER TABLE quizzes ADD COLUMN gives_coins BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN coin_reward INT DEFAULT 0;

-- Add coins to users table (if not exists already)
CREATE TABLE IF NOT EXISTS user_coins (
    user_id BIGINT PRIMARY KEY,
    balance INT NOT NULL DEFAULT 0,
    lifetime_earned INT NOT NULL DEFAULT 0,
    last_updated BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Track coin transactions
CREATE TABLE coin_transactions (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    reference_type VARCHAR(50),
    reference_id BIGINT,
    description TEXT,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_coin_transactions_user ON coin_transactions(user_id);
CREATE INDEX idx_coin_transactions_reference ON coin_transactions(reference_type, reference_id);
