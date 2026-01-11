package quiz

import (
	"time"

	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
)

// API Request/Response models for the quiz system

// DeckListItem represents a deck in list responses
type DeckListItem struct {
	ID            int64  `json:"id,string"`
	DeckKey       string `json:"deck_key"`
	Title         string `json:"title"`
	CategoryCount int    `json:"category_count"`
	QuestionCount int    `json:"question_count"`
}

// CategoryListItem represents a category in list responses
type CategoryListItem struct {
	ID            int64  `json:"id,string"`
	CategoryKey   string `json:"category_key"`
	Title         string `json:"title"`
	QuestionCount int    `json:"question_count"`
}

// QuizListItem represents a quiz in list responses
type QuizListItem struct {
	ID            int64     `json:"id,string"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	CreatorName   string    `json:"creator_name,omitempty"`
	QuestionCount int       `json:"question_count"`
	IsPublic      bool      `json:"is_public"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateQuizRequest represents a request to create a quiz
type CreateQuizRequest struct {
	Title           string                  `json:"title"`
	Description     string                  `json:"description,omitempty"`
	ManualQuestions []ManualQuestionRequest `json:"manual_questions,omitempty"` // User-created questions
	AutoGenerate    *AutoGenerateRequest    `json:"auto_generate,omitempty"`    // Auto-generation config
	IsPublic        bool                    `json:"is_public"`
	// Admin-only fields
	PassPercentage     *int   `json:"pass_percentage,omitempty"`             // Admin only
	GivesCoins         bool   `json:"gives_coins"`                           // Admin only (default false)
	CoinReward         int    `json:"coin_reward"`                           // Admin only (default 0)
	LevelOrder         int    `json:"level_order"`                           // Admin only (default 0)
	PrerequisiteQuizID *int64 `json:"prerequisite_quiz_id,string,omitempty"` // Admin only
	IsSystem           bool   `json:"is_system"`                             // Admin only (default false)
}

// DeckSelectionRequest represents a selection of categories from a specific deck
type DeckSelectionRequest struct {
	DeckID     int64    `json:"deck_id,string"`
	Categories []string `json:"categories,omitempty"` // category keys to include, empty means all
}

// CustomQuestionRequest represents a custom question in create request
type CustomQuestionRequest struct {
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	WrongAnswers  []string `json:"wrong_answers,omitempty"` // For MCQ
	QuestionType  string   `json:"question_type"`
}

// ManualQuestionRequest represents a manually created question
type ManualQuestionRequest struct {
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	Options       []string `json:"options,omitempty"` // For MCQ questions
	QuestionType  string   `json:"question_type"`     // "mcq", "write_word", "translate"
	Direction     string   `json:"direction"`         // "source_to_target", "target_to_source"
}

// AutoGenerateRequest represents configuration for auto-generating questions
type AutoGenerateRequest struct {
	DeckSelections []DeckSelectionRequest `json:"deck_selections"` // Categories to generate from
	QuestionTypes  []string               `json:"question_types"`  // "mcq", "write_word", "translate"
	Directions     []string               `json:"directions"`      // "source_to_target", "target_to_source"
	QuestionCount  int                    `json:"question_count"`
	Difficulty     string                 `json:"difficulty,omitempty"`
}

// QuizDetailResponse represents a detailed quiz response
type QuizDetailResponse struct {
	ID           int64                  `json:"id,string"`
	SessionID    string                 `json:"session_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description,omitempty"`
	Config       quizpkg.QuizConfig     `json:"config"`
	Questions    []QuizQuestionResponse `json:"questions,omitempty"`
	CurrentIndex int                    `json:"current_index"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Progress     QuizProgress           `json:"progress"`
}

// QuizQuestionResponse represents a question in API responses
type QuizQuestionResponse struct {
	ID           int64    `json:"id,string"`
	QuestionText string   `json:"question_text"`
	Options      []string `json:"options,omitempty"`
	QuestionType string   `json:"question_type"`
	Direction    string   `json:"direction"`
	UserAnswer   *string  `json:"user_answer,omitempty"`
	IsCorrect    *bool    `json:"is_correct,omitempty"`
	TimeSpent    *int     `json:"time_spent,omitempty"` // in seconds
}

// QuizProgress represents quiz progress
type QuizProgress struct {
	Answered   int     `json:"answered"`
	Total      int     `json:"total"`
	Percentage float64 `json:"percentage"`
	Correct    int     `json:"correct,omitempty"`
	Score      float64 `json:"score,omitempty"`
}

// SubmitAnswerRequest represents a request to submit an answer
type SubmitAnswerRequest struct {
	QuestionID int64  `json:"question_id,string"`
	Answer     string `json:"answer"`
	TimeSpent  int    `json:"time_spent,omitempty"` // in seconds
}

// SubmitAnswerResponse represents the response after submitting an answer
type SubmitAnswerResponse struct {
	IsCorrect     bool         `json:"is_correct"`
	CorrectAnswer string       `json:"correct_answer,omitempty"`
	Explanation   string       `json:"explanation,omitempty"`
	Progress      QuizProgress `json:"progress"`
}

// CreateQuestionRequest represents a request to create a custom question
type CreateQuestionRequest struct {
	DeckID        int64    `json:"deck_id,string"`
	CategoryKey   string   `json:"category_key"`
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	QuestionType  string   `json:"question_type"`
	WrongAnswers  []string `json:"wrong_answers,omitempty"`
}
