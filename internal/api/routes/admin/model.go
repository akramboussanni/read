package admin

import (
	"github.com/akramboussanni/gocode/internal/api/routes/quiz"
	"github.com/akramboussanni/gocode/internal/model"
)

// UserListResponse represents a user in the admin list
type UserListResponse struct {
	ID             int64  `json:"id,string"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	IsAdmin        bool   `json:"is_admin"`
	EmailConfirmed bool   `json:"email_confirmed"`
	CreatedAt      int64  `json:"created_at,string"`
}

// UserDetailResponse includes progression and stats
type UserDetailResponse struct {
	User                  UserListResponse `json:"user"`
	CurrentLevel          int              `json:"current_level"`
	TotalQuizzesCompleted int              `json:"total_quizzes_completed"`
	TotalCoinsEarned      int              `json:"total_coins_earned"`
	CoinBalance           int              `json:"coin_balance"`
	StreakDays            int              `json:"streak_days"`
	LastActivityDate      *string          `json:"last_activity_date,omitempty"`
}

// ChangePasswordRequest for admin to change user password
type ChangePasswordRequest struct {
	UserID      int64  `json:"user_id,string"`
	NewPassword string `json:"new_password"`
}

// CreateQuizRequest for admin quiz creation (same as regular quiz creation)
type CreateQuizRequest struct {
	Title           string                       `json:"title"`
	Description     string                       `json:"description,omitempty"`
	ManualQuestions []quiz.ManualQuestionRequest `json:"manual_questions,omitempty"` // User-created questions
	AutoGenerate    *quiz.AutoGenerateRequest    `json:"auto_generate,omitempty"`    // Auto-generation config
	IsPublic        bool                         `json:"is_public"`
	// Admin-only fields
	PassPercentage     *int   `json:"pass_percentage,omitempty"`             // Admin only
	GivesCoins         bool   `json:"gives_coins"`                           // Admin only (default false)
	CoinReward         int    `json:"coin_reward"`                           // Admin only (default 0)
	LevelOrder         int    `json:"level_order"`                           // Admin only (default 0)
	PrerequisiteQuizID *int64 `json:"prerequisite_quiz_id,string,omitempty"` // Admin only
	IsSystem           bool   `json:"is_system"`                             // Admin only (default false)
}

// CategorySelection for quiz creation
type CategorySelection struct {
	CategoryID    int64 `json:"category_id,string"`
	QuestionCount int   `json:"question_count"`
}

// QuizStatsResponse shows quiz statistics
type QuizStatsResponse struct {
	QuizID          int64   `json:"quiz_id,string"`
	Title           string  `json:"title"`
	CreatedBy       *int64  `json:"created_by,string,omitempty"`
	CreatorUsername *string `json:"creator_username,omitempty"`
	IsSystem        bool    `json:"is_system"`
	TotalAttempts   int     `json:"total_attempts"`
	AverageScore    float64 `json:"average_score"`
	PassRate        float64 `json:"pass_rate"`
	CreatedAt       int64   `json:"created_at,string"`
}

// UserQuizResponse for user-generated quizzes list
type UserQuizResponse struct {
	ID              int64  `json:"id,string"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	CreatedBy       int64  `json:"created_by,string"`
	CreatorUsername string `json:"creator_username"`
	IsPublic        bool   `json:"is_public"`
	TotalAttempts   int    `json:"total_attempts"`
	CreatedAt       int64  `json:"created_at,string"`
}

// ToUserListResponse converts model.User to response
func ToUserListResponse(user *model.User) UserListResponse {
	return UserListResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Role:           user.Role,
		IsAdmin:        user.IsAdmin,
		EmailConfirmed: user.EmailConfirmed,
		CreatedAt:      user.CreatedAt,
	}
}

// UpdateQuizRequest for updating quiz properties
type UpdateQuizRequest struct {
	Title              *string `json:"title,omitempty"`
	Description        *string `json:"description,omitempty"`
	PassPercentage     *int    `json:"pass_percentage,omitempty"`
	ShuffleQuestions   *bool   `json:"shuffle_questions,omitempty"`
	GivesCoins         *bool   `json:"gives_coins,omitempty"`
	CoinReward         *int    `json:"coin_reward,omitempty"`
	LevelOrder         *int    `json:"level_order,omitempty"`
	PrerequisiteQuizID *int64  `json:"prerequisite_quiz_id,string,omitempty"`
	IsPublic           *bool   `json:"is_public,omitempty"`
}

// SystemStatsResponse for system statistics
type SystemStatsResponse struct {
	TotalUsers     int     `json:"total_users"`
	TotalQuizzes   int     `json:"total_quizzes"`
	TotalAttempts  int     `json:"total_attempts"`
	ActiveUsers7d  int     `json:"active_users_7d"`
	AverageScore   float64 `json:"average_score"`
	CompletionRate float64 `json:"completion_rate"`
	TotalQuestions int     `json:"total_questions"`
}
