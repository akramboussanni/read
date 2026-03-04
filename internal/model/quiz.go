package model

// Deck represents a collection of entries (e.g., vocabulary deck)
type Deck struct {
	ID                     int64  `json:"id,string" db:"id"`
	DeckKey                string `json:"deck_key" db:"deck_key"`
	Title                  string `json:"title" db:"title"`
	Description            string `json:"description" db:"description"`
	Version                int    `json:"version" db:"version"`
	DeckType               string `json:"deck_type" db:"deck_type"`
	LanguagePair           string `json:"language_pair" db:"language_pair"`
	SupportedQuestionTypes string `json:"supported_question_types" db:"supported_question_types"`
	DefaultDirection       string `json:"default_direction" db:"default_direction"`
	DeckMetadata           string `json:"deck_metadata" db:"deck_metadata"`
	SourceFile             string `json:"source_file" db:"source_file"`
	IsSystem               bool   `json:"is_system" db:"is_system"`
	CreatedAt              int64  `json:"created_at,string" db:"created_at"`
}

// Category represents a subcategory within a deck
type Category struct {
	ID               int64  `json:"id,string" db:"id"`
	DeckID           int64  `json:"deck_id,string" db:"deck_id"`
	CategoryKey      string `json:"category_key" db:"category_key"`
	Title            string `json:"title" db:"title"`
	Difficulty       string `json:"difficulty" db:"difficulty"`
	CategoryMetadata string `json:"category_metadata" db:"category_metadata"`
	DisplayOrder     int    `json:"display_order" db:"display_order"`
	CreatedAt        int64  `json:"created_at,string" db:"created_at"`
}

// DeckEntry represents a flexible entry in a universal deck
type DeckEntry struct {
	ID         int64  `json:"id,string" db:"id"`
	DeckID     int64  `json:"deck_id,string" db:"deck_id"`
	CategoryID int64  `json:"category_id,string" db:"category_id"`
	EntryKey   string `json:"entry_key" db:"entry_key"`
	EntryData  string `json:"entry_data" db:"entry_data"` // JSON
	Tags       string `json:"tags" db:"tags"`             // JSON array
	CreatedAt  int64  `json:"created_at,string" db:"created_at"`
	UpdatedAt  *int64 `json:"updated_at,string,omitempty" db:"updated_at"`
}

// DeckCache represents cached deck data
type DeckCache struct {
	DeckID       int64  `json:"deck_id,string" db:"deck_id"`
	CachedData   string `json:"cached_data" db:"cached_data"`
	CacheVersion int    `json:"cache_version" db:"cache_version"`
	LastUpdated  int64  `json:"last_updated,string" db:"last_updated"`
}

// Quiz represents a quiz configuration
type Quiz struct {
	ID               int64   `json:"id,string" db:"id"`
	Title            string  `json:"title" db:"title"`
	Description      string  `json:"description" db:"description"`
	CourseID         *string `json:"course_id,omitempty" db:"course_id"`
	NodeID           *string `json:"node_id,omitempty" db:"node_id"`
	DeckID           *int64  `json:"deck_id,string,omitempty" db:"deck_id"`
	PassPercentage   *int    `json:"pass_percentage,omitempty" db:"pass_percentage"`
	ShuffleQuestions bool    `json:"shuffle_questions" db:"shuffle_questions"`
	QuestionMode     string  `json:"question_mode" db:"question_mode"`
	GivesCoins       bool    `json:"gives_coins" db:"gives_coins"`
	CoinReward       int     `json:"coin_reward" db:"coin_reward"`
	IsPublic         bool    `json:"is_public" db:"is_public"`
	IsSystem         bool    `json:"is_system" db:"is_system"`
	IsDynamic        bool    `json:"is_dynamic" db:"is_dynamic"`
	CreatedBy        *int64  `json:"created_by,string,omitempty" db:"created_by"`
	CreatedAt        int64   `json:"created_at,string" db:"created_at"`
	UpdatedAt        *int64  `json:"updated_at,string,omitempty" db:"updated_at"`
	IsActive         bool    `json:"is_active" db:"is_active"`
}

