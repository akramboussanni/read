CREATE TABLE quiz_decks (
    id BIGINT PRIMARY KEY,
    deck_key VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    source_file VARCHAR(255),
    is_system BOOLEAN NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL
);

CREATE TABLE quiz_categories (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE,
    UNIQUE(deck_id, category_key)
);

CREATE TABLE questions (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    question_key VARCHAR(100) NOT NULL UNIQUE,
    
    question_text TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    arabic TEXT NOT NULL,
    french TEXT NOT NULL,
    
    question_type VARCHAR(50) NOT NULL DEFAULT 'multiple_choice',
    difficulty VARCHAR(20),
    points INT NOT NULL DEFAULT 1,
    hint TEXT,
    explanation TEXT,
    
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id),
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id)
);

CREATE TABLE quizzes (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    deck_id BIGINT NOT NULL,
    
    time_limit INT,
    pass_percentage INT,
    shuffle_questions BOOLEAN NOT NULL DEFAULT 1,
    question_mode VARCHAR(20) NOT NULL DEFAULT 'ar_to_fr',
    
    is_system BOOLEAN NOT NULL DEFAULT 0,
    created_by BIGINT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE quiz_category_selections (
    quiz_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    question_count INT NOT NULL,
    PRIMARY KEY (quiz_id, category_id),
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id)
);

CREATE TABLE quiz_attempts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    quiz_id BIGINT NOT NULL,
    
    started_at BIGINT NOT NULL,
    completed_at BIGINT,
    score DECIMAL(10,2),
    max_score DECIMAL(10,2),
    percentage DECIMAL(5,2),
    passed BOOLEAN,
    time_taken INT,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id)
);

CREATE TABLE user_answers (
    id BIGINT PRIMARY KEY,
    attempt_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    points_earned DECIMAL(10,2) NOT NULL DEFAULT 0,
    answered_at BIGINT NOT NULL,
    
    FOREIGN KEY (attempt_id) REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id)
);

CREATE INDEX idx_quiz_decks_key ON quiz_decks(deck_key);
CREATE INDEX idx_categories_deck ON quiz_categories(deck_id);
CREATE INDEX idx_questions_deck ON questions(deck_id);
CREATE INDEX idx_questions_category ON questions(category_id);
CREATE INDEX idx_questions_key ON questions(question_key);
CREATE INDEX idx_quizzes_deck ON quizzes(deck_id);
CREATE INDEX idx_quiz_attempts_user ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz ON quiz_attempts(quiz_id);
CREATE INDEX idx_user_answers_attempt ON user_answers(attempt_id);
