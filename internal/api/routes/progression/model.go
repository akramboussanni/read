package progression

// ProgressionStatusResponse represents the user's progression status
type ProgressionStatusResponse struct {
	CurrentLevel          int     `json:"current_level"`
	TotalQuizzesCompleted int     `json:"total_quizzes_completed"`
	TotalCoinsEarned      int     `json:"total_coins_earned"`
	StreakDays            int     `json:"streak_days"`
	CoinBalance           int     `json:"coin_balance"`
	NextQuiz              *QuizPreview `json:"next_quiz,omitempty"`
}

// QuizPreview represents a quiz in list view
type QuizPreview struct {
	ID             int64   `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Level          int     `json:"level"`
	IsLocked       bool    `json:"is_locked"`
	IsCompleted    bool    `json:"is_completed"`
	BestScore      *float64 `json:"best_score,omitempty"`
	BestPercentage *float64 `json:"best_percentage,omitempty"`
	CoinReward     int     `json:"coin_reward"`
	QuestionCount  int     `json:"question_count"`
}

// StartQuizRequest is empty for now, but can be extended
type StartQuizRequest struct{}

// StartQuizResponse contains the quiz attempt and questions
type StartQuizResponse struct {
	AttemptID int64      `json:"attempt_id"`
	QuizID    int64      `json:"quiz_id"`
	Title     string     `json:"title"`
	TimeLimit *int       `json:"time_limit,omitempty"`
	Questions []QuestionWithOptions `json:"questions"`
}

// QuestionWithOptions represents a question with its answer options
type QuestionWithOptions struct {
	ID           int64    `json:"id"`
	QuestionText string   `json:"question_text"`
	QuestionType string   `json:"question_type"`
	Points       int      `json:"points"`
	Options      []Option `json:"options,omitempty"` // Only for MCQ
}

// Option represents an answer option (doesn't reveal correct answer)
type Option struct {
	ID         int64  `json:"id"`
	OptionText string `json:"option_text"`
}

// SubmitQuizRequest contains user's answers
type SubmitQuizRequest struct {
	AttemptID int64          `json:"attempt_id"`
	Answers   []SubmitAnswer `json:"answers"`
}

// SubmitAnswer represents a single answer
type SubmitAnswer struct {
	QuestionID int64  `json:"question_id"`
	Answer     string `json:"answer"` // Option ID for MCQ, text for written
}

// SubmitQuizResponse contains the results
type SubmitQuizResponse struct {
	Score          float64 `json:"score"`
	MaxScore       float64 `json:"max_score"`
	Percentage     float64 `json:"percentage"`
	Passed         bool    `json:"passed"`
	CoinsEarned    int     `json:"coins_earned"`
	NextUnlocked   *int64  `json:"next_unlocked,omitempty"`
	Results        []AnswerResult `json:"results"`
}

// AnswerResult shows correctness of each answer
type AnswerResult struct {
	QuestionID    int64   `json:"question_id"`
	UserAnswer    string  `json:"user_answer"`
	CorrectAnswer string  `json:"correct_answer"`
	IsCorrect     bool    `json:"is_correct"`
	PointsEarned  float64 `json:"points_earned"`
}
