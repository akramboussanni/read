-- Complete database schema for quiz application (v2 - Major Recode)
-- This replaces all previous migrations

-- ============================================================
-- USERS & AUTH
-- ============================================================

CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT DEFAULT '',
    password_hash TEXT NOT NULL,
    display_name TEXT DEFAULT '',
    avatar_url TEXT DEFAULT '',
    onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE,
    active_course_id TEXT, -- current course the user is viewing
    created_at BIGINT NOT NULL,
    user_role TEXT NOT NULL,
    email_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    email_confirm_token VARCHAR(64),
    email_confirm_issuedat BIGINT,
    password_reset_token VARCHAR(64),
    password_reset_issuedat BIGINT,
    jwt_session_id BIGINT,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
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
    user_id BIGINT NULL,
    ip_address VARCHAR(45) NOT NULL,
    attempted_at BIGINT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_failed_logins_user ON failed_logins(user_id);
CREATE INDEX idx_failed_logins_ip ON failed_logins(ip_address);
CREATE INDEX idx_failed_logins_attempted_at ON failed_logins(attempted_at);

CREATE TABLE lockouts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NULL,
    ip_address VARCHAR(45) NULL,
    locked_until BIGINT NOT NULL,
    reason VARCHAR(255) NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_lockouts_user ON lockouts(user_id);
CREATE INDEX idx_lockouts_ip ON lockouts(ip_address);
CREATE INDEX idx_lockouts_locked_until ON lockouts(locked_until);

-- ============================================================
-- DECKS & ENTRIES (content library)
-- ============================================================

CREATE TABLE quiz_decks (
    id BIGINT PRIMARY KEY,
    deck_key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    version INT NOT NULL DEFAULT 1,
    deck_type VARCHAR(50) DEFAULT 'vocabulary',
    language_pair JSONB,
    supported_question_types JSONB,
    default_direction VARCHAR(50) DEFAULT 'source_to_target',
    deck_metadata JSONB,
    source_file VARCHAR(255),
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at BIGINT NOT NULL,
    CONSTRAINT uk_deck_key_version UNIQUE(deck_key, version)
);

CREATE TABLE quiz_categories (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    difficulty VARCHAR(20) DEFAULT 'beginner',
    category_metadata JSONB,
    display_order INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE,
    UNIQUE(deck_id, category_key)
);

CREATE TABLE deck_entries (
    id BIGINT PRIMARY KEY,
    deck_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    entry_key VARCHAR(100) NOT NULL,
    entry_data JSONB NOT NULL,
    tags JSONB,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id) ON DELETE CASCADE,
    UNIQUE(deck_id, entry_key)
);

CREATE TABLE deck_cache (
    deck_id BIGINT PRIMARY KEY,
    cached_data JSONB NOT NULL,
    cache_version INT NOT NULL DEFAULT 1,
    last_updated BIGINT NOT NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE CASCADE
);

-- ============================================================
-- COURSES (Progression Paths)
-- ============================================================

