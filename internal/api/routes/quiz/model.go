package quiz

// ListQuizzesRequest contains filtering parameters
type ListQuizzesRequest struct {
	Category string `json:"category"`
	Sort     string `json:"sort"` // "popular", "recent", "rating"
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

// QuizListItem represents a quiz in list view
type QuizListItem struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	CreatorName   string   `json:"creator_name,omitempty"`
	QuestionCount int      `json:"question_count"`
	AttemptCount  int      `json:"attempt_count"`
	AverageScore  *float64 `json:"average_score,omitempty"`
	IsPublic      bool     `json:"is_public"`
	CreatedAt     int64    `json:"created_at"`
}

// QuizDetailResponse contains full quiz details
type QuizDetailResponse struct {
	ID               int64    `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	CreatorID        *int64   `json:"creator_id,omitempty"`
	CreatorName      string   `json:"creator_name,omitempty"`
	QuestionCount    int      `json:"question_count"`
	TimeLimit        *int     `json:"time_limit,omitempty"`
	PassPercentage   *int     `json:"pass_percentage,omitempty"`
	ShuffleQuestions bool     `json:"shuffle_questions"`
	IsPublic         bool     `json:"is_public"`
	IsSystem         bool     `json:"is_system"`
	GivesCoins       bool     `json:"gives_coins"`
	CoinReward       int      `json:"coin_reward"`
	AttemptCount     int      `json:"attempt_count"`
	AverageScore     *float64 `json:"average_score,omitempty"`
	CreatedAt        int64    `json:"created_at"`
}

// CreateQuizRequest represents request to create a quiz
type CreateQuizRequest struct {
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	DeckID             int64               `json:"deck_id"`
	CategorySelections []CategorySelection `json:"category_selections"`
	QuestionMode       string              `json:"question_mode"`
	TimeLimit          *int                `json:"time_limit,omitempty"`
	PassPercentage     *int                `json:"pass_percentage,omitempty"`
	ShuffleQuestions   bool                `json:"shuffle_questions"`
	IsPublic           bool                `json:"is_public"`
}

// CategorySelection represents category and question count
type CategorySelection struct {
	CategoryID    int64 `json:"category_id"`
	QuestionCount int   `json:"question_count"`
}

// GenerateQuizRequest represents request to generate a quiz using the universal system
type GenerateQuizRequest struct {
	DeckID         int64                `json:"deck_id"`
	QuestionTypes  []string             `json:"question_types"`
	Directions     []string             `json:"directions"`
	Categories     []string             `json:"categories,omitempty"`
	QuestionCount  int                  `json:"question_count"`
	Difficulty     string               `json:"difficulty,omitempty"`
	CustomQuestions []CustomQuestionRequest `json:"custom_questions,omitempty"`
}

// CustomQuestionRequest represents a user-defined question
type CustomQuestionRequest struct {
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	WrongAnswers  []string `json:"wrong_answers,omitempty"`
	QuestionType  string   `json:"question_type"`
}

// GeneratedQuestionResponse represents a generated question
type GeneratedQuestionResponse struct {
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	Options       []string `json:"options,omitempty"`
	QuestionType  string   `json:"question_type"`
	Direction     string   `json:"direction"`
}

// GenerateQuizResponse represents the response for quiz generation
type GenerateQuizResponse struct {
	Questions []*GeneratedQuestionResponse `json:"questions"`
}

// CreateQuestionRequest represents request to create a question
type CreateQuestionRequest struct {
	DeckID        int64  `json:"deck_id"`
	CategoryID    int64  `json:"category_id"`
	QuestionKey   string `json:"question_key"`
	QuestionText  string `json:"question_text"`
	CorrectAnswer string `json:"correct_answer"`
	QuestionType  string `json:"question_type"` // "multiple_choice", "written", "true_false"

	// For MCQ
	AnswerMode    string   `json:"answer_mode"` // "manual" or "auto_generated"
	ManualAnswers []string `json:"manual_answers,omitempty"`

	// For auto-generated answers
	GenerationRule   string `json:"generation_rule"`    // "same_category", "same_deck", etc.
	DataSourceName   string `json:"data_source_name"`   // "quran_words"
	WrongAnswerCount int    `json:"wrong_answer_count"` // default: 3
}

// DeckListItem represents a deck in list view
type DeckListItem struct {
	ID            int64  `json:"id,string"`
	DeckKey       string `json:"deck_key"`
	Title         string `json:"title"`
	CategoryCount int    `json:"category_count"`
	QuestionCount int    `json:"question_count"`
}

// CategoryListItem represents a category
type CategoryListItem struct {
	ID            int64  `json:"id,string"`
	CategoryKey   string `json:"category_key"`
	Title         string `json:"title"`
	QuestionCount int    `json:"question_count"`
}
