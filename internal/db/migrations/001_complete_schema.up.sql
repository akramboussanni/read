-- Complete database schema for quiz application
-- Combined from all migration files

-- Users and authentication
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT DEFAULT '',
    password_hash TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    user_role TEXT NOT NULL,
    email_confirmed BOOLEAN NOT NULL DEFAULT false,
    email_confirm_token VARCHAR(64),
    email_confirm_issuedat BIGINT,
    password_reset_token VARCHAR(64),
    password_reset_issuedat BIGINT,
    jwt_session_id BIGINT,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    pending_email TEXT DEFAULT ''
);

CREATE UNIQUE INDEX users_email_key ON users(email) WHERE email != '';

CREATE TABLE jwt_blacklist (
    jti VARCHAR(255) PRIMARY KEY,
    user_id BIGINT,
    expires_at BIGINT NOT NULL
);

CREATE TABLE failed_logins (
    id BIGINT PRIMARY KEY,
    user_id INT NULL,
    ip_address VARCHAR(45) NOT NULL,
    attempted_at BIGINT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_failed_logins_user ON failed_logins(user_id);
CREATE INDEX idx_failed_logins_ip ON failed_logins(ip_address);
CREATE INDEX idx_failed_logins_attempted_at ON failed_logins(attempted_at);

CREATE TABLE lockouts (
    id BIGINT PRIMARY KEY,
    user_id INT NULL,
    ip_address VARCHAR(45) NULL,
    locked_until BIGINT NOT NULL,
    reason VARCHAR(255) NULL,
    active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_lockouts_user ON lockouts(user_id);
CREATE INDEX idx_lockouts_ip ON lockouts(ip_address);
CREATE INDEX idx_lockouts_locked_until ON lockouts(locked_until);

-- Quiz system tables
CREATE TABLE quiz_decks (
    id BIGINT PRIMARY KEY,
    deck_key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    source_file VARCHAR(255),
    is_system BOOLEAN NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    deck_type VARCHAR(50) DEFAULT 'vocabulary',
    language_pair JSON,
    supported_question_types JSON,
    default_direction VARCHAR(50) DEFAULT 'source_to_target',
    deck_metadata JSON,
    CONSTRAINT uk_deck_key_version UNIQUE(deck_key, version)
);

CREATE TABLE quiz_categories (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    difficulty VARCHAR(20),
    category_metadata JSON,
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

    question_type VARCHAR(50) NOT NULL DEFAULT 'mcq',
    direction VARCHAR(50),
    difficulty VARCHAR(20),
    points INT NOT NULL DEFAULT 1,
    hint TEXT,
    explanation TEXT,
    additional_data JSON,
    tags JSON,
    llm_generated BOOLEAN DEFAULT FALSE,
    validation_rules JSON,

    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT 1,

    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id),
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id)
);

CREATE TABLE question_options (
    id BIGINT PRIMARY KEY,
    question_id BIGINT NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT 0,
    is_auto_generated BOOLEAN NOT NULL DEFAULT 0,
    generation_rule VARCHAR(100),
    display_order INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    option_metadata JSON,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
);

CREATE TABLE deck_entries (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    entry_key VARCHAR(100) NOT NULL,
    entry_data JSON NOT NULL,
    tags JSON,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id) ON DELETE CASCADE,
    UNIQUE(deck_id, entry_key)
);

CREATE TABLE deck_cache (
    deck_id BIGINT PRIMARY KEY,
    cached_data JSON NOT NULL,
    cache_version INT NOT NULL DEFAULT 1,
    last_updated BIGINT NOT NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE
);

CREATE TABLE quizzes (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    version INT NOT NULL DEFAULT 1,
    deck_id BIGINT,

    time_limit INT,
    pass_percentage INT,
    shuffle_questions BOOLEAN NOT NULL DEFAULT 1,
    question_mode VARCHAR(20) NOT NULL DEFAULT 'ar_to_fr',
    gives_coins BOOLEAN NOT NULL DEFAULT 0,
    coin_reward INT DEFAULT 0,
    level_order INT DEFAULT 0,
    prerequisite_quiz_id BIGINT,
    is_public BOOLEAN NOT NULL DEFAULT 1,

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

CREATE TABLE quiz_questions (
    id BIGINT PRIMARY KEY,
    quiz_id BIGINT NOT NULL,
    question_id BIGINT,
    question_text TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    options JSON,
    question_type VARCHAR(50) NOT NULL,
    direction VARCHAR(50),
    display_order INT NOT NULL,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id)
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
    coins_earned INT DEFAULT 0,

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

CREATE TABLE quiz_sessions (
    id VARCHAR(255) PRIMARY KEY,
    quiz_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    current_question_index INT NOT NULL DEFAULT 0,
    started_at BIGINT NOT NULL,
    last_activity BIGINT NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User progression and rewards
CREATE TABLE user_progression (
    user_id BIGINT PRIMARY KEY,
    current_level INT NOT NULL DEFAULT 1,
    unlocked_quiz_ids TEXT,
    last_completed_quiz_id BIGINT,
    total_quizzes_completed INT NOT NULL DEFAULT 0,
    total_coins_earned INT NOT NULL DEFAULT 0,
    streak_days INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_coins (
    user_id BIGINT PRIMARY KEY,
    balance INT NOT NULL DEFAULT 0,
    lifetime_earned INT NOT NULL DEFAULT 0,
    last_updated BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

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

-- Indexes for performance
CREATE INDEX idx_quiz_decks_key ON quiz_decks(deck_key);
CREATE INDEX idx_categories_deck ON quiz_categories(deck_id);
CREATE INDEX idx_questions_deck ON questions(deck_id);
CREATE INDEX idx_questions_category ON questions(category_id);
CREATE INDEX idx_questions_key ON questions(question_key);
CREATE INDEX idx_questions_type ON questions(question_type);
CREATE INDEX idx_questions_direction ON questions(direction);
CREATE INDEX idx_questions_tags ON questions USING GIN(tags);
CREATE INDEX idx_quizzes_deck ON quizzes(deck_id);
CREATE INDEX idx_quizzes_level_order ON quizzes(level_order, is_system);
CREATE INDEX idx_quizzes_public ON quizzes(is_public, is_active);
CREATE INDEX idx_quizzes_creator ON quizzes(created_by, is_active);
CREATE INDEX idx_quiz_attempts_user ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_completed ON quiz_attempts(user_id, completed_at);
CREATE INDEX idx_user_answers_attempt ON user_answers(attempt_id);
CREATE INDEX idx_question_options_question ON question_options(question_id);
CREATE INDEX idx_quiz_questions_quiz ON quiz_questions(quiz_id);
CREATE INDEX idx_quiz_questions_question ON quiz_questions(question_id);
CREATE INDEX idx_quiz_sessions_user ON quiz_sessions(user_id);
CREATE INDEX idx_quiz_sessions_quiz ON quiz_sessions(quiz_id);
CREATE INDEX idx_deck_entries_deck ON deck_entries(deck_id);
CREATE INDEX idx_deck_entries_category ON deck_entries(category_id);
CREATE INDEX idx_deck_entries_tags ON deck_entries USING GIN(tags);
CREATE INDEX idx_user_progression_user ON user_progression(user_id);
CREATE INDEX idx_coin_transactions_user ON coin_transactions(user_id);
CREATE INDEX idx_coin_transactions_reference ON coin_transactions(reference_type, reference_id);