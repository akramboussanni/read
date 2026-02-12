package model

// Deck represents a collection of questions (e.g., from a JSON file)
type Deck struct {
	ID                     int64    `json:"id,string" db:"id"`
	DeckKey                string   `json:"deck_key" db:"deck_key"`
	Title                  string   `json:"title" db:"title"`
	Version                int      `json:"version" db:"version"`
	DeckType               string   `json:"deck_type" db:"deck_type"`
	LanguagePair           []string `json:"language_pair" db:"language_pair"`
	SupportedQuestionTypes []string `json:"supported_question_types" db:"supported_question_types"`
	DefaultDirection       string   `json:"default_direction" db:"default_direction"`
	DeckMetadata           string   `json:"deck_metadata" db:"deck_metadata"`
	SourceFile             string   `json:"source_file" db:"source_file"`
	IsSystem               bool     `json:"is_system" db:"is_system"`
	CreatedAt              int64    `json:"created_at,string" db:"created_at"`
}

// Category represents a subcategory within a deck
type Category struct {
	ID               int64  `json:"id,string" db:"id"`
	DeckID           int64  `json:"deck_id,string" db:"deck_id"`
	CategoryKey      string `json:"category_key" db:"category_key"`
	Title            string `json:"title" db:"title"`
	Difficulty       string `json:"difficulty" db:"difficulty"`
	CategoryMetadata string `json:"category_metadata" db:"category_metadata"` // JSON data
	DisplayOrder     int    `json:"display_order" db:"display_order"`
	CreatedAt        int64  `json:"created_at,string" db:"created_at"`
}

// Question represents a single quiz question
type Question struct {
	ID            int64   `json:"id,string" db:"id"`
	DeckID        int64   `json:"deck_id,string" db:"deck_id"`
	CategoryID    int64   `json:"category_id,string" db:"category_id"`
	QuestionKey   string  `json:"question_key" db:"question_key"`
	QuestionText  string  `json:"question_text" db:"question_text"`
	CorrectAnswer string  `json:"correct_answer" db:"correct_answer"`
	Arabic        string  `json:"arabic" db:"arabic"`
	French        string  `json:"french" db:"french"`
	QuestionType  string  `json:"question_type" db:"question_type"`
	Difficulty    *string `json:"difficulty,omitempty" db:"difficulty"`
	Points        int     `json:"points" db:"points"`
	Hint          *string `json:"hint,omitempty" db:"hint"`
	Explanation   *string `json:"explanation,omitempty" db:"explanation"`
	CreatedAt     int64   `json:"created_at,string" db:"created_at"`
	UpdatedAt     *int64  `json:"updated_at,string,omitempty" db:"updated_at"`
	IsActive      bool    `json:"is_active" db:"is_active"`
}

// Quiz represents a quiz configuration
type Quiz struct {
	ID                 int64  `json:"id,string" db:"id"`
	Title              string `json:"title" db:"title"`
	Description        string `json:"description" db:"description"`
	Version            int    `json:"version" db:"version"`                  // Incremented on each update
	DeckID             *int64 `json:"deck_id,string,omitempty" db:"deck_id"` // Nullable for multi-deck user quizzes
	TimeLimit          *int   `json:"time_limit,omitempty" db:"time_limit"`  // Always nil
	PassPercentage     *int   `json:"pass_percentage,omitempty" db:"pass_percentage"`
	ShuffleQuestions   bool   `json:"shuffle_questions" db:"shuffle_questions"`
	QuestionMode       string `json:"question_mode" db:"question_mode"` // 'ar_to_fr', 'fr_to_ar'
	GivesCoins         bool   `json:"gives_coins" db:"gives_coins"`
	CoinReward         int    `json:"coin_reward" db:"coin_reward"` // Coins per correct answer
	LevelOrder         int    `json:"level_order" db:"level_order"`
	PrerequisiteQuizID *int64 `json:"prerequisite_quiz_id,string,omitempty" db:"prerequisite_quiz_id"`
	IsPublic           bool   `json:"is_public" db:"is_public"`
	IsSystem           bool   `json:"is_system" db:"is_system"`
	CreatedBy          *int64 `json:"created_by,string,omitempty" db:"created_by"`
	CreatedAt          int64  `json:"created_at,string" db:"created_at"`
	UpdatedAt          *int64 `json:"updated_at,string,omitempty" db:"updated_at"`
	IsActive           bool   `json:"is_active" db:"is_active"`
}

// QuizCategorySelection links quizzes to categories with question counts
type QuizCategorySelection struct {
	QuizID        int64 `json:"quiz_id,string" db:"quiz_id"`
	CategoryID    int64 `json:"category_id,string" db:"category_id"`
	QuestionCount int   `json:"question_count" db:"question_count"`
}

// QuizAttempt represents a user's attempt at a quiz
type QuizAttempt struct {
	ID          int64    `json:"id,string" db:"id"`
	UserID      int64    `json:"user_id,string" db:"user_id"`
	QuizID      int64    `json:"quiz_id,string" db:"quiz_id"`
	StartedAt   int64    `json:"started_at,string" db:"started_at"`
	CompletedAt *int64   `json:"completed_at,string,omitempty" db:"completed_at"`
	Score       *float64 `json:"score,omitempty" db:"score"`
	MaxScore    *float64 `json:"max_score,omitempty" db:"max_score"`
	Percentage  *float64 `json:"percentage,omitempty" db:"percentage"`
	Passed      *bool    `json:"passed,omitempty" db:"passed"`
	TimeTaken   *int     `json:"time_taken,omitempty" db:"time_taken"` // seconds
	CoinsEarned int      `json:"coins_earned" db:"coins_earned"`
}