CREATE TABLE courses (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    icon TEXT DEFAULT 'book',       -- lucide icon name
    color TEXT DEFAULT '#6C5CE7',   -- hex color for theming
    deck_id BIGINT,                 -- optional: source deck for auto-generation
    is_default BOOLEAN DEFAULT FALSE,
    is_published BOOLEAN DEFAULT TRUE,
    created_by BIGINT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE course_nodes (
    id TEXT PRIMARY KEY,
    course_id TEXT NOT NULL,
    node_type TEXT NOT NULL,        -- 'lesson', 'quiz', 'milestone', 'start', 'checkpoint'
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    icon TEXT DEFAULT '',
    position_x REAL NOT NULL DEFAULT 0,
    position_y REAL NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    -- Quiz configuration (for quiz nodes)
    quiz_config JSONB,              -- question types, directions, count, category constraints, etc.
    -- Lesson content (for lesson nodes)
    lesson_content JSONB,           -- markdown content, media links, etc.
    metadata JSONB,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE course_edges (
    id TEXT PRIMARY KEY,
    course_id TEXT NOT NULL,
    source_node_id TEXT NOT NULL,
    target_node_id TEXT NOT NULL,
    label TEXT DEFAULT '',
    edge_type TEXT DEFAULT 'required',  -- 'required', 'optional', 'bonus'
    created_at BIGINT NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (source_node_id) REFERENCES course_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (target_node_id) REFERENCES course_nodes(id) ON DELETE CASCADE
);

-- ============================================================
-- USER ENROLLMENT & PROGRESS
-- ============================================================

CREATE TABLE user_enrollments (
    id TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    course_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',  -- 'active', 'completed', 'paused'
    progress REAL DEFAULT 0,
    current_node_id TEXT,
    completed_nodes JSONB DEFAULT '[]',      -- array of completed node IDs
    enrolled_at BIGINT NOT NULL,
    completed_at BIGINT,
    last_accessed_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    UNIQUE(user_id, course_id)
);

-- ============================================================
-- QUIZZES & ATTEMPTS
-- ============================================================

CREATE TABLE quizzes (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    course_id TEXT,                  -- optional: belongs to a course
    node_id TEXT,                    -- optional: linked to a course node
    deck_id BIGINT,
    pass_percentage INT DEFAULT 70,
    shuffle_questions BOOLEAN NOT NULL DEFAULT TRUE,
    question_mode VARCHAR(30) DEFAULT 'mixed',  -- 'source_to_target', 'target_to_source', 'mixed'
    gives_coins BOOLEAN NOT NULL DEFAULT FALSE,
    coin_reward INT DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    is_dynamic BOOLEAN NOT NULL DEFAULT TRUE,   -- questions generated at runtime
    created_by BIGINT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL,
    FOREIGN KEY (node_id) REFERENCES course_nodes(id) ON DELETE SET NULL,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Question templates: describe HOW to generate questions at runtime
CREATE TABLE question_templates (
    id BIGINT PRIMARY KEY,
    quiz_id BIGINT NOT NULL,
    deck_id BIGINT,
    category_id BIGINT,
    question_types JSONB NOT NULL DEFAULT '["mcq","translate"]',
    directions JSONB NOT NULL DEFAULT '["source_to_target","target_to_source"]',
    generation_mode TEXT NOT NULL DEFAULT 'random_from_deck',  -- 'random_from_deck', 'llm', 'manual'
    llm_prompt TEXT,
    manual_data JSONB,              -- for manually defined questions: {question_text, correct_answer, options}
    question_count INT NOT NULL DEFAULT 5,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    FOREIGN KEY (deck_id) REFERENCES quiz_decks(id) ON DELETE SET NULL,
    FOREIGN KEY (category_id) REFERENCES quiz_categories(id) ON DELETE SET NULL
);

CREATE TABLE quiz_attempts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    quiz_id BIGINT NOT NULL,
    course_id TEXT,
    node_id TEXT,
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

-- Generated questions for a specific attempt
CREATE TABLE attempt_questions (
    id BIGINT PRIMARY KEY,
    attempt_id BIGINT NOT NULL,
    quiz_id BIGINT NOT NULL,
    question_text TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    options JSONB,                    -- for MCQ: ["opt1","opt2","opt3","opt4"]
    question_type VARCHAR(50) NOT NULL,
    direction VARCHAR(50),
    display_order INT NOT NULL,
    source_entry_id BIGINT,          -- reference to deck_entry if generated from deck
    generation_mode TEXT NOT NULL DEFAULT 'random_from_deck',
    created_at BIGINT NOT NULL,
    FOREIGN KEY (attempt_id) REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE
);

-- User answers per question
CREATE TABLE user_answers (
    id BIGINT PRIMARY KEY,
    attempt_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    points_earned DECIMAL(10,2) NOT NULL DEFAULT 0,
    ai_explanation TEXT,
    answered_at BIGINT NOT NULL,
    FOREIGN KEY (attempt_id) REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES attempt_questions(id) ON DELETE CASCADE
);

-- ============================================================
-- COINS & REWARDS
-- ============================================================

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

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_quiz_decks_key ON quiz_decks(deck_key);
CREATE INDEX idx_categories_deck ON quiz_categories(deck_id);
CREATE INDEX idx_deck_entries_deck ON deck_entries(deck_id);
CREATE INDEX idx_deck_entries_category ON deck_entries(category_id);
CREATE INDEX idx_deck_entries_tags ON deck_entries USING GIN(tags);

CREATE INDEX idx_courses_published ON courses(is_published);
CREATE INDEX idx_courses_default ON courses(is_default);
CREATE INDEX idx_course_nodes_course ON course_nodes(course_id);
CREATE INDEX idx_course_edges_course ON course_edges(course_id);

CREATE INDEX idx_enrollments_user ON user_enrollments(user_id);
CREATE INDEX idx_enrollments_course ON user_enrollments(course_id);

CREATE INDEX idx_quizzes_course ON quizzes(course_id);
CREATE INDEX idx_quizzes_public ON quizzes(is_public, is_active);
CREATE INDEX idx_quizzes_creator ON quizzes(created_by, is_active);
CREATE INDEX idx_question_templates_quiz ON question_templates(quiz_id);

CREATE INDEX idx_quiz_attempts_user ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz ON quiz_attempts(quiz_id);
CREATE INDEX idx_attempt_questions_attempt ON attempt_questions(attempt_id);
CREATE INDEX idx_user_answers_attempt ON user_answers(attempt_id);

CREATE INDEX idx_coin_transactions_user ON coin_transactions(user_id);
