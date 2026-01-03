package admin

import "github.com/akramboussanni/gocode/internal/model"

// UserListResponse represents a user in the admin list
type UserListResponse struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	IsAdmin        bool   `json:"is_admin"`
	EmailConfirmed bool   `json:"email_confirmed"`
	CreatedAt      int64  `json:"created_at"`
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
	UserID      int64  `json:"user_id"`
	NewPassword string `json:"new_password"`
}

// CreateQuizRequest for admin quiz creation
type CreateQuizRequest struct {
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	DeckID             int64   `json:"deck_id"`
	CategorySelections []CategorySelection `json:"category_selections"`
	TimeLimit          *int    `json:"time_limit,omitempty"`
	PassPercentage     *int    `json:"pass_percentage,omitempty"`
	ShuffleQuestions   bool    `json:"shuffle_questions"`
	QuestionMode       string  `json:"question_mode"` // 'ar_to_fr', 'fr_to_ar'
	GivesCoins         bool    `json:"gives_coins"`
	CoinReward         int     `json:"coin_reward"`
	LevelOrder         int     `json:"level_order"`
	PrerequisiteQuizID *int64  `json:"prerequisite_quiz_id,omitempty"`
	IsPublic           bool    `json:"is_public"`
	IsSystem           bool    `json:"is_system"`
}

// CategorySelection for quiz creation
type CategorySelection struct {
	CategoryID    int64 `json:"category_id"`
	QuestionCount int   `json:"question_count"`
}

// QuizStatsResponse shows quiz statistics
type QuizStatsResponse struct {
	QuizID          int64   `json:"quiz_id"`
	Title           string  `json:"title"`
	CreatedBy       *int64  `json:"created_by,omitempty"`
	CreatorUsername *string `json:"creator_username,omitempty"`
	IsSystem        bool    `json:"is_system"`
	TotalAttempts   int     `json:"total_attempts"`
	AverageScore    float64 `json:"average_score"`
	PassRate        float64 `json:"pass_rate"`
	CreatedAt       int64   `json:"created_at"`
}

// UserQuizResponse for user-generated quizzes list
type UserQuizResponse struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	CreatedBy        int64  `json:"created_by"`
	CreatorUsername  string `json:"creator_username"`
	IsPublic         bool   `json:"is_public"`
	TotalAttempts    int    `json:"total_attempts"`
	CreatedAt        int64  `json:"created_at"`
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
