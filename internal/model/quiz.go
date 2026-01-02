package model

// Deck represents a collection of questions (e.g., from a JSON file)
type Deck struct {
	ID         int64  `json:"id" db:"id"`
	DeckKey    string `json:"deck_key" db:"deck_key"`
	Title      string `json:"title" db:"title"`
	Version    int    `json:"version" db:"version"`
	SourceFile string `json:"source_file" db:"source_file"`
	IsSystem   bool   `json:"is_system" db:"is_system"`
	CreatedAt  int64  `json:"created_at" db:"created_at"`
}

// Category represents a subcategory within a deck
type Category struct {
	ID           int64  `json:"id" db:"id"`
	DeckID       int64  `json:"deck_id" db:"deck_id"`
	CategoryKey  string `json:"category_key" db:"category_key"`
	Title        string `json:"title" db:"title"`
	DisplayOrder int    `json:"display_order" db:"display_order"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

// Question represents a single quiz question
type Question struct {
	ID            int64   `json:"id" db:"id"`
	DeckID        int64   `json:"deck_id" db:"deck_id"`
	CategoryID    int64   `json:"category_id" db:"category_id"`
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
	CreatedAt     int64   `json:"created_at" db:"created_at"`
	UpdatedAt     *int64  `json:"updated_at,omitempty" db:"updated_at"`
	IsActive      bool    `json:"is_active" db:"is_active"`
}

// Quiz represents a quiz configuration
type Quiz struct {
	ID               int64  `json:"id" db:"id"`
	Title            string `json:"title" db:"title"`
	Description      string `json:"description" db:"description"`
	DeckID           int64  `json:"deck_id" db:"deck_id"`
	TimeLimit        *int   `json:"time_limit,omitempty" db:"time_limit"`
	PassPercentage   *int   `json:"pass_percentage,omitempty" db:"pass_percentage"`
	ShuffleQuestions bool   `json:"shuffle_questions" db:"shuffle_questions"`
	QuestionMode     string `json:"question_mode" db:"question_mode"` // 'ar_to_fr', 'fr_to_ar'
	IsSystem         bool   `json:"is_system" db:"is_system"`
	CreatedBy        *int64 `json:"created_by,omitempty" db:"created_by"`
	CreatedAt        int64  `json:"created_at" db:"created_at"`
	UpdatedAt        *int64 `json:"updated_at,omitempty" db:"updated_at"`
	IsActive         bool   `json:"is_active" db:"is_active"`
}

// QuizCategorySelection links quizzes to categories with question counts
type QuizCategorySelection struct {
	QuizID        int64 `json:"quiz_id" db:"quiz_id"`
	CategoryID    int64 `json:"category_id" db:"category_id"`
	QuestionCount int   `json:"question_count" db:"question_count"`
}

// QuizAttempt represents a user's attempt at a quiz
type QuizAttempt struct {
	ID          int64    `json:"id" db:"id"`
	UserID      int64    `json:"user_id" db:"user_id"`
	QuizID      int64    `json:"quiz_id" db:"quiz_id"`
	StartedAt   int64    `json:"started_at" db:"started_at"`
	CompletedAt *int64   `json:"completed_at,omitempty" db:"completed_at"`
	Score       *float64 `json:"score,omitempty" db:"score"`
	MaxScore    *float64 `json:"max_score,omitempty" db:"max_score"`
	Percentage  *float64 `json:"percentage,omitempty" db:"percentage"`
	Passed      *bool    `json:"passed,omitempty" db:"passed"`
	TimeTaken   *int     `json:"time_taken,omitempty" db:"time_taken"` // seconds
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
	ID           int64   `json:"id" db:"id"`
	AttemptID    int64   `json:"attempt_id" db:"attempt_id"`
	QuestionID   int64   `json:"question_id" db:"question_id"`
	UserAnswer   string  `json:"user_answer" db:"user_answer"`
	IsCorrect    bool    `json:"is_correct" db:"is_correct"`
	PointsEarned float64 `json:"points_earned" db:"points_earned"`
	AnsweredAt   int64   `json:"answered_at" db:"answered_at"`
}