// UserProgression tracks user's progress through system quizzes
type UserProgression struct {
	UserID                int64   `json:"user_id,string" db:"user_id"`
	CurrentLevel          int     `json:"current_level" db:"current_level"`
	UnlockedQuizIDs       string  `json:"unlocked_quiz_ids" db:"unlocked_quiz_ids"` // JSON array
	LastCompletedQuizID   *int64  `json:"last_completed_quiz_id,string,omitempty" db:"last_completed_quiz_id"`
	TotalQuizzesCompleted int     `json:"total_quizzes_completed" db:"total_quizzes_completed"`
	TotalCoinsEarned      int     `json:"total_coins_earned" db:"total_coins_earned"`
	StreakDays            int     `json:"streak_days" db:"streak_days"`
	LastActivityDate      *string `json:"last_activity_date,omitempty" db:"last_activity_date"` // DATE
	CreatedAt             int64   `json:"created_at,string" db:"created_at"`
	UpdatedAt             *int64  `json:"updated_at,string,omitempty" db:"updated_at"`
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
	ID           int64   `json:"id,string" db:"id"`
	AttemptID    int64   `json:"attempt_id,string" db:"attempt_id"`
	QuestionID   int64   `json:"question_id,string" db:"question_id"`
	UserAnswer   string  `json:"user_answer" db:"user_answer"`
	IsCorrect    bool    `json:"is_correct" db:"is_correct"`
	PointsEarned float64 `json:"points_earned" db:"points_earned"`
	AnsweredAt   int64   `json:"answered_at,string" db:"answered_at"`
}

// AnswerOption represents a possible answer for a question (MCQ)
type AnswerOption struct {
	ID              int64  `json:"id,string" db:"id"`
	QuestionID      int64  `json:"question_id,string" db:"question_id"`
	OptionText      string `json:"option_text" db:"option_text"`
	IsCorrect       bool   `json:"is_correct" db:"is_correct"`
	IsAutoGenerated bool   `json:"is_auto_generated" db:"is_auto_generated"`
	GenerationRule  string `json:"generation_rule,omitempty" db:"generation_rule"` // e.g., "same_category", "same_deck", "random"
	DisplayOrder    int    `json:"display_order" db:"display_order"`
	CreatedAt       int64  `json:"created_at,string" db:"created_at"`
}

// UserCoins tracks user coin balance
type UserCoins struct {
	UserID         int64 `json:"user_id,string" db:"user_id"`
	Balance        int   `json:"balance" db:"balance"`
	LifetimeEarned int   `json:"lifetime_earned" db:"lifetime_earned"`
	LastUpdated    int64 `json:"last_updated,string" db:"last_updated"`
}

// CoinTransaction represents a coin earning or spending event
type CoinTransaction struct {
	ID              int64   `json:"id,string" db:"id"`
	UserID          int64   `json:"user_id,string" db:"user_id"`
	Amount          int     `json:"amount" db:"amount"` // Positive for earn, negative for spend
	TransactionType string  `json:"transaction_type" db:"transaction_type"`
	ReferenceType   *string `json:"reference_type,omitempty" db:"reference_type"` // e.g., "quiz_attempt"
	ReferenceID     *int64  `json:"reference_id,string,omitempty" db:"reference_id"`
	Description     *string `json:"description,omitempty" db:"description"`
	CreatedAt       int64   `json:"created_at,string" db:"created_at"`
}

// DeckEntry represents a flexible entry in a universal deck
type DeckEntry struct {
	ID         int64  `json:"id,string" db:"id"`
	DeckID     int64  `json:"deck_id,string" db:"deck_id"`
	CategoryID int64  `json:"category_id,string" db:"category_id"`
	EntryKey   string `json:"entry_key" db:"entry_key"`
	EntryData  string `json:"entry_data" db:"entry_data"` // JSON data
	Tags       string `json:"tags" db:"tags"`             // JSON array
	CreatedAt  int64  `json:"created_at,string" db:"created_at"`
	UpdatedAt  *int64 `json:"updated_at,string,omitempty" db:"updated_at"`
}

// QuizQuestion represents a generated question for a specific quiz
type QuizQuestion struct {
	ID            int64   `json:"id" db:"id"`
	QuizID        int64   `json:"quiz_id" db:"quiz_id"`
	QuestionID    *int64  `json:"question_id,string,omitempty" db:"question_id"` // Reference to original question, NULL for custom
	QuestionText  string  `json:"question_text" db:"question_text"`
	CorrectAnswer string  `json:"correct_answer" db:"correct_answer"`
	Options       string  `json:"options" db:"options"` // JSON array
	QuestionType  string  `json:"question_type" db:"question_type"`
	Direction     *string `json:"direction,omitempty" db:"direction"`
	DisplayOrder  int     `json:"display_order" db:"display_order"`
	CreatedAt     int64   `json:"created_at" db:"created_at"`
}

// QuizSession represents temporary quiz state
type QuizSession struct {
	ID                   string `json:"id" db:"id"`
	QuizID               int64  `json:"quiz_id,string" db:"quiz_id"`
	UserID               int64  `json:"user_id,string" db:"user_id"`
	CurrentQuestionIndex int    `json:"current_question_index" db:"current_question_index"`
	StartedAt            int64  `json:"started_at,string" db:"started_at"`
	LastActivity         int64  `json:"last_activity,string" db:"last_activity"`
	IsCompleted          bool   `json:"is_completed" db:"is_completed"`
}

// DeckCache represents cached deck data for performance
type DeckCache struct {
	DeckID       int64  `json:"deck_id,string" db:"deck_id"`
	CachedData   string `json:"cached_data" db:"cached_data"` // JSON data
	CacheVersion int    `json:"cache_version" db:"cache_version"`
	LastUpdated  int64  `json:"last_updated,string" db:"last_updated"`
}