// QuestionTemplate defines HOW to generate questions at runtime
type QuestionTemplate struct {
	ID             int64   `json:"id,string" db:"id"`
	QuizID         int64   `json:"quiz_id,string" db:"quiz_id"`
	DeckID         *int64  `json:"deck_id,string,omitempty" db:"deck_id"`
	CategoryID     *int64  `json:"category_id,string,omitempty" db:"category_id"`
	QuestionTypes  string  `json:"question_types" db:"question_types"`   // JSON array
	Directions     string  `json:"directions" db:"directions"`           // JSON array
	GenerationMode string  `json:"generation_mode" db:"generation_mode"` // 'random_from_deck','llm','manual'
	LLMPrompt      string  `json:"llm_prompt,omitempty" db:"llm_prompt"`
	ManualData     *string `json:"manual_data,omitempty" db:"manual_data"` // JSON
	QuestionCount  int     `json:"question_count" db:"question_count"`
	CreatedAt      int64   `json:"created_at,string" db:"created_at"`
}

// QuizAttempt represents a user's attempt at a quiz
type QuizAttempt struct {
	ID           int64    `json:"id,string" db:"id"`
	UserID       int64    `json:"user_id,string" db:"user_id"`
	QuizID       int64    `json:"quiz_id,string" db:"quiz_id"`
	CourseID     *string  `json:"course_id,omitempty" db:"course_id"`
	NodeID       *string  `json:"node_id,omitempty" db:"node_id"`
	AssignmentID *int64   `json:"assignment_id,string,omitempty" db:"assignment_id"`
	StartedAt    int64    `json:"started_at,string" db:"started_at"`
	CompletedAt  *int64   `json:"completed_at,string,omitempty" db:"completed_at"`
	Score        *float64 `json:"score,omitempty" db:"score"`
	MaxScore     *float64 `json:"max_score,omitempty" db:"max_score"`
	Percentage   *float64 `json:"percentage,omitempty" db:"percentage"`
	Passed       *bool    `json:"passed,omitempty" db:"passed"`
	TimeTaken    *int     `json:"time_taken,omitempty" db:"time_taken"`
	CoinsEarned  int      `json:"coins_earned" db:"coins_earned"`
}

// AttemptQuestion represents a generated question for a specific attempt
type AttemptQuestion struct {
	ID             int64   `json:"id,string" db:"id"`
	AttemptID      int64   `json:"attempt_id,string" db:"attempt_id"`
	QuizID         int64   `json:"quiz_id,string" db:"quiz_id"`
	QuestionText   string  `json:"question_text" db:"question_text"`
	CorrectAnswer  string  `json:"correct_answer" db:"correct_answer"`
	Options        string  `json:"options" db:"options"` // JSON array
	QuestionType   string  `json:"question_type" db:"question_type"`
	Direction      *string `json:"direction,omitempty" db:"direction"`
	DisplayOrder   int     `json:"display_order" db:"display_order"`
	SourceEntryID  *int64  `json:"source_entry_id,string,omitempty" db:"source_entry_id"`
	GenerationMode string  `json:"generation_mode" db:"generation_mode"`
	CreatedAt      int64   `json:"created_at" db:"created_at"`
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
	ID            int64   `json:"id,string" db:"id"`
	AttemptID     int64   `json:"attempt_id,string" db:"attempt_id"`
	QuestionID    int64   `json:"question_id,string" db:"question_id"`
	UserAnswer    string  `json:"user_answer" db:"user_answer"`
	IsCorrect     bool    `json:"is_correct" db:"is_correct"`
	PointsEarned  float64 `json:"points_earned" db:"points_earned"`
	AIExplanation *string `json:"ai_explanation,omitempty" db:"ai_explanation"`
	AnsweredAt    int64   `json:"answered_at,string" db:"answered_at"`
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
	Amount          int     `json:"amount" db:"amount"`
	TransactionType string  `json:"transaction_type" db:"transaction_type"`
	ReferenceType   *string `json:"reference_type,omitempty" db:"reference_type"`
	ReferenceID     *int64  `json:"reference_id,string,omitempty" db:"reference_id"`
	Description     *string `json:"description,omitempty" db:"description"`
	CreatedAt       int64   `json:"created_at,string" db:"created_at"`
}
